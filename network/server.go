package network

import (
	"log"
	"net"

	"collectd.org/api"
)

// ListenAndDispatch listens on the provided UDP address, parses the received
// packets and dispatches them to the provided dispatcher.
func ListenAndDispatch(address string, d api.Dispatcher) error {
	laddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	var sock *net.UDPConn
	if laddr.IP.IsMulticast() {
		sock, err = net.ListenMulticastUDP("udp", nil /* interface */, laddr)
	} else {
		sock, err = net.ListenUDP("udp", laddr)
	}
	if err != nil {
		return err
	}
	defer sock.Close()

	buf := make([]byte, DefaultBufferSize)
	for {
		n, err := sock.Read(buf)
		if err != nil {
			return err
		}

		valueLists, err := Parse(buf[:n])
		if err != nil {
			log.Printf("error while parsing: %v", err)
			continue
		}

		go dispatch(valueLists, d)
	}
}

func dispatch(valueLists []api.ValueList, d api.Dispatcher) {
	for _, vl := range valueLists {
		if err := d.Dispatch(vl); err != nil {
			log.Printf("error while dispatching: %v", err)
		}
	}
}
