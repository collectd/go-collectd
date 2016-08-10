package rpc // import "collectd.org/rpc"

import (
	"io"

	pb "collectd.org/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// RegisterServer registers the implementation srv with the gRPC instance s.
func RegisterServer(s *grpc.Server, srv Interface) {
	pb.RegisterCollectdServer(s, &server{
		srv: srv,
	})
}

// server implements pb.CollectdServer using srv.
type server struct {
	srv Interface
}

// DispatchValues reads ValueLists from stream and calls the Write()
// implementation on each one.
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

// QueryValues calls the Query() implementation and streams all ValueLists from
// the channel back to the client.
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
