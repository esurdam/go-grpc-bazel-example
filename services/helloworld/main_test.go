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
	"strings"
	"testing"
	"time"

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
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
	t.Setenv("SSL_CERT_PATH", "/env/cert.pem")
	t.Setenv("SSL_KEY_PATH", "/env/key.pem")
	t.Setenv("SSL_CA_CERT_PATH", "/env/ca.pem")

	cert, key, ca := "", "", ""
	applyEnvFallbacks(&cert, &key, &ca)
	if cert != "/env/cert.pem" || key != "/env/key.pem" || ca != "/env/ca.pem" {
		t.Fatalf("applyEnvFallbacks() = (%q, %q, %q)", cert, key, ca)
	}
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
		cancel()
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
}

func waitHealthy(t *testing.T, url string, tlsConfig *tls.Config) {
	t.Helper()
	client := &http.Client{
		Timeout:   time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not healthy: %v", err)
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
