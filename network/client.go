package network

import (
	"net"

	"collectd.org/api"
)

// Client is a connection to a collectd server. It implements the
// api.Dispatcher interface.
type Client struct {
	udp    net.Conn
	buffer *Buffer
}

// Dial connects to the collectd server at address. "address" must be a network
// address accepted by net.Dial().
func Dial(address string) (*Client, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		udp:    c,
		buffer: NewBuffer(c),
	}, nil
}

// DialSigned connects to the collectd server at "address". Data is signed with
// the given username and password.
func DialSigned(address, username, password string) (*Client, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		udp:    c,
		buffer: NewBufferSigned(c, username, password),
	}, nil
}

// DialEncrypted connects to the collectd server at "address". Data is
// encrypted with the given username and password.
func DialEncrypted(address, username, password string) (*Client, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		udp:    c,
		buffer: NewBufferEncrypted(c, username, password),
	}, nil
}

// Dispatch adds a ValueList to the internal buffer. Data is only written to
// the network when the buffer is full.
func (c *Client) Dispatch(vl api.ValueList) error {
	return c.buffer.WriteValueList(vl)
}

// Flush writes the contents of the underlying buffer to the network
// immediately.
func (c *Client) Flush() error {
	return c.buffer.Flush()
}

// Close closes a connection. You must not use "c" after this call.
func (c *Client) Close() error {
	if err := c.buffer.Flush(); err != nil {
		return err
	}

	if err := c.udp.Close(); err != nil {
		return err
	}

	c.buffer = nil
	return nil
}
