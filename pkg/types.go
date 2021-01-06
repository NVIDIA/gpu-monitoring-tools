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

package main

import (
	"fmt"
	"net/http"
	"sync"
	"text/template"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/dcgm"
)

var (
	SkipDCGMValue   = "SKIPPING DCGM VALUE"
	FailedToConvert = "ERROR - FAILED TO CONVERT TO STRING"

	nvidiaResourceName = "nvidia.com/gpu"

	// Note standard resource attributes
	podAttribute       = "pod"
	namespaceAttribute = "namespace"
	containerAttribute = "container"
)

type KubernetesGPUIDType string

const (
	GPUUID     KubernetesGPUIDType = "uid"
	DeviceName KubernetesGPUIDType = "device-name"
)

type Config struct {
	CollectorsFile      string
	Address             string
	CollectInterval     int
	Kubernetes          bool
	KubernetesGPUIdType KubernetesGPUIDType
	CollectDCP          bool
}

type Transform interface {
	Process(metrics [][]Metric) error
	Name() string
}

type MetricsPipeline struct {
	config *Config

	transformations []Transform
	metricsFormat   *template.Template
	countersText    string

	gpuCollector *DCGMCollector
}

type DCGMCollector struct {
	Counters     []Counter
	DeviceFields []dcgm.Short
	Cleanups     []func()

	GpuToLastNotIdleTime map[string]int64
}

type Counter struct {
	FieldID   dcgm.Short
	FieldName string
	PromType  string
	Help      string
}

type Metric struct {
	Name  string
	Value string

	GPU       string
	GPUUUID   string
	GPUDevice string

	Attributes map[string]string
}

func (m Metric) getIDOfType(idType KubernetesGPUIDType) (string, error) {
	switch idType {
	case GPUUID:
		return m.GPUUUID, nil
	case DeviceName:
		return m.GPUDevice, nil
	}
	return "", fmt.Errorf("unsupported KubernetesGPUIDType for MetricID %s", idType)
}

var promMetricType = map[string]bool{
	"gauge":     true,
	"counter":   true,
	"histogram": true,
	"summary":   true,
}

type MetricsServer struct {
	sync.Mutex

	server      http.Server
	metrics     string
	metricsChan chan string
}

type PodMapper struct {
	Config *Config
}

type PodInfo struct {
	Name      string
	Namespace string
	Container string
}
