package rpc // import "collectd.org/rpc"

import (
	"fmt"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Server is an idiomatic Go interface for the CollectdServer in
// collectd.org/rpc/proto. Use RegisterServer() to hook an object, which
// implements this interface, up to the gRPC server.
type Server interface {
	api.Writer
	Query(api.Identifier) ([]*api.ValueList, error)
}

// RegisterServer registers the implementation srv with the gRPC instance s.
func RegisterServer(s *grpc.Server, srv Server) {
	pb.RegisterCollectdServer(s, &wrapper{
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
		return nil, fmt.Errorf("%T values are not supported", v)
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
		return nil, fmt.Errorf("%T values are not supported", pbValue)
	}
}

func MarshalIdentifier(id api.Identifier) *types.Identifier {
	return &types.Identifier{
		Host:           id.Host,
		Plugin:         id.Plugin,
		PluginInstance: id.PluginInstance,
		Type:           id.Type,
		TypeInstance:   id.TypeInstance,
	}
}

func UnmarshalIdentifier(in *types.Identifier) api.Identifier {
	return api.Identifier{
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
		Value:      pbValues,
		Time:       t,
		Interval:   ptypes.DurationProto(vl.Interval),
		Identifier: MarshalIdentifier(vl.Identifier),
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
	for _, pbValue := range in.GetValue() {
		v, err := UnmarshalValue(pbValue)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return &api.ValueList{
		Identifier: UnmarshalIdentifier(in.GetIdentifier()),
		Time:       t,
		Interval:   interval,
		Values:     values,
		// TODO(octo): DSNames
	}, nil
}

// wrapper implements pb.CollectdServer using srv.
type wrapper struct {
	srv Server
}

func (wrap *wrapper) DispatchValues(_ context.Context, req *pb.DispatchValuesRequest) (*pb.DispatchValuesReply, error) {
	vl, err := UnmarshalValueList(req.GetValues())
	if err != nil {
		return nil, err
	}

	if err := wrap.srv.Write(*vl); err != nil {
		return nil, err
	}

	return &pb.DispatchValuesReply{}, nil
}

func (wrap *wrapper) QueryValues(_ context.Context, req *pb.QueryValuesRequest) (*pb.QueryValuesReply, error) {
	id := UnmarshalIdentifier(req.GetIdentifier())

	vls, err := wrap.srv.Query(id)
	if err != nil {
		return nil, err
	}

	var valuesPb []*types.ValueList
	for _, vl := range vls {
		pb, err := MarshalValueList(vl)
		if err != nil {
			return nil, err
		}
		valuesPb = append(valuesPb, pb)
	}

	return &pb.QueryValuesReply{
		Values: valuesPb,
	}, nil
}
