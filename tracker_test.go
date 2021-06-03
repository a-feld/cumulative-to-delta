package cumulativetodelta

import (
	"reflect"
	"sync"
	"testing"
)

func TestMetricTracker_Record(t *testing.T) {
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
		want   map[MetricIdentity]State
	}{
		{
			name: "Initial Value Recorded",
			fields: fields{
				States: make(map[MetricIdentity]State),
			},
			args: args{
				in: Metric{
					Name:  "foobar",
					Value: 100,
				},
			},
			want: map[MetricIdentity]State{
				{}: {
					RunningTotal: 100,
					LatestValue:  100,
				},
			},
		},
		{
			name: "Higher Value Recorded",
			fields: fields{
				States: map[MetricIdentity]State{
					{}: {
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
			want: map[MetricIdentity]State{
				{}: {
					RunningTotal: 200,
					LatestValue:  200,
				},
			},
		},
		{
			name: "Lower Value Recorded - No Offset",
			fields: fields{
				States: map[MetricIdentity]State{
					{}: {
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
			want: map[MetricIdentity]State{
				{}: {
					RunningTotal: 280,
					LatestValue:  80,
					Offset:       200,
				},
			},
		},
		{
			name: "Lower Value Recorded - With Existing Offset",
			fields: fields{
				States: map[MetricIdentity]State{
					{}: {
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
			want: map[MetricIdentity]State{
				{}: {
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
			m := &MetricTracker{
				mu:     sync.Mutex{},
				States: tt.fields.States,
			}
			if got := m.Flush(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricTracker.Flush() = %v, want %v", got, tt.want)
			}
		})
	}
}
