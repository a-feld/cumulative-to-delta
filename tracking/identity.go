package tracking

import (
	"bytes"
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

type MetricIdentity struct {
	Resource               pdata.Resource
	InstrumentationLibrary pdata.InstrumentationLibrary
	MetricDataType         pdata.MetricDataType
	MetricIsMonotonic      bool
	MetricName             string
	MetricDescription      string
	MetricUnit             string
	LabelsMap              pdata.StringMap
}

// Derived from counting the minimum required length of the strings being
// written to an identity
const initialBytes = 17

func (mi *MetricIdentity) AsString() string {
	h := bytes.Buffer{}
	h.Grow(initialBytes)
	h.WriteString("t;")
	h.WriteString(fmt.Sprintf("%d", mi.MetricDataType))
	h.WriteString("r;")
	mi.Resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		h.WriteString(k)
		h.WriteString(";")
		h.WriteString(tracetranslator.AttributeValueToString(v))
		h.WriteString(";")
		return true
	})

	h.WriteString(";i;")
	h.WriteString(mi.InstrumentationLibrary.Name())
	h.WriteString(";")
	h.WriteString(mi.InstrumentationLibrary.Version())

	h.WriteString(";m;")
	h.WriteString(mi.MetricName)
	h.WriteString(";")
	h.WriteString(mi.MetricDescription)
	h.WriteString(";")
	h.WriteString(mi.MetricUnit)

	h.WriteString(";l;")
	mi.LabelsMap.Sort().Range(func(k, v string) bool {
		h.WriteString(k)
		h.WriteString(";")
		h.WriteString(v)
		h.WriteString(";")
		return true
	})
	return h.String()
}
