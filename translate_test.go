package cumulativetodelta

import (
	"reflect"
	"testing"
)

func TestMetric_Identity(t *testing.T) {
	type fields struct {
		Name  string
		Value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   MetricIdentity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metric{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			if got := m.Identity(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Metric.Identity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricAggregator_Record(t *testing.T) {
	type fields struct {
		States map[MetricIdentity]State
	}
	type args struct {
		in Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Recording a metric stores in states",
			fields: fields{
				States: make(map[MetricIdentity]State),
			},
			args:   args{
				in: Metric{
					Name:  "foo",
					Value: 80.0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetricAggregator{
				States: tt.fields.States,
			}
			m.Record(tt.args.in)
		})
	}
}

func TestMetricAggregator_Flush(t *testing.T) {
	type fields struct {
		States map[MetricIdentity]State
	}
	tests := []struct {
		name   string
		fields fields
		want   []Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetricAggregator{
				States: tt.fields.States,
			}
			if got := m.Flush(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricAggregator.Flush() = %v, want %v", got, tt.want)
			}
		})
	}
}
