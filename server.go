package gollectd

import (
	"log"
	"net"
)

// Listen for collectd network packets, parse , and send them over a channel
func Listen(addr string, c chan Packet) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalln("fatal: failed to resolve address", err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalln("fatal: failed to listen", err)
	}

	for {
		// 1452 is collectd 5's default buffer size. See:
		// https://collectd.org/wiki/index.php/Binary_protocol
		buf := make([]byte, 1452)

		n, err := conn.Read(buf[:])
		if err != nil {
			log.Println("error: Failed to receive packet", err)
			continue
		}

		packets, err := Packets(buf[0:n])
		if err != nil {
			log.Println("error: Failed to receive packet", err)
			continue
		}

		for _, p := range *packets {
			c <- p
		}
	}
}
