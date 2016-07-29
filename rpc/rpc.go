package rpc // import "collectd.org/rpc"

import (
	"fmt"

	"collectd.org/api"
	pb "collectd.org/rpc/proto"
	"collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
)

type Server interface {
	api.Writer
	Query(api.Identifier) ([]*api.ValueList, error)
}

type wrapper struct {
	srv Server
}

func Wrap(s Server) pb.CollectdServer {
	return &wrapper{
		srv: s,
	}
}

func marshalValue(v api.Value) (*types.Value, error) {
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

func unmarshalValue(in *types.Value) (api.Value, error) {
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

func marshalIdentifier(id api.Identifier) *types.Identifier {
	return &types.Identifier{
		Host:           id.Host,
		Plugin:         id.Plugin,
		PluginInstance: id.PluginInstance,
		Type:           id.Type,
		TypeInstance:   id.TypeInstance,
	}
}

func unmarshalIdentifier(in *types.Identifier) api.Identifier {
	return api.Identifier{
		Host:           in.Host,
		Plugin:         in.Plugin,
		PluginInstance: in.PluginInstance,
		Type:           in.Type,
		TypeInstance:   in.TypeInstance,
	}
}

func marshalValueList(vl *api.ValueList) (*types.ValueList, error) {
	t, err := ptypes.TimestampProto(vl.Time)
	if err != nil {
		return nil, err
	}

	var pbValues []*types.Value
	for _, v := range vl.Values {
		pbValue, err := marshalValue(v)
		if err != nil {
			return nil, err
		}

		pbValues = append(pbValues, pbValue)
	}

	return &types.ValueList{
		Value:      pbValues,
		Time:       t,
		Interval:   ptypes.DurationProto(vl.Interval),
		Identifier: marshalIdentifier(vl.Identifier),
	}, nil
}

func unmarshalValueList(in *types.ValueList) (*api.ValueList, error) {
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
		v, err := unmarshalValue(pbValue)
		if err != nil {
			return nil, err
		}

		values = append(values, v)
	}

	return &api.ValueList{
		Identifier: unmarshalIdentifier(in.GetIdentifier()),
		Time:       t,
		Interval:   interval,
		Values:     values,
		// TODO(octo): DSNames
	}, nil
}

func (wrap *wrapper) DispatchValues(_ context.Context, req *pb.DispatchValuesRequest) (*pb.DispatchValuesReply, error) {
	vl, err := unmarshalValueList(req.GetValues())
	if err != nil {
		return nil, err
	}

	if err := wrap.srv.Write(*vl); err != nil {
		return nil, err
	}

	return &pb.DispatchValuesReply{}, nil
}

func (wrap *wrapper) QueryValues(_ context.Context, req *pb.QueryValuesRequest) (*pb.QueryValuesReply, error) {
	id := unmarshalIdentifier(req.GetIdentifier())

	vls, err := wrap.srv.Query(id)
	if err != nil {
		return nil, err
	}

	var valuesPb []*types.ValueList
	for _, vl := range vls {
		pb, err := marshalValueList(vl)
		if err != nil {
			return nil, err
		}
		valuesPb = append(valuesPb, pb)
	}

	return &pb.QueryValuesReply{
		Values: valuesPb,
	}, nil
}
