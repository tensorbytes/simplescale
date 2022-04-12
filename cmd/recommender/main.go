package main

import (
	"context"
	"flag"
	"time"

	"github.com/tensorbytes/simplescale/scalefactor"
	"github.com/tensorbytes/simplescale/scalefactor/history"

	kubeflag "k8s.io/component-base/cli/flag"
	klogv2 "k8s.io/klog/v2"
)

var (
	runOnceInterval   = flag.Duration("runonce-interval", 10*time.Second, `How often the recommender run once`)
	promethuesAddress = flag.String("promethues-address", "http://127.0.0.1:9090", `The address of prometheus which collect metric`)
)

func main() {
	klogv2.InitFlags(nil)
	kubeflag.InitFlags()

	simpleRecommender, err := scalefactor.NewSimpleRecommender()
	if err != nil {
		klogv2.Fatalf("initialize error for simplerecommender: %v", err)
	}
	algorithm := scalefactor.NewSimpleScaleModel()
	simpleRecommender.RegisterAlgorithm(algorithm)

	collector, err := history.NewPrometheusHistoryProvider(history.PrometheusHistoryProviderConfig{
		Address:   *promethuesAddress,
		QueryStep: 5 * time.Minute,
	})
	if err != nil {
		klogv2.Fatalf("initialize promethues collector error: %v", err)
	}
	simpleRecommender.RegisterCollector(&collector)

	klogv2.Info("Simple Recommender are Ready")

	ticker := time.Tick(*runOnceInterval)
	for range ticker {
		ctx, cancel := context.WithTimeout(context.Background(), *runOnceInterval)
		defer cancel()
		simpleRecommender.RunOnce(ctx)
	}
}
