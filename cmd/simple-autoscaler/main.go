package main

import (
	"context"
	"flag"
	"time"

	"github.com/tensorbytes/simplescale/simpleautoscaler"
	kubeflag "k8s.io/component-base/cli/flag"
	klogv2 "k8s.io/klog/v2"
)

var (
	runOnceInterval = flag.Duration("runonce-interval", 10*time.Second, `How often the recommender run once`)
)

func main() {
	klogv2.InitFlags(nil)
	kubeflag.InitFlags()

	autoScaler, err := simpleautoscaler.NewSimpleAutoScalerController()
	if err != nil {
		klogv2.Fatalf("initialize error for simpleautoscaler controller: %v", err)
	}

	klogv2.Info("Simple Autosacler Controller are Ready")

	ticker := time.Tick(*runOnceInterval)
	for range ticker {
		ctx, cancel := context.WithTimeout(context.Background(), *runOnceInterval)
		defer cancel()
		autoScaler.RunOnce(ctx)
	}
}
