package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

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

	go func() {
		for {
			time.Sleep(1 * time.Second)

			conn, err := lis.Accept()

			check(err, "Accepted connection")
			createClientConnection(conn)
		}
	}()

	ChittyChat.RegisterChittyChatServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
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
		if _, err := conn.Write([]byte("Hello, " + name)); err != nil {
			logger("hello err, "+name, name+"err")
		}
		logger("hello, "+name, name)
	}
}

func (s *Server) GetBroadcast(ctx context.Context) (*ChittyChat.Response, error) {
	latestBroadcast := GetLatestLocalBroadcast()
	return &ChittyChat.Response{response: latestBroadcast, ClientsConnected: latestBroadcast.}, nil
}

func (s *Server) Publish(ctx context.Context, message *ChittyChat.PublishRequest) (*ChittyChat.Response, error) {
	validateMessage, err := ValidateMessage(message.GetRequest())
	if validateMessage {
		Broadcast(message.GetRequest(), int(message.GetClientId()))
		logger("Request was accepted, and is being broadcasted, "+strconv.Itoa(int(message.GetClientId())), "ServerLogFile")
		return &ChittyChat.Response{Name: "Request was accepted, and is being broadcasted"}, nil
	} else {
		logger(err.Error(), "errorLog")
		return &ChittyChat.Response{Name: err.Error()}, err
	}
}

// helper method
func GetLatestLocalBroadcast([]int ){


}
func logger(message string, logFileName string) {
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message)
}

func ValidateMessage(message string) (bool, error) {
	valid := utf8.Valid([]byte(message))
	if valid == false {
		return false, errors.New("not UTF-8")
	}
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

func Broadcast(message string) {
	// not implemented yet
}
