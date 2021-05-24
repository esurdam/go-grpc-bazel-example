package server

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
)

// Server implements pb.GreeterServer
type Server struct{}

// SayHello implements pb.GreeterServer
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Hello %s!", req.Name)}, nil
}
