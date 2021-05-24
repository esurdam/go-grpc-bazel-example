package server

import (
	"context"
	"fmt"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
)

// Server implements pb.GreeterServer
type Server struct{}

// SayHello implements pb.GreeterServer
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: fmt.Sprintf("Hellos %s", req.Name)}, nil
}
