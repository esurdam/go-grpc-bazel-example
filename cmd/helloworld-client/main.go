package main

import (
	"context"
	"flag"
	"log"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc"
)

type Cache[V comparable] struct {
}

func main() {
	var serverAddr = flag.String("server-addr", ":9000", "grpc server address")
	var name = flag.String("name", "", "name to greet")
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)
	reply, err := cli.SayHello(context.Background(), &pb.HelloRequest{
		Name: *name,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(reply)
}
