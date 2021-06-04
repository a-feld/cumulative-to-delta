package tracking

import (
	"reflect"
	"sync"
	"testing"
)

func Test_metricIdentity_Metric(t *testing.T) {
	type fields struct {
		hash    uint64
		metrics map[metricIdentity]*Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   *Metric
	}{
		{
			name: "Metric Present",
			fields: fields{
				hash: 0,
				metrics: map[metricIdentity]*Metric{
					{}: &foobar,
				},
			},
			want: &foobar,
		},
		{
			name: "Metric Missing",
			fields: fields{
				hash: 0,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testMetrics := sync.Map{}
			for i, m := range tt.fields.metrics {
				testMetrics.Store(i, m)
			}
			original := metrics
			metrics = &testMetrics
			t.Cleanup(func() { metrics = original })
			i := metricIdentity{
				hash: tt.fields.hash,
			}
			if got := i.Metric(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricIdentity.Metric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeMetricIdentity(t *testing.T) {
	type args struct {
		m Metric
	}
	tests := []struct {
		name string
		args args
		want metricIdentity
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComputeMetricIdentity(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComputeMetricIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}
