package main

import (
	"fmt"
	collectd "github.com/kimor79/gollectd"
)

func main() {
	c := make(chan collectd.Packet)
	go collectd.Listen("127.0.0.1:25826", c)

	for {
		packet := <-c
		fmt.Printf("%+v\n", packet)
	}
}
