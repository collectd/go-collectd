package network

import (
	"log"
	"net"
	"time"

	"collectd.org/api"
)

func Example_client() {
	conn, err := Dial(net.JoinHostPort("example.com", DefaultService))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	vl := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Now(),
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Gauge(42.0)},
	}

	if err := conn.WriteValueList(vl); err != nil {
		log.Fatal(err)
	}
}
