package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lottejd/DISYSMP2/ChittyChat"
	"google.golang.org/grpc"
)

type bufferedMessage struct {
	message         string
	vectorTimeStamp []int
}

const (
	address = "localhost:8080"
)

var clientId int

var buffer chan (string)

func main() {
	buffer = make(chan string, 5)
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	chat := ChittyChat.NewChittyChatServiceClient(conn)

	// Contact the server and print out its response.
	ctx := context.Background()

	//go GetBroadcast(buffer, chat)

	for {
		var input string
		fmt.Scanln(&input)

		PublishFromClient(input, ctx, chat)
	}
}

func GetBroadcast(buffer chan bufferedMessage, ctx context.Context, chat ChittyChat.ChittyChatServiceClient) {

	for {
		time.Sleep(time.Second * 5)
		//chat.GetBroadCast(ctx, )
	}
}

func PublishFromClient(input string, ctx context.Context, chittyServer ChittyChat.ChittyChatServiceClient) {

	inputFromClient := ChittyChat.PublishRequest{Request: input, ClientId: int32(clientId)}

	response, err := chittyServer.Publish(ctx, &inputFromClient)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf(response.GetName())
}
