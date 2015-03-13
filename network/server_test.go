package network

import (
	"log"
	"net"
)

func Example_proxy() {
	client, err := Dial(net.JoinHostPort("example.com", DefaultService), ClientOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sopts := ServerOptions{
		PasswordLookup: NewAuthFile("/path/to/file"),
		BufferSize:     1500,
	}

	// blocks
	err = ListenAndDispatch(net.JoinHostPort("::", DefaultService), client, sopts)
	if err != nil {
		log.Fatal(err)
	}
}
