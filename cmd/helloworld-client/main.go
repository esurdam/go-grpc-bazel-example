package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"time"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	var serverAddr = flag.String("server-addr", "localhost:443", "grpc server address")
	var name = flag.String("name", "", "name to greet")
	var sslCACert = flag.String("ca-cert", "", "path to tls ca cert")
	var insecure = flag.Bool("insecure", false, "enable insecure tls skip verify")
	flag.Parse()

	// add the cert as CA
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if *sslCACert != "" {
		certPEMBlock, _ := os.ReadFile(*sslCACert)
		ok := rootCAs.AppendCertsFromPEM([]byte(certPEMBlock))
		if !ok {
			log.Fatal("unable to append certs from PEM: bad certs")
		}
	}

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         *serverAddr,
		RootCAs:            rootCAs,
		InsecureSkipVerify: *insecure,
	})
	dopts := []grpc.DialOption{
		grpc.WithTransportCredentials(dcreds),
	}
	conn, err := grpc.Dial(*serverAddr, dopts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cli := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reply, err := cli.SayHello(ctx, &pb.HelloRequest{
		Name: *name,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(reply)
}
