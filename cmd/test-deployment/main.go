package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	var (
		addr         = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
		changePeriod = flag.Duration("change-period", 5*time.Second, "The duration of the rate change period.")
	)

	flag.Parse()

	var cpuUtilizationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_utilization_rate",
			Help: "CPU Utilization rate.",
		},
		[]string{"service"},
	)

	prometheus.MustRegister(cpuUtilizationRate)

	go func() {
		rand.Seed(time.Now().Unix())
		value := rand.Float64() * 100
		addOrsub := 1
		for {
			if addOrsub == -1 {
				value = value - rand.Float64()*10
			}
			if addOrsub == 1 {
				value = value + rand.Float64()*10
			}
			if addOrsub == 0 {
				value = rand.Float64() * 100
			}
			if value > 90 {
				addOrsub = -1
			}
			if value < 10 {
				addOrsub = 1
			}
			if (value > 47) && (value < 52) {
				addOrsub = 0
			}
			log.Println(value)
			cpuUtilizationRate.WithLabelValues("random-metrics").Set(value)
			time.Sleep(*changePeriod)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
