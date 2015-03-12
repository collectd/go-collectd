package network

import (
	"log"
	"net"
)

func Example_proxy() {
	client, err := Dial(net.JoinHostPort("example.com", DefaultService))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// blocks
	err = ListenAndDispatch(net.JoinHostPort("::", DefaultService), client)
	if err != nil {
		log.Fatal(err)
	}
}
