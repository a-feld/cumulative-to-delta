package tracking

import (
	"reflect"
	"testing"

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

	m := MetricTracker{}

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
}
