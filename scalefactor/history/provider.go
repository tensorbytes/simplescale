package history

import (
	"context"
	"errors"
	"fmt"
	"time"

	klogv2 "k8s.io/klog/v2"

	prometheusClient "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promethuesmodel "github.com/prometheus/common/model"
	"github.com/tensorbytes/simplescale/scalefactor/metric"
)

const (
	DEFAULT_MAX_QUERY_SIZE = 1 << 20
)

type PrometheusHistoryProviderConfig struct {
	Address      string
	QueryTimeout time.Duration
	QueryStep    time.Duration
}

func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) (provider PrometheusHistoryProvider, err error) {
	client, err := prometheusClient.NewClient(prometheusClient.Config{
		Address: config.Address,
	})
	if err != nil {
		return
	}
	queryer := prometheusv1.NewAPI(client)
	provider = PrometheusHistoryProvider{
		ClientAPI:    queryer,
		QueryStep:    config.QueryStep,
		QueryMaxSize: DEFAULT_MAX_QUERY_SIZE,
	}
	return
}

type HistoryProvider interface {
	Query(ctx context.Context, expression string) (metric.HistoryMetric, error)
	QueryRange(ctx context.Context, expression string, duration time.Duration) (metric.HistoryMetric, error)
}

type PrometheusHistoryProvider struct {
	ClientAPI    prometheusv1.API
	QueryStep    time.Duration
	QueryMaxSize int
}

// can
func (p *PrometheusHistoryProvider) Query(ctx context.Context, expression string) (metric.HistoryMetric, error) {
	var metricValue = metric.NewHistoryMetric()
	var err error
	queryValue, warnings, err := p.ClientAPI.Query(ctx, expression, time.Now())
	if len(warnings) > 0 {
		klogv2.Infof("promethues warnings: %v", warnings)
	}
	if err != nil {
		return metricValue, err
	}
	if !metricValue.SetMaxSize(p.QueryMaxSize) {
		err = errors.New("metricValue set max size failed")
		return metricValue, err
	}
	p.AddPromethuesModelToHistoryMetric(queryValue, &metricValue)
	return metricValue, err
}

func (p *PrometheusHistoryProvider) QueryRange(ctx context.Context, expression string, duration time.Duration) (metric.HistoryMetric, error) {
	var metricValue = metric.NewHistoryMetric()
	var err error
	queryRange := prometheusv1.Range{
		Start: time.Now().Add(-duration),
		End:   time.Now(),
		Step:  p.QueryStep,
	}
	queryValue, warnings, err := p.ClientAPI.QueryRange(ctx, expression, queryRange)
	if len(warnings) > 0 {
		klogv2.Infof("promethues warnings: %v", warnings)
	}
	if err != nil {
		err = fmt.Errorf("%w; query content: %s", err, expression)
		return metricValue, err
	}
	if !metricValue.SetMaxSize(p.QueryMaxSize) {
		err = errors.New("metricValue set max size failed")
		return metricValue, err
	}
	p.AddPromethuesModelToHistoryMetric(queryValue, &metricValue)
	return metricValue, err
}

func (p *PrometheusHistoryProvider) AddPromethuesModelToHistoryMetric(promeValue promethuesmodel.Value, metricHistoryMetric *metric.HistoryMetric) {
	switch promeValue.Type() {
	case promethuesmodel.ValScalar:
		val := promeValue.(*promethuesmodel.Scalar)
		metricHistoryMetric.AppendValue(metric.TimestampMetric{
			Timestamp: time.Unix(val.Timestamp.Unix(), 0),
			Value:     float64(val.Value),
		})
	case promethuesmodel.ValVector:
		val := promeValue.(promethuesmodel.Vector)
		for _, v := range val {
			metricHistoryMetric.AppendValue(metric.TimestampMetric{
				Timestamp: time.Unix(v.Timestamp.Unix(), 0),
				Value:     float64(v.Value),
			})
		}
	case promethuesmodel.ValMatrix:
		val := promeValue.(promethuesmodel.Matrix)
		for _, matrixPointer := range val {
			for _, v := range matrixPointer.Values {
				metricHistoryMetric.AppendValue(metric.TimestampMetric{
					Timestamp: time.Unix(v.Timestamp.Unix(), 0),
					Value:     float64(v.Value),
				})
			}
		}
	}
}
