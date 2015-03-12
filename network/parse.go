package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"

	"collectd.org/api"
	"collectd.org/cdtime"
)

// Common error conditions.
var (
	ErrorInvalid     = errors.New("Invalid packet")
	ErrorUnsupported = errors.New("Unsupported packet")
)

// Parse parses the binary network format and returns a slice of ValueLists. If
// a parse error is encountered, all ValueLists parsed to this point are
// returned as well as the error. Unknown "parts" are silently ignored.
func Parse(b []byte) ([]api.ValueList, error) {
	var valueLists []api.ValueList

	var state api.ValueList
	buf := bytes.NewBuffer(b)

	for buf.Len() > 0 {
		partType := binary.BigEndian.Uint16(buf.Next(2))
		partLength := int(binary.BigEndian.Uint16(buf.Next(2)))

		if partLength < 5 || partLength-4 > buf.Len() {
			return valueLists, fmt.Errorf("invalid length %d", partLength)
		}

		// First 4 bytes were already read
		partLength -= 4

		payload := buf.Next(partLength)
		if len(payload) != partLength {
			return valueLists, fmt.Errorf("invalid length: want %d, got %d", partLength, len(payload))
		}

		switch partType {
		case typeHost, typePlugin, typePluginInstance, typeType, typeTypeInstance:
			str, err := parseString(payload)
			if err != nil {
				return valueLists, err
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
		case typeInterval, typeIntervalHR, typeTime, typeTimeHR:
			i, err := parseInt(payload)
			if err != nil {
				return valueLists, err
			}
			switch partType {
			case typeInterval:
				state.Interval = time.Duration(i) * time.Second
			case typeIntervalHR:
				state.Interval = cdtime.Time(i).Duration()
			case typeTime:
				state.Time = time.Unix(int64(i), 0)
			case typeTimeHR:
				state.Time = cdtime.Time(i).Time()
			}
		case typeValues:
			vl := state
			var err error
			if vl.Values, err = parseValues(payload); err != nil {
				return valueLists, err
			}

			valueLists = append(valueLists, vl)

		default:
			log.Printf("ignoring field of type %#x", partType)
		}
	}

	return valueLists, nil
}

func parseValues(b []byte) ([]api.Value, error) {
	buffer := bytes.NewBuffer(b)

	var n uint16
	if err := binary.Read(buffer, binary.BigEndian, &n); err != nil {
		return nil, err
	}

	if int(n*9) != buffer.Len() {
		return nil, fmt.Errorf("parseValues: length mismatch: %d vs %d", n*9, buffer.Len())
	}

	types := make([]byte, n)
	values := make([]api.Value, n)

	if _, err := buffer.Read(types); err != nil {
		return nil, err
	}

	for i, typ := range types {
		switch typ {
		case dsTypeGauge:
			var v float64
			if err := binary.Read(buffer, binary.LittleEndian, &v); err != nil {
				return nil, err
			}
			values[i] = api.Gauge(v)

		case dsTypeDerive, dsTypeCounter:
			var v int64
			if err := binary.Read(buffer, binary.BigEndian, &v); err != nil {
				return nil, err
			}
			values[i] = api.Derive(v)

		case dsTypeAbsolute:
			return nil, ErrorUnsupported

		default:
			return nil, fmt.Errorf("parseValues: invalid data source %d", typ)
		}
	}

	return values, nil
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
