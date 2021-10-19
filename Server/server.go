package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"

	"github.com/lottejd/DISYSMP2/ChittyChat"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
)

var clientsConnected []int
var buffer chan (bufferedMessage)

type Server struct {
	ChittyChat.UnimplementedChittyChatServiceServer
}

type bufferedMessage struct {
	message         string
	vectorTimeStamp []int
}

func main() {
	clientsConnected = make([]int, 0, 0)
	buffer = make(chan bufferedMessage, 10)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	ChittyChat.RegisterChittyChatServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
	}

	for {
		conn, err := lis.Accept()
		check(err, "Accepted connection")
		go createClientConnection(conn)
	}
}

func createClientConnection(conn net.Conn) {
	clientBuf := bufio.NewReader(conn)
	for {
		name, err := clientBuf.ReadString('\n')
		if err != nil {
			log.Printf("Client disconnected.\n")
			break
		}
		conn.Write([]byte("Hello, " + name))
	}
}

// func (s *Server) GetBroadcast(ctx context.Context, message *ChittyChat.BroadcastRequest) (*ChittyChat.Response, error) {
// 	not implemented
// }

func (s *Server) Publish(ctx context.Context, message *ChittyChat.PublishRequest) (*ChittyChat.Response, error) {
	validateMessage, err := ValidateMessage(message.GetRequest())
	if validateMessage {
		Broadcast(message.GetRequest(), int(message.GetClientId()))
		return &ChittyChat.Response{Name: "Request was accepted, and is being broadcasted"}, nil
	} else {
		log.Fatalf(err.Error())
		return &ChittyChat.Response{Name: err.Error()}, err
	}

}

// helper method
func ValidateMessage(message string) (bool, error) {
	if len(message) > 128 {
		return false, errors.New("too long")
	}
	return true, nil
}

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	log.Printf("%s\n", message)
}

func Broadcast(message string, clientId int) {
	// not implemented yet
}
