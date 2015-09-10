package network // import "collectd.org/network"

import (
	"fmt"
	"log"
	"net"
	"sync"

	"collectd.org/api"
)

// ListenAndWrite listens on the provided UDP address, parses the received
// packets and writes them to the provided api.Writer.
// This is a convenience function for a minimally configured server. If you
// need more control, see the "Server" type below.
func ListenAndWrite(address string, d api.Writer) error {
	srv := &Server{
		Addr:   address,
		Writer: d,
	}

	return srv.ListenAndWrite()
}

// Server holds parameters for running a collectd server.
type Server struct {
	Addr           string         // UDP address to listen on.
	Writer         api.Writer     // Object used to send incoming ValueLists to.
	BufferSize     uint16         // Maximum packet size to accept.
	PasswordLookup PasswordLookup // User to password lookup.
	SecurityLevel  SecurityLevel  // Minimal required security level
	// Interface is the name of the interface to use when subscribing to a
	// multicast group. Has no effect when using unicast.
	Interface string
	shutdown  chan (bool) // channel used to enable clean socket shutdown
}

// ListenAndWrite listens on the provided UDP address, parses the received
// packets and writes them to the provided api.Writer. This is a blocking call.
func (srv *Server) ListenAndWrite() error {
	return srv.listenAndWrite(nil)
}

// ListenAndWriteAsync listens on the provided UDP address, parses the received
// packets and writes them to the provided api.Writer. Runs in goroutine and
// returns to caller when socket is listening.
func (srv *Server) ListenAndWriteAsync() {
	var wg sync.WaitGroup
	wg.Add(1)
	go srv.listenAndWrite(&wg)
	wg.Wait()
}

func (srv *Server) listenAndWrite(wg *sync.WaitGroup) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":" + DefaultService
	}

	laddr, err := net.ResolveUDPAddr("udp", srv.Addr)
	if err != nil {
		return err
	}

	var sock *net.UDPConn
	if laddr.IP != nil && laddr.IP.IsMulticast() {
		var ifi *net.Interface
		if srv.Interface != "" {
			if ifi, err = net.InterfaceByName(srv.Interface); err != nil {
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

	if srv.BufferSize <= 0 {
		srv.BufferSize = DefaultBufferSize
	}
	buf := make([]byte, srv.BufferSize)

	popts := ParseOpts{
		PasswordLookup: srv.PasswordLookup,
		SecurityLevel:  srv.SecurityLevel,
	}

	values := make(chan []api.ValueList)
	errors := make(chan error)
	srv.shutdown = make(chan bool)

	go func() {
		for {
			n, err := sock.Read(buf)
			if err != nil {
				errors <- err
				continue
			}

			valueLists, err := Parse(buf[:n], popts)
			if err != nil {
				errors <- err
				continue
			}
			values <- valueLists
		}
	}()

	if wg != nil {
		wg.Done()
	}

	for {
		select {
		case valueLists := <-values:
			go dispatch(valueLists, srv.Writer)
		case err := <-errors:
			if err != nil {
				log.Printf("error while parsing: %v", err)
				return err
			}
		case <-srv.shutdown:
			srv.shutdown = nil
			return nil
		}
	}
}

// Shutdown acts on a server that is actively listening, shutting down it's UDP socket
func (srv *Server) Shutdown() error {
	if srv.shutdown == nil {
		return fmt.Errorf("Cannot shutdown server. Not listening.")
	}
	srv.shutdown <- true
	return nil
}

func dispatch(valueLists []api.ValueList, d api.Writer) {
	for _, vl := range valueLists {
		if err := d.Write(vl); err != nil {
			log.Printf("error while dispatching: %v", err)
		}
	}
}
