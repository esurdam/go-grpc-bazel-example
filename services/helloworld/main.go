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

func grpcHandlerFunc(grpcServer *grpc.Server, httpServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	})
}

func applyEnvFallbacks(cert, key, ca *string) {
	if *cert == "" {
		*cert = os.Getenv("SSL_CERT_PATH")
	}
	if *key == "" {
		*key = os.Getenv("SSL_KEY_PATH")
	}
	if *ca == "" {
		*ca = os.Getenv("SSL_CA_CERT_PATH")
	}
	// In this example, the generated cert contains the CA.
	// In a production environment, the CA cert should be separate.
	if *ca == "" {
		*ca = *cert
	}
}

func loadRootCAs(caPath string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("system cert pool: %w", err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if caPath == "" {
		return rootCAs, nil
	}
	certPEMBlock, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert %q: %w", caPath, err)
	}
	if !rootCAs.AppendCertsFromPEM(certPEMBlock) {
		return nil, fmt.Errorf("append ca certs from PEM: bad certs")
	}
	return rootCAs, nil
}

func tlsServerName(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort
	}
	return host
}

// newMux builds the HTTP/gRPC multiplexed handler. TLS is terminated by
// http.Server; do not attach grpc.Creds when using ServeHTTP.
func newMux(ctx context.Context, dialAddr string, rootCAs *x509.CertPool, skipVerify bool, swagger []byte) (http.Handler, *grpc.Server, error) {
	grpcServer := grpc.NewServer(zerolog.UnaryInterceptor())
	pb.RegisterGreeterServer(grpcServer, &server.Server{})

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.Greeter_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName:         tlsServerName(dialAddr),
		RootCAs:            rootCAs,
		InsecureSkipVerify: skipVerify,
		MinVersion:         tls.VersionTLS12,
	})
	gwmux := runtime.NewServeMux()
	if err := pb.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, dialAddr, []grpc.DialOption{
		grpc.WithTransportCredentials(dcreds),
	}); err != nil {
		return nil, nil, fmt.Errorf("register gateway: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	openapi.Mount(mux, swagger, "Helloworld API")
	mux.Handle("/", gwmux)

	return grpcHandlerFunc(grpcServer, mux), grpcServer, nil
}

func main() {
	flag.Parse()
	applyEnvFallbacks(sslCert, sslKey, sslCACert)

	pair, err := tls.LoadX509KeyPair(*sslCert, *sslKey)
	if err != nil {
		log.Fatalf("unable to load ssl cert or key -- required: %v\n", err)
	}

	rootCAs, err := loadRootCAs(*sslCACert)
	if err != nil {
		log.Fatalf("unable to load ca certs: %v\n", err)
	}

	addr := fmt.Sprintf("localhost:%d", *httpPort)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler, _, err := newMux(ctx, addr, rootCAs, *insecure, Data)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", *httpPort))
	if err != nil {
		log.Fatalf("unable to listen on tcp port %d, %v\n", *httpPort, err)
	}
	gwServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
			// Advertise both so kubelet HTTPS probes (HTTP/1.1) and gRPC (h2) work.
			NextProtos: []string{"h2", "http/1.1"},
			MinVersion: tls.VersionTLS12,
		},
	}
	go func() {
		log.Printf("serving at https://%s\n", addr)
		if err := gwServer.Serve(tls.NewListener(conn, gwServer.TLSConfig)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctxClos, cancelClose := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelClose()
	cancel() // stop gateway dial context
	if err := gwServer.Shutdown(ctxClos); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exiting")
}
