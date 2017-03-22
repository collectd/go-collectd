package network // import "collectd.org/network"

import (
	"context"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"collectd.org/format"
)

// This example demonstrates how to listen to encrypted network traffic and
// dump it to STDOUT using format.Putval.
func ExampleServer_ListenAndWrite() {
	srv := &Server{
		Addr:           net.JoinHostPort("::", DefaultService),
		Writer:         format.NewPutval(os.Stdout),
		PasswordLookup: NewAuthFile("/etc/collectd/users"),
	}

	// blocks
	log.Fatal(srv.ListenAndWrite(context.Background()))
}

// This example demonstrates how to forward received IPv6 multicast traffic to
// a unicast address, using PSK encryption.
func ExampleListenAndWrite() {
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
	log.Fatal(ListenAndWrite(context.Background(), ":"+DefaultService, client))
}

func TestServer_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var srvErr error
	go func() {
		srv := &Server{
			Addr: "localhost:" + DefaultService,
		}

		srvErr = srv.ListenAndWrite(ctx)
		wg.Done()
	}()

	// wait for a bit, then shut down the server
	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	if srvErr != context.Canceled {
		t.Errorf("srvErr = %#v, want %#v", srvErr, context.Canceled)
	}
}
