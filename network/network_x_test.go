package network_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/net/nettest"
)

type testPasswordLookup map[string]string

func (l testPasswordLookup) Password(user string) (string, error) {
	pw, ok := l[user]
	if !ok {
		return "", fmt.Errorf("user %q not found", user)
	}
	return pw, nil
}

func TestNetwork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const (
		username = "TestNetwork"
		password = `oi5aGh7oLo0mai5oaG8zei8a`
	)

	conn, err := nettest.NewLocalPacketListener("udp")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ch := make(chan *api.ValueList)
	go func() {
		srv := &network.Server{
			Conn: conn.(*net.UDPConn),
			Writer: api.WriterFunc(func(_ context.Context, vl *api.ValueList) error {
				ch <- vl
				return nil
			}),
			PasswordLookup: testPasswordLookup{
				username: password,
			},
		}

		err := srv.ListenAndWrite(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Server.ListenAndWrite() = %v, want %v", err, context.Canceled)
		}
		close(ch)
	}()

	var want []*api.ValueList
	go func() {
		client, err := network.Dial(conn.LocalAddr().String(),
			network.ClientOptions{
				SecurityLevel: network.Encrypt,
				Username:      username,
				Password:      password,
			})
		if err != nil {
			t.Fatal(err)
		}

		vl := &api.ValueList{
			Identifier: api.Identifier{
				Host:   "example.com",
				Plugin: "TestNetwork",
				Type:   "gauge",
			},
			Time:     time.Unix(1588164686, 0),
			Interval: 10 * time.Second,
			Values:   []api.Value{api.Gauge(42)},
		}

		for i := 0; i < 30; i++ {
			if err := client.Write(ctx, vl); err != nil {
				t.Errorf("client.Write() = %v", err)
				break
			}
			want = append(want, vl.Clone())

			vl.Time = vl.Time.Add(vl.Interval)
		}

		if err := client.Close(); err != nil {
			t.Errorf("client.Close() = %v", err)
		}
	}()

	var got []*api.ValueList
loop:
	for {
		select {
		case vl, ok := <-ch:
			if !ok {
				break loop
			}
			got = append(got, vl)
		case <-time.After(100 * time.Millisecond):
			// cancel the context so the server returns.
			cancel()
		}
	}

	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("sent and received value lists differ (+got/-want):\n%s", diff)
	}
}
