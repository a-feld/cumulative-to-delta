package cumulativetodelta

import (
	"reflect"
	"sync"
	"testing"
)

var (
	foobar   Metric         = Metric{Name: "foobar"}
	foobarId MetricIdentity = ComputeMetricIdentity(foobar)
)

func TestMetricTracker_Record(t *testing.T) {
	type fields struct {
		States  map[MetricIdentity]*State
		Metrics map[MetricIdentity]Metric
	}
	type args struct {
		in Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[MetricIdentity]*State
	}{
		{
			name: "Initial Value Recorded",
			fields: fields{
				States: make(map[MetricIdentity]*State),
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 100,
				},
			},
			want: map[MetricIdentity]*State{
				foobarId: {
					RunningTotal: 100,
					LatestValue:  100,
				},
			},
		},
		{
			name: "Higher Value Recorded",
			fields: fields{
				States: map[MetricIdentity]*State{
					foobarId: {
						RunningTotal: 100,
						LatestValue:  100,
					},
				},
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 200,
				},
			},
			want: map[MetricIdentity]*State{
				foobarId: {
					RunningTotal: 200,
					LatestValue:  200,
				},
			},
		},
		{
			name: "Lower Value Recorded - No Offset",
			fields: fields{
				States: map[MetricIdentity]*State{
					foobarId: {
						RunningTotal: 200,
						LatestValue:  200,
					},
				},
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 80,
				},
			},
			want: map[MetricIdentity]*State{
				foobarId: {
					RunningTotal: 280,
					LatestValue:  80,
					Offset:       200,
				},
			},
		},
		{
			name: "Lower Value Recorded - With Existing Offset",
			fields: fields{
				States: map[MetricIdentity]*State{
					foobarId: {
						RunningTotal: 280,
						LatestValue:  80,
						Offset:       200,
					},
				},
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 20,
				},
			},
			want: map[MetricIdentity]*State{
				foobarId: {
					RunningTotal: 300,
					LatestValue:  20,
					Offset:       280,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricTracker{
				mu:     sync.Mutex{},
				States: tt.fields.States,
			}
			m.Record(tt.args.in)
			if got := m.States; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricTracker.States = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricTracker_Flush(t *testing.T) {
	type fields struct {
		States map[MetricIdentity]*State
	}
	tests := []struct {
		name   string
		fields fields
		want   []Metric
	}{
		{
			name: "empty",
			want: []Metric{},
		},
		{
			name: "single",
			fields: fields{
				States: map[MetricIdentity]*State{
					foobarId: {
						RunningTotal: 62,
						LatestValue:  20,
						Offset:       42,
						LastFlushed:  0,
					},
				},
			},
			want: []Metric{{
				Name:  "foobar",
				Value: 62,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetricTracker{
				mu:     sync.Mutex{},
				States: tt.fields.States,
			}
			if got := m.Flush(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricTracker.Flush() = %v, want %v", got, tt.want)
			}

			for _, metric := range tt.want {
				id := ComputeMetricIdentity(metric)
				if !reflect.DeepEqual(0.0, m.States[id].RunningTotal-m.States[id].LastFlushed) {
					t.Errorf("Post-Flush MetricTracker.States[%v] total - last flushed = %v, want 0", id, m.States[id].RunningTotal-m.States[id].LastFlushed)
				}
			}
		})
	}
}
