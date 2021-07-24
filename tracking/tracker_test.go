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
				Value:             100.0,
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				Value:          100.0,
			},
		},
		{
			name: "Higher Value Recorded",
			point: MetricPoint{
				ObservedTimestamp: 50,
				Value:             225.0,
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				Value:          125.0,
			},
		},
		{
			name: "Lower Value Recorded - No Previous Offset",
			point: MetricPoint{
				ObservedTimestamp: 100,
				Value:             75.0,
			},
			wantOut: DeltaValue{
				StartTimestamp: 50,
				Value:          75.0,
			},
		},
		{
			name: "Record delta above first recorded value",
			point: MetricPoint{
				ObservedTimestamp: 150,
				Value:             300.0,
			},
			wantOut: DeltaValue{
				StartTimestamp: 100,
				Value:          225.0,
			},
		},
		{
			name: "Lower Value Recorded - Previous Offset Recorded",
			point: MetricPoint{
				ObservedTimestamp: 200,
				Value:             25.0,
			},
			wantOut: DeltaValue{
				StartTimestamp: 150,
				Value:          25.0,
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
			intPoint.Point.Value = int64(tt.point.Value.(float64))
			wantOutInt := tt.wantOut
			wantOutInt.Value = int64(wantOutInt.Value.(float64))

			if gotOut := m.Convert(floatPoint); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("MetricTracker.Convert(MetricDataTypeSum) = %v, want %v", gotOut, tt.wantOut)
			}

			if gotOut := m.Convert(intPoint); !reflect.DeepEqual(gotOut, wantOutInt) {
				t.Errorf("MetricTracker.Convert(MetricDataTypeIntSum) = %v, want %v", gotOut, wantOutInt)
			}
		})
	}

	t.Run("Invalid metric identity", func(t *testing.T) {
		invalidId := miIntSum
		invalidId.MetricDataType = pdata.MetricDataTypeGauge
		delta := m.Convert(DataPoint{
			Identity: invalidId,
			Point: MetricPoint{
				ObservedTimestamp: 0,
				Value:             100.0,
			},
		})
		if delta.Value != nil {
			t.Error("Expected delta value to be nil for non cumulative metric")
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
