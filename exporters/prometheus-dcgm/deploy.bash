docker build -t runai/dcgm-exporter:1.7.2 -f dcgm-exporter/Dockerfile .
docker push runai/dcgm-exporter:1.7.2
kubectl delete pod -n runai -l app=pod-gpu-metrics-exporter
