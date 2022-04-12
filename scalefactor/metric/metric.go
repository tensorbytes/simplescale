package metric

import (
	"errors"
	"fmt"
	"time"
)

const (
	INT_MAX = int(^uint(0) >> 1)
)

func NewHistoryMetric() HistoryMetric {
	return HistoryMetric{
		Timestamps: make([]time.Time, 0),
		Values:     make([]float64, 0),
		length:     0,
		capacity:   INT_MAX,
	}
}

type TimestampMetric struct {
	Timestamp time.Time
	Value     float64
}

type HistoryMetric struct {
	Timestamps []time.Time // time from early to late
	Values     []float64
	length     int
	capacity   int
}

func (m *HistoryMetric) Len() int {
	return m.length
}

func (m *HistoryMetric) Empty() bool {
	return m.length <= 0
}

func (m *HistoryMetric) SetMaxSize(size int) bool {
	if size >= m.length {
		m.capacity = size
		return true
	}
	return false
}

func (m *HistoryMetric) Capacity() int {
	return m.capacity
}

func (m *HistoryMetric) String() (s string) {
	for i := 0; i < m.Len(); i++ {
		s = s + fmt.Sprintf("%d, %s, %f \n", i, m.Timestamps[i].Format(time.RFC3339), m.Values[i])
	}
	return s
}

func (m *HistoryMetric) AppendValue(metric TimestampMetric) (int, bool) {
	if m.length >= m.capacity {
		return m.length, false
	}
	if m.Empty() {
		m.append(metric)
		return m.length, true
	}
	lastTime := m.Timestamps[m.length-1]
	if lastTime.Before(metric.Timestamp) {
		m.append(metric)
		return m.length, true
	}
	return m.length, false
}

func (m *HistoryMetric) append(metric TimestampMetric) {
	m.Timestamps = append(m.Timestamps, metric.Timestamp)
	m.Values = append(m.Values, metric.Value)
	m.length = m.length + 1
}

// get sub history metric
func (m *HistoryMetric) SubHistoryMetricInRangeTime(start, end time.Time) (subMetric HistoryMetric) {
	if start.After(end) || (m.length < 1) {
		return
	}
	for i, t := range m.Timestamps {
		if start.Before(t) && end.After(t) {
			subMetric.append(
				TimestampMetric{
					Timestamp: m.Timestamps[i],
					Value:     m.Values[i],
				})
		}
	}
	return
}

// Get max value from specific time
func (m *HistoryMetric) GetMaxValueInRangeTime(start, end time.Time) (float64, error) {
	if start.After(end) {
		return 0, errors.New("start time is later than end time")
	}
	if m.Empty() {
		return 0, errors.New("HistoryMetric is emtry")
	}
	maxValue := m.Values[0]
	for i, t := range m.Timestamps {
		if start.Before(t) && end.After(t) {
			if m.Values[i] > maxValue {
				maxValue = m.Values[i]
			}
		}
	}
	return maxValue, nil
}
