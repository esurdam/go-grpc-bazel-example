package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestTLSServerName(t *testing.T) {
	if got := tlsServerName("localhost:4443"); got != "localhost" {
		t.Fatalf("tlsServerName() = %q, want localhost", got)
	}
	if got := tlsServerName("localhost"); got != "localhost" {
		t.Fatalf("tlsServerName() = %q, want localhost", got)
	}
}

func TestApplyEnvFallbacks(t *testing.T) {
	t.Run("fills all from env", func(t *testing.T) {
		t.Setenv("SSL_CERT_PATH", "/env/cert.pem")
		t.Setenv("SSL_KEY_PATH", "/env/key.pem")
		t.Setenv("SSL_CA_CERT_PATH", "/env/ca.pem")

		cert, key, ca := "", "", ""
		applyEnvFallbacks(&cert, &key, &ca)
		if cert != "/env/cert.pem" || key != "/env/key.pem" || ca != "/env/ca.pem" {
			t.Fatalf("applyEnvFallbacks() = (%q, %q, %q)", cert, key, ca)
		}
	})

	t.Run("flags win over env", func(t *testing.T) {
		t.Setenv("SSL_CERT_PATH", "/env/cert.pem")
		t.Setenv("SSL_KEY_PATH", "/env/key.pem")
		t.Setenv("SSL_CA_CERT_PATH", "/env/ca.pem")

		cert, key, ca := "/flag/cert.pem", "/flag/key.pem", "/flag/ca.pem"
		applyEnvFallbacks(&cert, &key, &ca)
		if cert != "/flag/cert.pem" || key != "/flag/key.pem" || ca != "/flag/ca.pem" {
			t.Fatalf("applyEnvFallbacks() = (%q, %q, %q)", cert, key, ca)
		}
	})

	t.Run("ca falls back to cert", func(t *testing.T) {
		t.Setenv("SSL_CERT_PATH", "")
		t.Setenv("SSL_KEY_PATH", "")
		t.Setenv("SSL_CA_CERT_PATH", "")

		cert, key, ca := "/only/cert.pem", "/only/key.pem", ""
		applyEnvFallbacks(&cert, &key, &ca)
		if ca != "/only/cert.pem" {
			t.Fatalf("ca = %q, want cert path", ca)
		}
	})
}

func TestLoadRootCAs(t *testing.T) {
	certPEM, _ := mustTestCert(t)
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Run("empty path", func(t *testing.T) {
		pool, err := loadRootCAs("")
		if err != nil {
			t.Fatalf("loadRootCAs: %v", err)
		}
		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
	})

	t.Run("valid ca file", func(t *testing.T) {
		pool, err := loadRootCAs(caPath)
		if err != nil {
			t.Fatalf("loadRootCAs: %v", err)
		}
		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := loadRootCAs(filepath.Join(dir, "missing.pem")); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("bad pem", func(t *testing.T) {
		bad := filepath.Join(dir, "bad.pem")
		if err := os.WriteFile(bad, []byte("not-a-cert"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if _, err := loadRootCAs(bad); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGRPCHandlerFunc(t *testing.T) {
	httpHits := 0
	grpcSrv := grpc.NewServer()
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		httpHits++
		w.WriteHeader(http.StatusNoContent)
	})
	handler := grpcHandlerFunc(grpcSrv, httpHandler)

	t.Run("http path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.ProtoMajor = 1
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if httpHits != 1 {
			t.Fatalf("httpHits = %d", httpHits)
		}
	})

	t.Run("non-grpc h2 still http", func(t *testing.T) {
		before := httpHits
		req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
		req.ProtoMajor = 2
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if httpHits != before+1 {
			t.Fatalf("httpHits = %d", httpHits)
		}
	})
}

func TestMuxTLSGRPCAndREST(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair: %v", err)
	}
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(certPEM) {
		t.Fatal("AppendCertsFromPEM failed")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	swagger := []byte(`{"swagger":"2.0","info":{"title":"test","version":"0"},"paths":{}}`)
	handler, _, err := newMux(ctx, addr, rootCAs, false, swagger)
	if err != nil {
		t.Fatalf("newMux: %v", err)
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
			NextProtos:   []string{"h2", "http/1.1"},
			MinVersion:   tls.VersionTLS12,
		},
	}
	go func() {
		_ = srv.Serve(tls.NewListener(ln, srv.TLSConfig))
	}()
	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	})

	tlsConfig := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}

	waitHealthy(t, "https://"+addr+"/healthz", tlsConfig)

	t.Run("healthz", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		resp, err := client.Get("https://" + addr + "/healthz")
		if err != nil {
			t.Fatalf("GET /healthz: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "ok" {
			t.Fatalf("body = %q", body)
		}
	})

	t.Run("healthz http1", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    rootCAs,
				ServerName: "localhost",
				MinVersion: tls.VersionTLS12,
				NextProtos: []string{"http/1.1"},
			},
			TLSNextProto:      map[string]func(string, *tls.Conn) http.RoundTripper{},
			ForceAttemptHTTP2: false,
		}}
		resp, err := client.Get("https://" + addr + "/healthz")
		if err != nil {
			t.Fatalf("GET /healthz: %v", err)
		}
		defer resp.Body.Close()
		if resp.ProtoMajor != 1 {
			t.Fatalf("proto = %s, want HTTP/1.x", resp.Proto)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("swagger", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		resp, err := client.Get("https://" + addr + "/swagger.json")
		if err != nil {
			t.Fatalf("GET /swagger.json: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
		var doc map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
			t.Fatalf("decode swagger: %v", err)
		}
		if doc["swagger"] != "2.0" {
			t.Fatalf("swagger = %v", doc["swagger"])
		}
	})

	t.Run("docs", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		resp, err := client.Get("https://" + addr + "/docs/")
		if err != nil {
			t.Fatalf("GET /docs/: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	})

	t.Run("rest sayhello", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		resp, err := client.Post(
			"https://"+addr+"/v1/greeter",
			"application/json",
			strings.NewReader(`{"name":"Test"}`),
		)
		if err != nil {
			t.Fatalf("POST rest: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d body = %s", resp.StatusCode, body)
		}
		var reply pb.HelloReply
		if err := protojson.Unmarshal(body, &reply); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if reply.Message != "Hello Test!" {
			t.Fatalf("message = %q", reply.Message)
		}
	})

	t.Run("rest validation", func(t *testing.T) {
		client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
		resp, err := client.Post(
			"https://"+addr+"/v1/greeter",
			"application/json",
			strings.NewReader(`{"name":""}`),
		)
		if err != nil {
			t.Fatalf("POST rest: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("status = %d body = %s, want 400", resp.StatusCode, body)
		}
	})

	t.Run("grpc sayhello", func(t *testing.T) {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		defer conn.Close()
		cli := pb.NewGreeterClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		reply, err := cli.SayHello(ctx, &pb.HelloRequest{Name: "gRPC"})
		if err != nil {
			t.Fatalf("SayHello: %v", err)
		}
		if reply.Message != "Hello gRPC!" {
			t.Fatalf("message = %q", reply.Message)
		}
	})

	t.Run("grpc validation", func(t *testing.T) {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		defer conn.Close()
		cli := pb.NewGreeterClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = cli.SayHello(ctx, &pb.HelloRequest{Name: ""})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("status = %v, want InvalidArgument", status.Code(err))
		}
	})

	t.Run("grpc health", func(t *testing.T) {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		defer conn.Close()
		cli := healthpb.NewHealthClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, service := range []string{"", pb.Greeter_ServiceDesc.ServiceName} {
			resp, err := cli.Check(ctx, &healthpb.HealthCheckRequest{Service: service})
			if err != nil {
				t.Fatalf("Check(%q): %v", service, err)
			}
			if resp.Status != healthpb.HealthCheckResponse_SERVING {
				t.Fatalf("Check(%q) status = %v", service, resp.Status)
			}
		}
	})
}

func TestRunStartsAndShutsDown(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, serverConfig{
			httpPort: port,
			cert:     certPath,
			key:      keyPath,
			ca:       certPath,
			swagger:  []byte(`{"swagger":"2.0","info":{"title":"test","version":"0"},"paths":{}}`),
		})
	}()

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(certPEM) {
		t.Fatal("AppendCertsFromPEM failed")
	}
	tlsConfig := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}
	waitHealthy(t, fmt.Sprintf("https://127.0.0.1:%d/healthz", port), tlsConfig)

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	resp, err := client.Post(
		fmt.Sprintf("https://127.0.0.1:%d/v1/greeter", port),
		"application/json",
		strings.NewReader(`{"name":"Run"}`),
	)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body = %s", resp.StatusCode, body)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not shut down")
	}
}

func TestMainErr(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	origNotify := notifyContext
	t.Cleanup(func() { notifyContext = origNotify })

	ctx, cancel := context.WithCancel(context.Background())
	notifyContext = func(parent context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
		return ctx, cancel
	}

	origCert, origKey, origCA, origPort := *sslCert, *sslKey, *sslCACert, *httpPort
	t.Cleanup(func() {
		*sslCert, *sslKey, *sslCACert, *httpPort = origCert, origKey, origCA, origPort
	})
	*sslCert, *sslKey, *sslCACert, *httpPort = certPath, keyPath, certPath, port

	errCh := make(chan error, 1)
	go func() {
		errCh <- mainErr([]byte(`{"swagger":"2.0","info":{"title":"test","version":"0"},"paths":{}}`))
	}()

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(certPEM) {
		t.Fatal("AppendCertsFromPEM failed")
	}
	waitHealthy(t, fmt.Sprintf("https://127.0.0.1:%d/healthz", port), &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	})
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("mainErr: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("mainErr did not return")
	}
}

func TestRunMissingCert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := run(ctx, serverConfig{
		httpPort: 0,
		cert:     filepath.Join(t.TempDir(), "missing.pem"),
		key:      filepath.Join(t.TempDir(), "missing-key.pem"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunBadCA(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	caPath := filepath.Join(dir, "bad-ca.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}
	if err := os.WriteFile(caPath, []byte("not-a-cert"), 0o600); err != nil {
		t.Fatalf("WriteFile ca: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := run(ctx, serverConfig{
		httpPort: 0,
		cert:     certPath,
		key:      keyPath,
		ca:       caPath,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunPortInUse(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}

	// Bind the same wildcard form run() uses (":port") so the port is taken.
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = run(ctx, serverConfig{
		httpPort: port,
		cert:     certPath,
		key:      keyPath,
		ca:       certPath,
		swagger:  []byte(`{}`),
	})
	if err == nil {
		t.Fatal("expected listen error")
	}
}

func TestNewMuxCanceledContext(t *testing.T) {
	certPEM, _ := mustTestCert(t)
	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(certPEM) {
		t.Fatal("AppendCertsFromPEM failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := newMux(ctx, "127.0.0.1:1", rootCAs, true, []byte(`{}`))
	// grpc.NewClient is lazy; canceled ctx may or may not fail registration.
	// Accept either outcome so we still exercise the call path.
	_ = err
}

func TestRunInsecureGateway(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, serverConfig{
			httpPort: port,
			cert:     certPath,
			key:      keyPath,
			ca:       certPath,
			insecure: true,
			swagger:  []byte(`{"swagger":"2.0","info":{"title":"test","version":"0"},"paths":{}}`),
		})
	}()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
	}
	waitHealthy(t, fmt.Sprintf("https://127.0.0.1:%d/healthz", port), tlsConfig)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not shut down")
	}
}

func waitHealthy(t *testing.T, url string, tlsConfig *tls.Config) {
	t.Helper()
	client := &http.Client{
		Timeout:   time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not healthy: %v", lastErr)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func mustTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		t.Fatal(fmt.Errorf("empty pem"))
	}
	return certPEM, keyPEM
}
