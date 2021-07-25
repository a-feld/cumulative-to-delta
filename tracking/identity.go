package tracking

import (
	"bytes"
	"strconv"

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
	StartTimestamp         pdata.Timestamp
	LabelsMap              pdata.StringMap
}

const A = int32('A')

func (mi *MetricIdentity) Write(b *bytes.Buffer) {
	b.WriteString("t;")
	b.WriteRune(A + int32(mi.MetricDataType))
	b.WriteString("r")
	mi.Resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		b.WriteString(";")
		b.WriteString(k)
		b.WriteString(";")
		b.WriteString(tracetranslator.AttributeValueToString(v))
		return true
	})

	b.WriteString(";i;")
	b.WriteString(mi.InstrumentationLibrary.Name())
	b.WriteString(";")
	b.WriteString(mi.InstrumentationLibrary.Version())

	b.WriteString(";m;")
	b.WriteString(mi.MetricName)
	b.WriteString(";")
	b.WriteString(mi.MetricUnit)

	b.WriteString(";l")
	mi.LabelsMap.Sort().Range(func(k, v string) bool {
		b.WriteString(";")
		b.WriteString(k)
		b.WriteString(";")
		b.WriteString(v)
		return true
	})
	b.WriteString(";s;")
	b.WriteString(strconv.FormatInt(int64(mi.StartTimestamp), 36))
}
