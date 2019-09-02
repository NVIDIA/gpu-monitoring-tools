#!/bin/bash
docker build -f Dockerfile . -t runai/pod-gpu-metrics-exporter
docker push runai/pod-gpu-metrics-exporter
kubectl delete pod -l app=pod-gpu-metrics-exporter -n runai
