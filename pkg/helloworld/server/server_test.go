package server_test

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/esurdam/go-grpc-bazel-example/pb/helloworld"
	"github.com/esurdam/go-grpc-bazel-example/pkg/helloworld/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_SayHello(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.HelloRequest
	}
	tests := []struct {
		name     string
		args     args
		want     *pb.HelloReply
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "greets by name",
			args: args{
				ctx: context.Background(),
				req: &pb.HelloRequest{Name: "TestName"},
			},
			want: &pb.HelloReply{Message: "Hello TestName!"},
		},
		{
			name: "greets user",
			args: args{
				ctx: context.Background(),
				req: &pb.HelloRequest{Name: "user"},
			},
			want: &pb.HelloReply{Message: "Hello user!"},
		},
		{
			name: "empty name is invalid argument",
			args: args{
				ctx: context.Background(),
				req: &pb.HelloRequest{Name: ""},
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server.Server{}
			got, err := s.SayHello(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SayHello() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if status.Code(err) != tt.wantCode {
					t.Fatalf("SayHello() status = %v, want %v", status.Code(err), tt.wantCode)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SayHello() got = %v, want %v", got, tt.want)
			}
		})
	}
}
