package tracking

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

func TestMetricTracker_Convert(t *testing.T) {
	miSum := MetricIdentity{
		Resource:               pdata.NewResource(),
		InstrumentationLibrary: pdata.NewInstrumentationLibrary(),
		MetricDataType:         pdata.MetricDataTypeSum,
		MetricIsMonotonic:      true,
		MetricName:             "",
		MetricDescription:      "",
		MetricUnit:             "",
		LabelsMap:              pdata.NewStringMap(),
	}
	miIntSum := miSum
	miIntSum.MetricDataType = pdata.MetricDataTypeIntSum

	m := NewMetricTracker(context.Background(), 0)

	tests := []struct {
		name    string
		point   MetricPoint
		wantOut DeltaValue
	}{
		{
			name: "Initial Value recorded",
			point: MetricPoint{
				ObservedTimestamp: 10,
				FloatValue:        100.0,
				IntValue:          100,
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				FloatValue:     100.0,
				IntValue:       100,
			},
		},
		{
			name: "Higher Value Recorded",
			point: MetricPoint{
				ObservedTimestamp: 50,
				FloatValue:        225.0,
				IntValue:          225,
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				FloatValue:     125.0,
				IntValue:       125,
			},
		},
		{
			name: "Lower Value Recorded - No Previous Offset",
			point: MetricPoint{
				ObservedTimestamp: 100,
				FloatValue:        75.0,
				IntValue:          75,
			},
			wantOut: DeltaValue{
				StartTimestamp: 50,
				FloatValue:     75.0,
				IntValue:       75,
			},
		},
		{
			name: "Record delta above first recorded value",
			point: MetricPoint{
				ObservedTimestamp: 150,
				FloatValue:        300.0,
				IntValue:          300,
			},
			wantOut: DeltaValue{
				StartTimestamp: 100,
				FloatValue:     225.0,
				IntValue:       225,
			},
		},
		{
			name: "Lower Value Recorded - Previous Offset Recorded",
			point: MetricPoint{
				ObservedTimestamp: 200,
				FloatValue:        25.0,
				IntValue:          25,
			},
			wantOut: DeltaValue{
				StartTimestamp: 150,
				FloatValue:     25.0,
				IntValue:       25,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			floatPoint := DataPoint{
				Identity: miSum,
				Point:    tt.point,
			}
			intPoint := DataPoint{
				Identity: miIntSum,
				Point:    tt.point,
			}

			if gotOut, valid := m.Convert(floatPoint); !valid || !reflect.DeepEqual(gotOut.StartTimestamp, tt.wantOut.StartTimestamp) || !reflect.DeepEqual(gotOut.FloatValue, tt.wantOut.FloatValue) {
				t.Errorf("MetricTracker.Convert(MetricDataTypeSum) = %v, want %v", gotOut, tt.wantOut)
			}

			if gotOut, valid := m.Convert(intPoint); !valid || !reflect.DeepEqual(gotOut.StartTimestamp, tt.wantOut.StartTimestamp) || !reflect.DeepEqual(gotOut.IntValue, tt.wantOut.IntValue) {
				t.Errorf("MetricTracker.Convert(MetricDataTypeIntSum) = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}

	t.Run("Invalid metric identity", func(t *testing.T) {
		invalidId := miIntSum
		invalidId.MetricDataType = pdata.MetricDataTypeGauge
		_, valid := m.Convert(DataPoint{
			Identity: invalidId,
			Point: MetricPoint{
				ObservedTimestamp: 0,
				FloatValue:        100.0,
				IntValue:          100,
			},
		})
		if valid {
			t.Error("Expected invalid for non cumulative metric")
		}
	})
}

func Test_metricTracker_RemoveStale(t *testing.T) {
	currentTime := pdata.Timestamp(100)
	freshPoint := MetricPoint{
		ObservedTimestamp: currentTime,
	}
	stalePoint := MetricPoint{
		ObservedTimestamp: currentTime - 1,
	}

	type fields struct {
		MaxStale time.Duration
		States   map[string]*State
	}
	tests := []struct {
		name    string
		fields  fields
		wantOut map[string]*State
	}{
		{
			name: "Removes stale entry, leaves fresh entry",
			fields: fields{
				MaxStale: 0, // This logic isn't tested here
				States: map[string]*State{
					"stale": {
						PrevPoint: stalePoint,
					},
					"fresh": {
						PrevPoint: freshPoint,
					},
				},
			},
			wantOut: map[string]*State{
				"fresh": {
					PrevPoint: freshPoint,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncMap := sync.Map{}
			for k, v := range tt.fields.States {
				syncMap.Store(k, v)
			}
			tr := &metricTracker{
				MaxStale: tt.fields.MaxStale,
				States:   syncMap,
			}
			tr.RemoveStale(currentTime)

			gotOut := make(map[string]*State)
			tr.States.Range(func(key, value interface{}) bool {
				gotOut[key.(string)] = value.(*State)
				return true
			})

			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("MetricTracker.RemoveStale() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
