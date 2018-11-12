## Changelog for prometheus-operator

Following changes are made to prometheus-operator helm charts [release-0.18](https://github.com/coreos/prometheus-operator/tree/release-0.18/helm) to enable NVIDIA GPU metrics:

* exporter-node - The node-exporter daemonset is modified to collect GPU metrics from NVIDIA dcgm-exporter. Changed files:
	* exporter-node/values.yaml
	* templates/daemonset.yaml
	* templates/service.yaml
	* templates/NOTES.txt

* grafana - Some GPU metrics graphs are added to grafana nodes dashboard. Changed Files:
	* grafana/dashboards/nodes-dashboard.json
