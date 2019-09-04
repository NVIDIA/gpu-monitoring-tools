package main

import (
	"flag"
	"syscall"
	"os"
	"github.com/golang/glog"
)

// res: curl localhost:9400/metrics
func main() {
	defer glog.Flush()
	flag.Parse()

	glog.Info("Starting service...")

	// http port serving metrics
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9400"
	}
	port = ":" + port 

	glog.Info("Starting OS watcher.")
	sigs := sigWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// watch and write gpu metrics to dcgm-pod.prom
	go func() {
		glog.Info("Starting FS watcher.")
		watchAndWriteGPUmetrics()
	}()

	server := newHttpServer(port)
	defer stopHttp(server)

	// expose metrics to localhost:9400/metrics
	go func() {
		glog.V(1).Infof("Running http server on localhost%s", port)
		startHttp(server)
	}()

	sig := <-sigs
	glog.V(2).Infof("Received signal \"%v\", shutting down.", sig)
	return
}
