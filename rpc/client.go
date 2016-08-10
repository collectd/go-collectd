package rpc // import "collectd.org/rpc"

import (
	"io"
	"log"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// client is a wrapper around pb.CollectdClient implementing Interface.
type client struct {
	ctx    context.Context
	client pb.CollectdClient
}

func NewClient(ctx context.Context, conn *grpc.ClientConn) Interface {
	return &client{
		ctx:    ctx,
		client: pb.NewCollectdClient(conn),
	}
}

func (c *client) Query(ctx context.Context, id *api.Identifier) (<-chan *api.ValueList, error) {
	stream, err := c.client.QueryValues(ctx, &pb.QueryValuesRequest{
		Identifier: MarshalIdentifier(id),
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan *api.ValueList)

	go func() {
		defer close(ch)

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error while receiving value lists: %v", err)
				return
			}

			vl, err := UnmarshalValueList(res.GetValueList())
			if err != nil {
				log.Printf("received malformed response: %v", err)
				continue
			}

			ch <- vl
		}
	}()

	return ch, nil
}

func (c *client) Write(vl api.ValueList) error {
	pbVL, err := MarshalValueList(&vl)
	if err != nil {
		return err
	}

	stream, err := c.client.DispatchValues(c.ctx)
	if err != nil {
		return err
	}

	req := &pb.DispatchValuesRequest{
		ValueList: pbVL,
	}
	if err := stream.Send(req); err != nil {
		stream.CloseSend()
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}
