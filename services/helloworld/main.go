// Program main is the entrypoint to our helloworld GreeterServer.
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

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"github.com/esurdam/go-grpc-bazel-example/pkg/helloworld/server"
	"github.com/esurdam/go-grpc-bazel-example/pkg/openapi"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	zerolog "github.com/philip-bui/grpc-zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	httpPort  = flag.Int("http-port", 443, "http port to listen on (serve json API)")
	sslCert   = flag.String("cert", "", "path to tls cert")
	sslKey    = flag.String("key", "", "path to tls key")
	sslCACert = flag.String("ca-cert", "", "path to tls ca cert")
	insecure  = flag.Bool("insecure", false, "enable insecure tls skip verify for grpc-gateway")
)

// grpcHandlerFunc routes HTTP/2 gRPC to the gRPC server and everything else
// to the HTTP handler (gateway, healthz, docs).
func grpcHandlerFunc(grpcServer *grpc.Server, other http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}
		other.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	if *sslCert == "" {
		*sslCert = os.Getenv("SSL_CERT_PATH")
	}
	if *sslKey == "" {
		*sslKey = os.Getenv("SSL_KEY_PATH")
	}
	if *sslCACert == "" {
		*sslCACert = os.Getenv("SSL_CA_CERT_PATH")
	}
	// In this example the generated cert contains the CA; production should
	// keep the CA separate.
	if *sslCACert == "" {
		*sslCACert = *sslCert
	}

	pair, err := tls.LoadX509KeyPair(*sslCert, *sslKey)
	if err != nil {
		log.Fatalf("unable to load ssl cert or key: %v", err)
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil || rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if *sslCACert != "" {
		pemBlock, err := os.ReadFile(*sslCACert)
		if err != nil {
			log.Fatalf("unable to read ca cert: %v", err)
		}
		if !rootCAs.AppendCertsFromPEM(pemBlock) {
			log.Fatal("unable to append ca certs from PEM")
		}
	}

	addr := fmt.Sprintf("localhost:%d", *httpPort)

	// TLS is terminated by http.Server; do not attach grpc.Creds when using ServeHTTP.
	grpcServer := grpc.NewServer(zerolog.UnaryInterceptor())
	pb.RegisterGreeterServer(grpcServer, &server.Server{})

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.Greeter_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	// Gateway dial lifecycle is independent of the shutdown signal so
	// http.Server.Shutdown can drain in-flight REST requests first.
	gwCtx, gwCancel := context.WithCancel(context.Background())
	defer gwCancel()

	gwmux := runtime.NewServeMux()
	if err := pb.RegisterGreeterHandlerFromEndpoint(gwCtx, gwmux, addr, []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:         "localhost",
			RootCAs:            rootCAs,
			InsecureSkipVerify: *insecure,
			MinVersion:         tls.VersionTLS12,
		})),
	}); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	openapi.Mount(mux, Data, "Helloworld API")
	mux.Handle("/", gwmux)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		log.Fatalf("unable to listen on tcp port %d: %v", *httpPort, err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           grpcHandlerFunc(grpcServer, mux),
		ReadHeaderTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
			// Advertise both so kubelet HTTPS probes (HTTP/1.1) and gRPC (h2) work.
			NextProtos: []string{"h2", "http/1.1"},
			MinVersion: tls.VersionTLS12,
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("serving at https://%s", addr)
		if err := srv.Serve(tls.NewListener(ln, srv.TLSConfig)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}
	log.Println("server exiting")
}
