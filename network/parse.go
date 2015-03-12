package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"collectd.org/api"
	cdtime "collectd.org/time"
)

const (
	// Values taken from commit 633c3966f7 of
	// https://github.com/collectd/collectd/commits/master/src/network.h

	// Notifications
	ParseMessage  = 0x0100
	ParseSeverity = 0x0101
)

var ErrorInvalid = errors.New("Invalid packet")
var ErrorUnsupported = errors.New("Unsupported packet")
var ErrorUnknownType = errors.New("Unknown value type")
var ErrorUnknownDataType = errors.New("Unknown data source type")

var ValueTypeNames = map[string]uint8{
	"absolute": dsTypeAbsolute,
	"counter":  dsTypeCounter,
	"derive":   dsTypeDerive,
	"gauge":    dsTypeGauge,
}

var ValueTypeValues = map[uint8]string{
	dsTypeAbsolute: "absolute",
	dsTypeCounter:  "counter",
	dsTypeDerive:   "derive",
	dsTypeGauge:    "gauge",
}

type Packet struct {
	Hostname       string
	Interval       uint64
	IntervalHR     uint64
	Plugin         string
	PluginInstance string
	Time           uint64
	TimeHR         uint64
	Type           string
	TypeInstance   string
	Values         []Value
}

type Type struct {
	Name string
	Type uint8
	Min  string
	Max  string
}

type Types map[string][]*Type

type Value struct {
	Name     string
	Type     uint8
	TypeName string
	Value    float64
}

func (p Packet) FormatName() string {
	metricName := p.Hostname + "/" + p.Plugin

	if len(p.PluginInstance) > 0 {
		metricName += "-" + p.PluginInstance
	}

	metricName += "/" + p.Type

	if len(p.TypeInstance) > 0 {
		metricName += "-" + p.TypeInstance
	}

	return metricName
}

func Parse(b []byte, types Types) ([]api.ValueList, []error) {
	var valueLists []api.ValueList
	var errors []error

	var state api.ValueList
	buf := bytes.NewBuffer(b)

	for buf.Len() > 0 {
		partType := binary.BigEndian.Uint16(buf.Next(2))
		partLength := int(binary.BigEndian.Uint16(buf.Next(2)))

		if partLength < 5 || partLength-4 > buf.Len() {
			errors = append(errors, ErrorInvalid)
			return valueLists, errors
		}

		// First 4 bytes were already read
		partLength -= 4

		payload := buf.Next(partLength)
		if len(payload) != partLength {
			errors = append(errors, ErrorInvalid)
			return valueLists, errors
		}

		switch partType {
		case typeHost, typePlugin, typePluginInstance, typeType, typeTypeInstance:
			str, err := parseString(payload)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			switch partType {
			case typeHost:
				state.Identifier.Host = str
			case typePlugin:
				state.Identifier.Plugin = str
			case typePluginInstance:
				state.Identifier.PluginInstance = str
			case typeType:
				state.Identifier.Type = str
			case typeTypeInstance:
				state.Identifier.TypeInstance = str
			}
		case typeInterval:
			i, err := parseInt(payload)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			state.Interval = time.Duration(i) * time.Second
		case typeIntervalHR:
			d, err := parseDuration(payload)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			state.Interval = d
		case typeTime:
			i, err := parseInt(payload)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			state.Time = time.Unix(int64(i), 0)
		case typeTimeHR:
			t, err := parseTime(payload)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			state.Time = t
		case typeValues:
			vl := state
			var err error
			if vl.Values, err = parseValues(payload); err != nil {
				errors = append(errors, err)
				continue
			}

			valueLists = append(valueLists, vl)

		default:
			// Ignore unknown fields
		}
	}

	return valueLists, nil
}

func parseValues(b []byte) ([]api.Value, error) {
	if len(b)%9 != 0 {
		return nil, ErrorInvalid
	}

	n := len(b) / 9
	types := b[:n]
	buffer := bytes.NewBuffer(b[n:])
	values := make([]api.Value, n)

	for i, typ := range types {
		switch typ {
		case dsTypeGauge:
			var f float64
			if err := binary.Read(buffer, binary.LittleEndian, &f); err != nil {
				return nil, err
			}
			values[i] = api.Gauge(f)

		case dsTypeDerive, dsTypeCounter:
			var i int64
			if err := binary.Read(buffer, binary.BigEndian, &i); err != nil {
				return nil, err
			}
			values[i] = api.Derive(i)

		case dsTypeAbsolute:
			return nil, ErrorUnsupported

		default:
			return nil, ErrorInvalid
		}
	}

	return values, nil
}

func parseTime(b []byte) (time.Time, error) {
	s, err := parseInt(b)
	if err != nil {
		return time.Time{}, err
	}

	return cdtime.Time(s).Time(), nil
}

func parseDuration(b []byte) (time.Duration, error) {
	s, err := parseInt(b)
	if err != nil {
		return time.Duration(0), err
	}

	return cdtime.Time(s).Duration(), nil
}

func parseInt(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, ErrorInvalid
	}

	var i uint64
	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, &i); err != nil {
		return 0, err
	}

	return i, nil
}

func parseString(b []byte) (string, error) {
	if b[len(b)-1] != 0 {
		return "", ErrorInvalid
	}

	buf := bytes.NewBuffer(b[:len(b)-1])
	return buf.String(), nil
}

func TypesDBFile(path string) (Types, error) {
	// See https://collectd.org/documentation/manpages/types.db.5.shtml

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return TypesDB(b)
}

func TypesDB(b []byte) (Types, error) {
	types := make(Types)
	content := string(b)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Skip empty & comment lines
		if line == "" || line[0] == '#' {
			continue
		}
		dataSetName, dataSetSources, err := ParseDataSet(line)
		if err != nil {
			return nil, fmt.Errorf("Line %d: %s", i+1, err.Error())
		}

		types[dataSetName] = dataSetSources
	}

	return types, nil
}

// ParseDataSet parses one line from a collectd types.db file and returns
// the data-set name & a Type struct
func ParseDataSet(s string) (string, []*Type, error) {
	splitFn := func(c rune) bool {
		return c == '\t' || c == ' ' || c == ','
	}
	fields := strings.FieldsFunc(s, splitFn)

	if len(fields) < 2 {
		return "", nil, fmt.Errorf("Minimum of 2 fields required \"%s\"", s)
	}

	dataSetName := fields[0]
	dataSetSources := make([]*Type, len(fields[1:]))

	// Parse each data source
	for i, field := range fields[1:] {
		// Split data source fields
		dsFields := strings.Split(field, ":")
		if len(dsFields) != 4 {
			return "", nil, fmt.Errorf("Exactly 4 fields required \"%s\"", field)
		}

		// Parse data source type
		dsTypeStr := strings.ToLower(dsFields[1])
		dsType, ok := ValueTypeNames[dsTypeStr]
		if !ok {
			return "", nil, fmt.Errorf("Invalid data-source type \"%s\"", dsTypeStr)
		}

		dataSetSources[i] = &Type{
			Name: dsFields[0],
			Type: dsType,
			Min:  dsFields[2],
			Max:  dsFields[3],
		}
	}

	return dataSetName, dataSetSources, nil
}
