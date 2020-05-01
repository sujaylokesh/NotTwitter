// main.go

package main

import (
	"log"
	"net"
	"time"

	pb "github.com/fk652/import/commonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	port = ":50051"
)

type server struct {
	pb.UnimplementedBackendServer
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)
	pb.RegisterBackendServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
