/* Package rpc ... */
package rpc // import "collectd.org/rpc"

import (
	"collectd.org/api"
	"golang.org/x/net/context"
)

// Interface is an idiomatic Go interface for the Collectd gRPC service.
//
// To implement a client, pass a client connection to NewClient() to get back
// an object implementing this interface.
//
// To implement a server, use RegisterServer() to hook an object, which
// implements Interface, up to a gRPC server.
type Interface interface {
	api.Writer
	Query(context.Context, *api.Identifier) (<-chan *api.ValueList, error)
}
