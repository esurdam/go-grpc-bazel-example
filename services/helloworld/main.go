// Program main is the entrpoint to our helloworld GreeterServer
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"github.com/AdGreetz/go-grpc-bazel-example/pkg/helloworld/server"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "grpc port to listen on")
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(grpcServer, &server.Server{})
	log.Printf("Now listening for grpc on localhost:%d\n", *port)
	log.Fatal(grpcServer.Serve(lis))
}
