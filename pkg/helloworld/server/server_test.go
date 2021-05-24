package server_test

import (
	"context"
	"reflect"
	"testing"

	pb "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld"
	"github.com/AdGreetz/go-grpc-bazel-example/pkg/helloworld/server"
)

func TestServer_SayHello(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.HelloRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.HelloReply
		wantErr bool
	}{
		{
			name: "TestServer_SayHello",
			args: args{
				ctx: context.Background(),
				req: &pb.HelloRequest{
					Name: "TestName",
				},
			},
			want: &pb.HelloReply{
				Message: "Hello TestName!",
			},
			wantErr: false,
		},
		{
			name: "TestServer_SayHelloErr",
			args: args{
				ctx: context.Background(),
				req: &pb.HelloRequest{
					Name: "",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server.Server{}
			got, err := s.SayHello(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SayHello() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SayHello() got = %v, want %v", got, tt.want)
			}
		})
	}
}
