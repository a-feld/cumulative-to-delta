package tracking

import (
	"reflect"
	"testing"

	"go.opentelemetry.io/collector/model/pdata"
)

func TestMetricTracker_Convert(t *testing.T) {
	mi := MetricIdentity{
		Resource:               pdata.NewResource(),
		InstrumentationLibrary: pdata.NewInstrumentationLibrary(),
		MetricDataType:         pdata.MetricDataTypeSum,
		MetricIsMonotonic:      true,
		MetricName:             "",
		MetricDescription:      "",
		MetricUnit:             "",
		LabelsMap:              pdata.NewStringMap(),
	}

	m := MetricTracker{}

	tests := []struct {
		name      string
		dataPoint DataPoint
		wantOut   DeltaValue
	}{
		{
			name: "Initial Value recorded",
			dataPoint: DataPoint{
				Identity: mi,
				Point: MetricPoint{
					ObservedTimestamp: 10,
					Value:             100.0,
				},
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				Value:          100.0,
			},
		},
		{
			name: "Higher Value Recorded",
			dataPoint: DataPoint{
				Identity: mi,
				Point: MetricPoint{
					ObservedTimestamp: 50,
					Value:             225.0,
				},
			},
			wantOut: DeltaValue{
				StartTimestamp: 10,
				Value:          125.0,
			},
		},
		{
			name: "Lower Value Recorded - No Previous Offset",
			dataPoint: DataPoint{
				Identity: mi,
				Point: MetricPoint{
					ObservedTimestamp: 100,
					Value:             75.0,
				},
			},
			wantOut: DeltaValue{
				StartTimestamp: 50,
				Value:          75.0,
			},
		},
		{
			name: "Record delta above first recorded value",
			dataPoint: DataPoint{
				Identity: mi,
				Point: MetricPoint{
					ObservedTimestamp: 150,
					Value:             300.0,
				},
			},
			wantOut: DeltaValue{
				StartTimestamp: 100,
				Value:          225.0,
			},
		},
		{
			name: "Lower Value Recorded - Previous Offset Recorded",
			dataPoint: DataPoint{
				Identity: mi,
				Point: MetricPoint{
					ObservedTimestamp: 200,
					Value:             25.0,
				},
			},
			wantOut: DeltaValue{
				StartTimestamp: 150,
				Value:          25.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := m.Convert(tt.dataPoint); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("MetricTracker.Convert() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
