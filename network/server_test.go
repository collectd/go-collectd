package network // import "collectd.org/network"

import (
	"log"
	"net"
	"os"

	"collectd.org/format"
)

// This example demonstrates how to listen to encrypted network traffic and
// dump it to STDOUT using format.Putval.
func ExampleListenAndDispatch_decrypt() {
	opts := ServerOptions{
		PasswordLookup: NewAuthFile("/etc/collectd/users"),
	}

	// blocks
	if err := ListenAndDispatch(net.JoinHostPort("::", DefaultService), format.NewPutval(os.Stdout), opts); err != nil {
		log.Fatal(err)
	}
}

// This example demonstrates how to forward received IPv6 multicast traffic to
// a unicast address, using PSK encryption.
func ExampleListenAndDispatch_proxy() {
	opts := ClientOptions{
		SecurityLevel: Encrypt,
		Username:      "collectd",
		Password:      "dieXah7e",
	}
	client, err := Dial(net.JoinHostPort("example.com", DefaultService), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// blocks
	if err = ListenAndDispatch(net.JoinHostPort(DefaultIPv6Address, DefaultService), client, ServerOptions{}); err != nil {
		log.Fatal(err)
	}
}
