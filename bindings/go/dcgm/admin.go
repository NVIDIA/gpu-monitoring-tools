/*
 * Copyright (c) 2020, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dcgm

/*
#cgo linux LDFLAGS: -ldl -Wl,--unresolved-symbols=ignore-in-object-files
#cgo darwin LDFLAGS: -ldl -Wl,-undefined,dynamic_lookup


#include <dlfcn.h>
#include "./dcgm_agent.h"
#include "./dcgm_structs.h"
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"
)

type mode int

// const for DCGM hostengine running modes: Embedded, Standalone or StartHostengine
const (
	Embedded mode = iota
	Standalone
	StartHostengine
)

type dcgmHandle struct{ handle C.dcgmHandle_t }

var (
	dcgmLibHandle        unsafe.Pointer
	stopMode             mode
	handle               dcgmHandle
	hostengineAsChildPid int
)

func initDcgm(m mode, args ...string) (err error) {
	const (
		dcgmLib = "libdcgm.so"
	)
	lib := C.CString(dcgmLib)
	defer freeCString(lib)

	dcgmLibHandle = C.dlopen(lib, C.RTLD_LAZY|C.RTLD_GLOBAL)
	if dcgmLibHandle == nil {
		return fmt.Errorf("%s not Found", dcgmLib)
	}

	// set the stopMode for shutdown()
	stopMode = m

	switch m {
	case Embedded:
		return startEmbedded()
	case Standalone:
		return connectStandalone(args...)
	case StartHostengine:
		return startHostengine()
	}

	return nil
}

func shutdown() (err error) {
	switch stopMode {
	case Embedded:
		err = stopEmbedded()
	case Standalone:
		err = disconnectStandalone()
	case StartHostengine:
		err = stopHostengine()
	}

	C.dlclose(dcgmLibHandle)
	return
}

func startEmbedded() (err error) {
	result := C.dcgmInit()
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error initializing DCGM: %s", err)
	}

	var cHandle C.dcgmHandle_t
	result = C.dcgmStartEmbedded(C.DCGM_OPERATION_MODE_AUTO, &cHandle)
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error starting nv-hostengine: %s", err)
	}
	handle = dcgmHandle{cHandle}
	return
}

func stopEmbedded() (err error) {
	result := C.dcgmStopEmbedded(handle.handle)
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error stopping nv-hostengine: %s", err)
	}

	result = C.dcgmShutdown()
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error shutting down DCGM: %s", err)
	}
	return
}

func connectStandalone(args ...string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("Missing dcgm address and / or port")
	}

	result := C.dcgmInit()
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error initializing DCGM: %s", err)
	}

	var cHandle C.dcgmHandle_t
	addr := C.CString(args[0])
	defer freeCString(addr)
	var connectParams C.dcgmConnectV2Params_t
	connectParams.version = makeVersion2(unsafe.Sizeof(connectParams))

	sck, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("Error parsing %s: %v\n", args[1], err)
	}
	connectParams.addressIsUnixSocket = C.uint(sck)

	result = C.dcgmConnect_v2(addr, &connectParams, &cHandle)
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error connecting to nv-hostengine: %s", err)
	}

	handle = dcgmHandle{cHandle}
	return
}

func disconnectStandalone() (err error) {
	result := C.dcgmDisconnect(handle.handle)
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error disconnecting from nv-hostengine: %s", err)
	}

	result = C.dcgmShutdown()
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error shutting down DCGM: %s", err)
	}
	return
}

func startHostengine() (err error) {
	bin, err := exec.LookPath("nv-hostengine")
	if err != nil {
		return fmt.Errorf("Error finding nv-hostengine: %s", err)
	}
	var procAttr syscall.ProcAttr
	procAttr.Files = []uintptr{
		uintptr(syscall.Stdin),
		uintptr(syscall.Stdout),
		uintptr(syscall.Stderr)}
	procAttr.Sys = &syscall.SysProcAttr{Setpgid: true}

	dir := "/tmp"
	tmpfile, err := ioutil.TempFile(dir, "dcgm")
	if err != nil {
		return fmt.Errorf("Error creating temporary file in %s directory: %s", dir, err)
	}
	socketPath := tmpfile.Name()
	defer os.Remove(socketPath)

	connectArg := "--domain-socket"
	hostengineAsChildPid, err = syscall.ForkExec(bin, []string{bin, connectArg, socketPath}, &procAttr)
	if err != nil {
		return fmt.Errorf("Error fork-execing nv-hostengine: %s", err)
	}

	result := C.dcgmInit()
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error initializing DCGM: %s", err)
	}

	var cHandle C.dcgmHandle_t
	var connectParams C.dcgmConnectV2Params_t
	connectParams.version = makeVersion2(unsafe.Sizeof(connectParams))
	isSocket := C.uint(1)
	connectParams.addressIsUnixSocket = isSocket
	cSockPath := C.CString(socketPath)
	defer freeCString(cSockPath)
	result = C.dcgmConnect_v2(cSockPath, &connectParams, &cHandle)
	if err = errorString(result); err != nil {
		return fmt.Errorf("Error connecting to nv-hostengine: %s", err)
	}

	handle = dcgmHandle{cHandle}
	return
}

func stopHostengine() (err error) {
	if err = disconnectStandalone(); err != nil {
		return
	}

	// terminate nv-hostengine
	cmd := exec.Command("nv-hostengine", "--term")
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Error terminating nv-hostengine: %s", err)
	}
	log.Println("Successfully terminated nv-hostengine.")

	return syscall.Kill(hostengineAsChildPid, syscall.SIGKILL)
}
