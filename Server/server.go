package main

import (
	"log"
	"net"

	chat "github.com/lottejd/DISYSMP2/ChittyChat"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
)

type Server struct {
	chat.UnimplementedChittChatServiceServer
}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	chat.RegisterChittChatServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
	}
}
