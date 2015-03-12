/*
Package network implements collectd's binary network protocol.
*/
package network

import (
	"net"

	"collectd.org/api"
)

// Well-known addresses and port.
const (
	DefaultIPv4Address = "239.192.74.66"
	DefaultIPv6Address = "ff18::efc0:4a42"
	DefaultService     = "25826"
)

// Conn is a client connection to a collectd server. It implements the
// api.Dispatcher interface.
type Conn struct {
	udp    net.Conn
	buffer *Buffer
}

// Dial connects to the collectd server at address. "address" must be a network
// address accepted by net.Dial().
func Dial(address string) (*Conn, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Conn{
		udp:    c,
		buffer: NewBuffer(c),
	}, nil
}

// DialSigned connects to the collectd server at "address". Data is signed with
// the given username and password.
func DialSigned(address, username, password string) (*Conn, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Conn{
		udp:    c,
		buffer: NewBufferSigned(c, username, password),
	}, nil
}

// DialEncrypted connects to the collectd server at "address". Data is
// encrypted with the given username and password.
func DialEncrypted(address, username, password string) (*Conn, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Conn{
		udp:    c,
		buffer: NewBufferEncrypted(c, username, password),
	}, nil
}

// Dispatch adds a ValueList to the internal buffer. Data is only written to
// the network when the buffer is full.
func (c *Conn) Dispatch(vl api.ValueList) error {
	return c.buffer.WriteValueList(vl)
}

// Flush writes the contents of the underlying buffer to the network
// immediately.
func (c *Conn) Flush() error {
	return c.buffer.Flush()
}

// Close closes a connection. You must not use "c" after this call.
func (c *Conn) Close() error {
	if err := c.buffer.Flush(); err != nil {
		return err
	}

	if err := c.udp.Close(); err != nil {
		return err
	}

	c.buffer = nil
	return nil
}
