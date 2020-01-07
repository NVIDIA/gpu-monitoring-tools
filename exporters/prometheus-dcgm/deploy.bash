docker build -t runai/dcgm-exporter:1.7.3 -f dcgm-exporter/Dockerfile .
docker push runai/dcgm-exporter:1.7.3
kubectl delete pod -n runai -l app=pod-gpu-metrics-exporter
