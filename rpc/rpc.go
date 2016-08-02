package rpc // import "collectd.org/rpc"

import (
	"io"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// CollectdServer is an idiomatic Go interface for proto.CollectdServer Use
// RegisterCollectdServer() to hook an object, which implements this interface,
// up to the gRPC server.
type CollectdServer interface {
	Query(*api.Identifier) (<-chan *api.ValueList, error)
}

// RegisterCollectdServer registers the implementation srv with the gRPC instance s.
func RegisterCollectdServer(s *grpc.Server, srv CollectdServer) {
	pb.RegisterCollectdServer(s, &collectdWrapper{
		srv: srv,
	})
}

type DispatchServer interface {
	api.Writer
}

func RegisterDispatchServer(s *grpc.Server, srv DispatchServer) {
	pb.RegisterDispatchServer(s, &dispatchWrapper{
		srv: srv,
	})
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

// dispatchWrapper implements pb.DispatchServer using srv.
type dispatchWrapper struct {
	srv DispatchServer
}

func (wrap *dispatchWrapper) DispatchValues(stream pb.Dispatch_DispatchValuesServer) error {
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

		if err := wrap.srv.Write(*vl); err != nil {
			return grpc.Errorf(codes.Internal, "Write(%v): %v", vl, err)
		}
	}

	return stream.SendAndClose(&pb.DispatchValuesResponse{})
}

// collectdWrapper implements pb.CollectdServer using srv.
type collectdWrapper struct {
	srv CollectdServer
}

func (wrap *collectdWrapper) QueryValues(req *pb.QueryValuesRequest, stream pb.Collectd_QueryValuesServer) error {
	id := UnmarshalIdentifier(req.GetIdentifier())

	ch, err := wrap.srv.Query(id)
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
