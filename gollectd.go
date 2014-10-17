// Parse data from the collectd network plugin
// This is modeled after collectd's own network.c
package gollectd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

const (
	TypeCounter  = 0
	TypeGauge    = 1
	TypeDerive   = 2
	TypeAbsolute = 3
)

const (
	// Values taken from commit 633c3966f7 of
	// https://github.com/collectd/collectd/commits/master/src/network.h

	ParseHost           = 0x0000
	ParseTime           = 0x0001
	ParsePlugin         = 0x0002
	ParsePluginInstance = 0x0003
	ParseType           = 0x0004
	ParseTypeInstance   = 0x0005
	ParseValues         = 0x0006
	ParseInterval       = 0x0007
	ParseTimeHR         = 0x0008
	ParseIntervalHR     = 0x0009

	// Notifications
	ParseMessage  = 0x0100
	ParseSeverity = 0x0101

	ParseSignature  = 0x0200
	ParseEncryption = 0x0210
)

var ErrorInvalid = errors.New("gollectd: Invalid packet")
var ErrorUnsupported = errors.New("gollectd: Unsupported packet")
var ErrorUnknownType = errors.New("gollectd: Unknown value type")
var ErrorUnknownDataType = errors.New("gollectd: Unknown data source type")

var ValueTypeNames = map[string]uint8{
	"absolute": TypeAbsolute,
	"counter":  TypeCounter,
	"derive":   TypeDerive,
	"gauge":    TypeGauge,
}

var ValueTypeValues = map[uint8]string{
	TypeAbsolute: "absolute",
	TypeCounter:  "counter",
	TypeDerive:   "derive",
	TypeGauge:    "gauge",
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

type Types map[string][]Type

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

func Packets(b []byte, types Types) (*[]Packet, error) {
	packets := make([]Packet, 0)

	buf := bytes.NewBuffer(b)

	var packetHeader struct {
		PartType   uint16
		PartLength uint16
	}

	var err error
	var packet Packet
	var time uint64
	var valueCount uint16
	var valueTypes []uint8

	for buf.Len() > 0 {
		err = binary.Read(buf, binary.BigEndian, &packetHeader)
		if err != nil {
			return nil, err
		}

		if packetHeader.PartLength < 5 {
			return nil, ErrorInvalid
		}

		nextPos := int(packetHeader.PartLength) - 4

		partBytes := buf.Next(nextPos)
		if len(partBytes) < nextPos {
			return nil, ErrorInvalid
		}

		partBuffer := bytes.NewBuffer(partBytes)

		switch packetHeader.PartType {
		case ParseEncryption:
			return nil, ErrorUnsupported
		case ParseHost:
			str := partBuffer.String()
			packet.Hostname = str[0 : len(str)-1]
		case ParseInterval:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.Interval = time
		case ParseIntervalHR:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.IntervalHR = time
		case ParseMessage:
			// ignore (notification)
		case ParsePlugin:
			str := partBuffer.String()
			packet.Plugin = str[0 : len(str)-1]
		case ParsePluginInstance:
			str := partBuffer.String()
			packet.PluginInstance = str[0 : len(str)-1]
		case ParseSeverity:
			// ignore (notification)
		case ParseSignature:
			return nil, ErrorUnsupported
		case ParseTime:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.Time = time
		case ParseTimeHR:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.TimeHR = time
		case ParseType:
			str := partBuffer.String()
			packet.Type = str[0 : len(str)-1]
		case ParseTypeInstance:
			str := partBuffer.String()
			packet.TypeInstance = str[0 : len(str)-1]
		case ParseValues:
			err = binary.Read(partBuffer, binary.BigEndian, &valueCount)
			if err != nil {
				return nil, err
			}

			valueTypes = make([]uint8, valueCount, valueCount)
			packet.Values = make([]Value, valueCount, valueCount)
			var packetValue Value

			err = binary.Read(partBuffer, binary.BigEndian, &valueTypes)
			if err != nil {
				return nil, err
			}

			for i, t := range valueTypes {
				packetValue.Type = t

				if typeName, ok := ValueTypeValues[t]; ok {
					packetValue.TypeName = typeName
				}

				if _, ok := types[packet.Type]; ok {
					packetValue.Name = types[packet.Type][i].Name
				}

				switch t {
				case TypeAbsolute:
					var value uint64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeCounter:
					var value uint64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeDerive:
					var value int64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeGauge:
					var value float64
					err = binary.Read(partBuffer, binary.LittleEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				default:
					return nil, ErrorUnknownType
				}

				packet.Values[i] = packetValue
			}

			packets = append(packets, packet)
		default:
			// Ignore unknown fields
		}
	}

	return &packets, nil
}

func TypesDB(path string) (Types, error) {
	// See https://collectd.org/documentation/manpages/types.db.5.shtml

	types := make(Types)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var dsSpec Type

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "\t", " ", -1)
		fields := strings.Split(line, " ")

		if len(fields) < 2 {
			continue
		}

		if string(fields[0]) == "#" {
			continue
		}

		dataSet := fields[0]
		types[dataSet] = make([]Type, 0)

		for _, dataSources := range fields[1:] {
			if len(dataSources) == 0 {
				continue
			}

			dataSources = strings.Trim(dataSources, ",")

			dataSource := strings.Split(dataSources, ":")

			if len(dataSource) != 4 {
				// set ErrorUnknownDataType somehow
				continue
			}

			dsSpec.Name = dataSource[0]

			if dsType, ok := ValueTypeNames[strings.ToLower(dataSource[1])]; ok {
				dsSpec.Type = dsType
			} else {
				// set ErrorUnknownDataType somehow
				continue
			}

			dsSpec.Min = dataSource[2]
			dsSpec.Max = dataSource[3]

			types[dataSet] = append(types[dataSet], dsSpec)
		}
	}

	return types, nil
}
