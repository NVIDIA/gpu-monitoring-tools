#!/usr/bin/env bash

docker build -t bashimao/pod-gpu-metrics-exporter:v1 .

docker push bashimao/pod-gpu-metrics-exporter:v1
