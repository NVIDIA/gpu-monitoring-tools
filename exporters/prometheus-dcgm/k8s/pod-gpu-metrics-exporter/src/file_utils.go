// Copyright (c) 2018, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

func writeDestFile(tmpF string, destFile string) error {
	err := os.Rename(tmpF, destFile)
	if err != nil {
		return fmt.Errorf("error replacing temp file with %s: %v", destFile, err)
	}

	// Set read permissions
	mode := os.FileMode(0644)
	err = os.Chmod(destFile, mode)
	if err != nil {
		return fmt.Errorf("error setting %s file permissions: %v", destFile, err)
	}
	return nil
}

func createMetricsDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("err creating directory %s: %v", dir, err)
	}
	return nil
}

func processHealthMetric(line string, tmpF *os.File) {
	lineSlice := strings.Split(line, " ")
	xidCh := lineSlice[len(lineSlice)-1]
	xidCh = strings.Trim(xidCh, "\n")
	xid, err := strconv.Atoi(xidCh)
	if err != nil {
		glog.Errorf("err parse xid %s", err)
		return
	}
	health := 0
	if xid != 0 && xid != 31 && xid != 43 && xid != 45 {
		health = 1
		glog.Infof("xid: %d", xid)
	}
	content := strings.Split(strings.Split(line, "}")[0], "{")[1]
	metric := fmt.Sprintf("dcgm_gpu_health{%s} %d\n", content, health)
	_, err = tmpF.WriteString(metric)
	if err != nil {
		glog.Errorf("error writing health metric: %s", err)
	}
}
