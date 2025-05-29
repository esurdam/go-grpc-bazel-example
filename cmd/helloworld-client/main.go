package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"time"

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	serverAddr := flag.String("server-addr", "localhost:443", "grpc server address")
	name := flag.String("name", "", "name to greet")
	sslCACert := flag.String("ca-cert", "", "path to tls ca cert")
	insecure := flag.Bool("insecure", false, "enable insecure tls skip verify")
	flag.Parse()

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("failed to get system cert pool: %v", err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if *sslCACert != "" {
		certPEMBlock, err := os.ReadFile(*sslCACert)
		if err != nil {
			log.Fatalf("failed to read CA cert file: %v", err)
		}
		if !rootCAs.AppendCertsFromPEM(certPEMBlock) {
			log.Fatal("unable to append certs from PEM: bad certs")
		}
	}

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         *serverAddr,
		RootCAs:            rootCAs,
		InsecureSkipVerify: *insecure,
	})
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(dcreds))
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := cli.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("failed to say hello: %v", err)
	}
	log.Printf("Received reply: %v", reply)
}
