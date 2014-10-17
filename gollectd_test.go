package gollectd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FormatTests struct {
	Description string
	Expected    string
	Packet      Packet
}

var formatTests = []FormatTests{
	{
		Description: "leeloo/cpu-0/cpu-idle",
		Expected:    "leeloo/cpu-0/cpu-idle",
		Packet: Packet{
			Hostname:       "leeloo",
			Plugin:         "cpu",
			PluginInstance: "0",
			Type:           "cpu",
			TypeInstance:   "idle",
		},
	},
	{
		Description: "alyja/memory/memory-used",
		Expected:    "alyja/memory/memory-used",
		Packet: Packet{
			Hostname:     "alyja",
			Plugin:       "memory",
			Type:         "memory",
			TypeInstance: "used",
		},
	},
	{
		Description: "wanda/disk-hdc1/disk_octets",
		Expected:    "wanda/disk-hdc1/disk_octets",
		Packet: Packet{
			Hostname:       "wanda",
			Plugin:         "disk",
			PluginInstance: "hdc1",
			Type:           "disk_octets",
		},
	},
	{
		Description: "leeloo/load/load",
		Expected:    "leeloo/load/load",
		Packet: Packet{
			Hostname: "leeloo",
			Plugin:   "load",
			Type:     "load",
		},
	},
}

func TestFormat(t *testing.T) {
	assert := assert.New(t)

	for _, test := range formatTests {
		got := test.Packet.FormatName()
		assert.Equal(test.Expected, got, test.Description)
	}
}
