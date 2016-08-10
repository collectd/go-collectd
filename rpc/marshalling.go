package rpc // import "collectd.org/rpc"

import (
	"collectd.org/api"
	"collectd.org/rpc/proto/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

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
