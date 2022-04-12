package scalefactor

import (
	"time"

	"github.com/tensorbytes/simplescale/scalefactor/metric"
)

type ScaleAlgorithm interface {
	PredictNextStage(metricValue metric.HistoryMetric, stageTime time.Duration) ModelResult
}

func NewSimpleScaleModel() *SimpleScaleModel {
	return &SimpleScaleModel{
		timePeriod: 24 * time.Hour,
	}

}

// recommend model and algorithm
type SimpleScaleModel struct {
	timePeriod time.Duration
}

// the function predict next stage value
// use Periodic laws
func (m *SimpleScaleModel) PredictNextStage(metricValue metric.HistoryMetric, stageTime time.Duration) (result ModelResult) {
	// The maximum of this time yesterday, range 1 hour time
	startTime := time.Now().UTC().Add(-m.timePeriod)
	endTime := startTime.Add(stageTime)
	maxValue, err := metricValue.GetMaxValueInRangeTime(startTime, endTime)
	result.Value = maxValue
	result.ErrorMessage = err
	return
}

type ModelResult struct {
	Value        float64
	ErrorMessage error
}
