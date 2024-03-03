// Program main is the entrypoint to our helloworld GreeterServer
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"github.com/AdGreetz/go-grpc-bazel-example/pkg/helloworld/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	zerolog "github.com/philip-bui/grpc-zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	httpPort  = flag.Int("http-port", 443, "http port to listen on (serve json API)")
	sslCert   = flag.String("cert", "", "path to tls cert")
	sslKey    = flag.String("key", "", "path to tls key")
	sslCACert = flag.String("ca-cert", "", "path to tls ca cert")
	insecure  = flag.Bool("insecure", false, "enable insecure tls skip verify for grpc-gateway")
)

func grpcHandlerFunc(grpcServer *grpc.Server, httpServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	})
}

func main() {
	flag.Parse()

	// Source cert info in production
	switch {
	case *sslCert == "":
		*sslCert = os.Getenv("SSL_CERT_PATH")
	case *sslKey == "":
		*sslKey = os.Getenv("SSL_KEY_PATH")
	case *sslCACert == "":
		*sslCACert = os.Getenv("SSL_CA_CERT_PATH")
	}
	// In this example, the generated cert contains the CA.
	// In a production environment, the CA Cert should be seperate
	if *sslCACert == "" {
		*sslCACert = *sslCert
	}

	pair, err := tls.LoadX509KeyPair(*sslCert, *sslKey)
	if err != nil {
		log.Fatalf("unable to load ssl cert or key -- required: %v\n", err)
	}

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
	addr := fmt.Sprintf("localhost:%d", *httpPort)

	opts := []grpc.ServerOption{
		zerolog.UnaryInterceptor(),
		grpc.Creds(credentials.NewClientTLSFromCert(rootCAs, addr)),
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(grpcServer, &server.Server{})

	// Register Greeter
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         addr,
		RootCAs:            rootCAs,
		InsecureSkipVerify: *insecure,
	})
	dopts := []grpc.DialOption{
		grpc.WithTransportCredentials(dcreds),
	}
	gwmux := runtime.NewServeMux()
	if err := pb.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, addr, dopts); err != nil {
		log.Fatalln("failed to register gateway:", err)
	}

	// Handle swagger
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(Data)
	})
	mux.Handle("/", gwmux)

	// Handle server
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		log.Fatalf("unable to listen on tcp port %d, %v\n", *httpPort, err)
	}
	gwServer := &http.Server{
		Addr:    addr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
			NextProtos:   []string{"h2"},
		},
	}
	go func() {
		log.Printf("serving at https://%s\n", addr)
		if err := gwServer.Serve(tls.NewListener(conn, gwServer.TLSConfig)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctxClos, cancelClose := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelClose()
	if err := gwServer.Shutdown(ctxClos); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exiting")
}
