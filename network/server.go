package network

import (
	"log"
	"net"

	"collectd.org/api"
)

type ServerOptions struct {
	PasswordLookup PasswordLookup
	Interface      string
	BufferSize     uint16
}

// ListenAndDispatch listens on the provided UDP address, parses the received
// packets and dispatches them to the provided dispatcher.
func ListenAndDispatch(address string, d api.Dispatcher, opts ServerOptions) error {
	laddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	var sock *net.UDPConn
	if laddr.IP.IsMulticast() {
		var ifi *net.Interface
		if opts.Interface != "" {
			if ifi, err = net.InterfaceByName(opts.Interface); err != nil {
				return err
			}
		}
		sock, err = net.ListenMulticastUDP("udp", ifi, laddr)
	} else {
		sock, err = net.ListenUDP("udp", laddr)
	}
	if err != nil {
		return err
	}
	defer sock.Close()

	if opts.BufferSize <= 0 {
		opts.BufferSize = DefaultBufferSize
	}
	buf := make([]byte, opts.BufferSize)

	popts := ParseOpts{
		PasswordLookup: opts.PasswordLookup,
	}

	for {
		n, err := sock.Read(buf)
		if err != nil {
			return err
		}

		valueLists, err := Parse(buf[:n], popts)
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
