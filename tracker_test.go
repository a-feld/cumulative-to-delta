package cumulativetodelta

import (
	"reflect"
	"sync"
	"testing"
)

var (
	foobar   Metric         = Metric{Name: "foobar"}
	foobarId metricIdentity = ComputeMetricIdentity(foobar)
)

func TestMetricTracker_Record(t *testing.T) {
	type fields struct {
		States  sync.Map
		Metrics map[metricIdentity]Metric
	}
	type args struct {
		in Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[metricIdentity]*State
	}{
		{
			name: "Initial Value Recorded",
			fields: fields{
				States: sync.Map{},
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 100,
				},
			},
			want: map[metricIdentity]*State{
				foobarId: {
					RunningTotal: 100,
					LatestValue:  100,
				},
			},
		},
		{
			name: "Higher Value Recorded",
			fields: fields{
				States: func() sync.Map {
					m := sync.Map{}
					m.Store(foobarId, &State{
						RunningTotal: 100,
						LatestValue:  100,
					})
					return m
				}(),
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 200,
				},
			},
			want: map[metricIdentity]*State{
				foobarId: {
					RunningTotal: 200,
					LatestValue:  200,
				},
			},
		},
		{
			name: "Lower Value Recorded - No Offset",
			fields: fields{
				States: func() sync.Map {
					m := sync.Map{}
					m.Store(foobarId, &State{
						RunningTotal: 200,
						LatestValue:  200,
					})
					return m
				}(),
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 80,
				},
			},
			want: map[metricIdentity]*State{
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
				States: func() sync.Map {
					m := sync.Map{}
					m.Store(foobarId, &State{
						RunningTotal: 280,
						LatestValue:  80,
						Offset:       200,
					})
					return m
				}(),
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 20,
				},
			},
			want: map[metricIdentity]*State{
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
				States: tt.fields.States,
			}
			m.Record(tt.args.in)
			got := make(map[metricIdentity]*State)
			m.States.Range(func(key, value interface{}) bool {
				id := key.(metricIdentity)
				v := value.(*State)
				got[id] = v
				return true
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricTracker.States = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricTracker_Flush(t *testing.T) {
	type fields struct {
		States sync.Map
	}
	tests := []struct {
		name   string
		fields fields
		want   []Metric
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "single",
			fields: fields{
				States: func() sync.Map {
					m := sync.Map{}
					m.Store(foobarId, &State{
						RunningTotal: 62,
						LatestValue:  20,
						Offset:       42,
						LastFlushed:  0,
					})
					return m
				}(),
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
				States: tt.fields.States,
			}
			if got := m.Flush(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricTracker.Flush() = %v, want %v", got, tt.want)
			}

			for _, metric := range tt.want {
				id := ComputeMetricIdentity(metric)
				if s, _ := m.States.Load(id); !reflect.DeepEqual(0.0, s.(*State).RunningTotal-s.(*State).LastFlushed) {
					t.Errorf("Post-Flush MetricTracker.States[%v] total - last flushed = %v, want 0", id, s.(*State).RunningTotal-s.(*State).LastFlushed)
				}
			}
		})
	}
}
