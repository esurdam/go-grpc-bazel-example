// Program main is the entrypoint to our helloworld GreeterServer
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"github.com/AdGreetz/go-grpc-bazel-example/pkg/helloworld/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	zerolog "github.com/philip-bui/grpc-zerolog"
	"google.golang.org/grpc"
)

var (
	port     = flag.Int("port", 1000, "grpc port to listen on")
	httpPort = flag.Int("http-port", 80, "http port to listen on (serve json API)")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	opts = append(opts, zerolog.UnaryInterceptor())
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(grpcServer, &server.Server{})
	log.Printf("Now listening for grpc on localhost:%d\n", *port)

	go func() {
		log.Fatal(grpcServer.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("localhost:%d", *port),
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = pb.RegisterGreeterHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(Data)
	})
	mux.Handle("/", gwmux)
	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: mux,
	}

	log.Printf("Serving http at http://0.0.0.0:%d", *httpPort)
	log.Fatalln(gwServer.ListenAndServe())
}
