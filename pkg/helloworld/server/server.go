package server

import (
	"context"
	"fmt"

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.GreeterServer
type Server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements pb.GreeterServer
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Hello %s!", req.Name)}, nil
}
