# Copyright (c) 2020, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DOCKER   ?= docker
MKDIR    ?= mkdir
REGISTRY ?= nvidia

DCGM_VERSION   := 2.0.13
GOLANG_VERSION := 1.14.2
VERSION        := 2.1.1
FULL_VERSION   := $(DCGM_VERSION)-$(VERSION)

.PHONY: all binary install check-format
all: ubuntu18.04 ubuntu20.04 ubi8

binary:
	go build -o dcgm-exporter github.com/NVIDIA/gpu-monitoring-tools/pkg

install: binary
	install -m 557 dcgm-exporter /usr/bin/dcgm-exporter
	install -m 557 -D ./etc/dcgm-exporter/default-counters.csv /etc/dcgm-exporter/default-counters.csv
	install -m 557 -D ./etc/dcgm-exporter/dcp-metrics-included.csv /etc/dcgm-exporter/dcp-metrics-included.csv

check-format:
	#test $$(gofmt -l pkg bindings | tee /dev/stderr | wc -l) -eq 0

push:
	#$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu20.04"
	#$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu18.04"
	$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubi8"

push-short:
	$(DOCKER) tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu18.04" "gcr.io/run-ai-lab/dcgm-exporter:$(DCGM_VERSION)"
	$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:$(DCGM_VERSION)"

push-ci:
	$(DOCKER) tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu18.04" "gcr.io/run-ai-lab/dcgm-exporter:$(VERSION)"
	$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:$(VERSION)"

push-latest:
	$(DOCKER) tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu18.04" "gcr.io/run-ai-lab/dcgm-exporter:latest"
	$(DOCKER) push "gcr.io/run-ai-lab/dcgm-exporter:latest"

ubuntu20.04:
	$(DOCKER) build --pull \
		--build-arg "GOLANG_VERSION=$(GOLANG_VERSION)" \
		--build-arg "DCGM_VERSION=$(DCGM_VERSION)" \
		--tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu20.04" \
		--file docker/Dockerfile.ubuntu20.04 .

ubuntu18.04:
	$(DOCKER) build --pull \
		--build-arg "GOLANG_VERSION=$(GOLANG_VERSION)" \
		--build-arg "DCGM_VERSION=$(DCGM_VERSION)" \
		--tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubuntu18.04" \
		--file docker/Dockerfile.ubuntu18.04 .

ubi8:
	$(DOCKER) build --pull \
		--build-arg "GOLANG_VERSION=$(GOLANG_VERSION)" \
		--build-arg "DCGM_VERSION=$(DCGM_VERSION)" \
		--build-arg "VERSION=$(FULL_VERSION)" \
		--tag "gcr.io/run-ai-lab/dcgm-exporter:$(FULL_VERSION)-ubi8" \
		--file docker/Dockerfile.ubi8 .
