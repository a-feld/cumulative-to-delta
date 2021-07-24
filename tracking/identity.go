package tracking

import (
	"bytes"
	"sync"

	"go.opentelemetry.io/collector/model/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

// Allocate a minimum of 64 bytes to the builder initially
const initialBytes = 64

var identityBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, initialBytes))
	},
}

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
const A = int32('A')

func (mi *MetricIdentity) AsString() string {
	b := identityBufferPool.Get().(*bytes.Buffer)
	defer identityBufferPool.Put(b)
	b.Reset()
	b.WriteString("t;")
	b.WriteRune(A + int32(mi.MetricDataType))
	b.WriteString("r;")
	mi.Resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		b.WriteString(k)
		b.WriteString(";")
		b.WriteString(tracetranslator.AttributeValueToString(v))
		b.WriteString(";")
		return true
	})

	b.WriteString(";i;")
	b.WriteString(mi.InstrumentationLibrary.Name())
	b.WriteString(";")
	b.WriteString(mi.InstrumentationLibrary.Version())

	b.WriteString(";m;")
	b.WriteString(mi.MetricName)
	b.WriteString(";")
	b.WriteString(mi.MetricDescription)
	b.WriteString(";")
	b.WriteString(mi.MetricUnit)

	b.WriteString(";l;")
	mi.LabelsMap.Sort().Range(func(k, v string) bool {
		b.WriteString(k)
		b.WriteString(";")
		b.WriteString(v)
		b.WriteString(";")
		return true
	})
	return b.String()
}
