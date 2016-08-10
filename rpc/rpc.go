package rpc // import "collectd.org/rpc"

import (
	"io"
	"log"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// RegisterServer registers the implementation srv with the gRPC instance s.
func RegisterServer(s *grpc.Server, srv Interface) {
	pb.RegisterCollectdServer(s, &server{
		srv: srv,
	})
}

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

func MarshalValue(v api.Value) (*types.Value, error) {
	switch v := v.(type) {
	case api.Counter:
		return &types.Value{
			Value: &types.Value_Counter{Counter: uint64(v)},
		}, nil
	case api.Derive:
		return &types.Value{
			Value: &types.Value_Derive{Derive: int64(v)},
		}, nil
	case api.Gauge:
		return &types.Value{
			Value: &types.Value_Gauge{Gauge: float64(v)},
		}, nil
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "%T values are not supported", v)
	}
}

func UnmarshalValue(in *types.Value) (api.Value, error) {
	switch pbValue := in.GetValue().(type) {
	case *types.Value_Counter:
		return api.Counter(pbValue.Counter), nil
	case *types.Value_Derive:
		return api.Derive(pbValue.Derive), nil
	case *types.Value_Gauge:
		return api.Gauge(pbValue.Gauge), nil
	default:
		return nil, grpc.Errorf(codes.Internal, "%T values are not supported", pbValue)
	}
}

func MarshalIdentifier(id *api.Identifier) *types.Identifier {
	return &types.Identifier{
		Host:           id.Host,
		Plugin:         id.Plugin,
		PluginInstance: id.PluginInstance,
		Type:           id.Type,
		TypeInstance:   id.TypeInstance,
	}
}

func UnmarshalIdentifier(in *types.Identifier) *api.Identifier {
	return &api.Identifier{
		Host:           in.Host,
		Plugin:         in.Plugin,
		PluginInstance: in.PluginInstance,
		Type:           in.Type,
		TypeInstance:   in.TypeInstance,
	}
}

func MarshalValueList(vl *api.ValueList) (*types.ValueList, error) {
	t, err := ptypes.TimestampProto(vl.Time)
	if err != nil {
		return nil, err
	}

	var pbValues []*types.Value
	for _, v := range vl.Values {
		pbValue, err := MarshalValue(v)
		if err != nil {
			return nil, err
		}

		pbValues = append(pbValues, pbValue)
	}

	return &types.ValueList{
		Values:     pbValues,
		Time:       t,
		Interval:   ptypes.DurationProto(vl.Interval),
		Identifier: MarshalIdentifier(&vl.Identifier),
	}, nil
}

func UnmarshalValueList(in *types.ValueList) (*api.ValueList, error) {
	t, err := ptypes.Timestamp(in.GetTime())
	if err != nil {
		return nil, err
	}

	interval, err := ptypes.Duration(in.GetInterval())
	if err != nil {
		return nil, err
	}

	var values []api.Value
	for _, pbValue := range in.GetValues() {
		v, err := UnmarshalValue(pbValue)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return &api.ValueList{
		Identifier: *UnmarshalIdentifier(in.GetIdentifier()),
		Time:       t,
		Interval:   interval,
		Values:     values,
		DSNames:    in.DsNames,
	}, nil
}

// server implements pb.CollectdServer using srv.
type server struct {
	srv Interface
}

func (wrap *server) DispatchValues(stream pb.Collectd_DispatchValuesServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		vl, err := UnmarshalValueList(req.GetValueList())
		if err != nil {
			return err
		}

		// TODO(octo): pass stream.Context() to srv.Write() once the interface allows that.
		if err := wrap.srv.Write(*vl); err != nil {
			return grpc.Errorf(codes.Internal, "Write(%v): %v", vl, err)
		}
	}

	return stream.SendAndClose(&pb.DispatchValuesResponse{})
}

func (wrap *server) QueryValues(req *pb.QueryValuesRequest, stream pb.Collectd_QueryValuesServer) error {
	id := UnmarshalIdentifier(req.GetIdentifier())

	ch, err := wrap.srv.Query(stream.Context(), id)
	if err != nil {
		return grpc.Errorf(codes.Internal, "Query(%v): %v", id, err)
	}

	for vl := range ch {
		pbVL, err := MarshalValueList(vl)
		if err != nil {
			return err
		}

		res := &pb.QueryValuesResponse{
			ValueList: pbVL,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}
