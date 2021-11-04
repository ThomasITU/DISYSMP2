package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ThomasITU/DISYSMP2/DISYSMP2/ChittyChat"
	"google.golang.org/grpc"
)

const (
	port        = ":9080"
	logFileName = "ServerLog"
)

type chittyChatServer struct {
	ChittyChat.UnimplementedChittyChatServiceServer
	clientChannels []chan *ChittyChat.Message
	arbiter        sync.Mutex
}

var (
	vectorClock      []int32
	clientsConnected map[string]int
	nextClientIndex  int
)

func main() {
	lis, err := net.Listen("tcp", "localhost"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	ChittyChat.RegisterChittyChatServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)

}

func (s *chittyChatServer) JoinChat(joinRequest *ChittyChat.JoinRequest, msgStream ChittyChat.ChittyChatService_JoinChatServer) error {

	// create and add a messageChannel for the new user, save in a slice to Broadcast messages later
	username := joinRequest.GetUsername()
	msgChannel := make(chan *ChittyChat.Message)
	s.clientChannels = append(s.clientChannels, msgChannel)
	msg := username + " joined the chat"

	s.arbiter.Lock()
	clientId := nextClientIndex
	vectorClock = append(vectorClock, 0)
	vectorClock[clientId]++
	clientsConnected[username] = clientId
	nextClientIndex++
	chittyMsg := generateChittyMsg(vectorClock, msg, username)
	Logger(msg, vectorClock, logFileName)
	s.arbiter.Unlock()

	Broadcast(chittyMsg, s)

	// runs until server is shutdown
	for {
		select {
		case <-msgStream.Context().Done():
			return nil
		case msg := <-msgChannel:
			// add logging
			fmt.Printf("GO ROUTINE (got message): %v \n", msg)
			msgStream.Send(msg)
		}
	}
}

func (s *chittyChatServer) LeaveChat(ctx context.Context, leaveChatRequest *ChittyChat.LeaveChatRequest) (*ChittyChat.LeaveChatResponse, error) {
	username := leaveChatRequest.GetUsername()
	index := clientsConnected[username]
	s.clientChannels[index] = nil
	//delete(s.clientChannels[index]) -- muligvis ændre til et map så der kan delete
	delete(clientsConnected, username)
	return &ChittyChat.LeaveChatResponse{Msg: "You succesfully left, " + username}, nil
}

func (s *chittyChatServer) Publish(msgStream ChittyChat.ChittyChatService_PublishServer) error {
	msg, err := msgStream.Recv()

	if err == io.EOF {
		return nil
	}
	checkErr(err)

	valid, err := ValidateMessage(msg.GetMsgAndVectorClock().Msg)

	if valid {
		ack := ChittyChat.PublishResponse{MsgStatus: "Hopefully Sent"}
		msgStream.SendAndClose(&ack)

		Broadcast(msg, s)
	} else {
		msg := msg.MsgAndVectorClock.GetMsg() + ", error msg " + err.Error()
		Logger(msg, vectorClock, logFileName)
		return err
	}

	return nil
}

func Broadcast(msg *ChittyChat.Message, s *chittyChatServer) {

	s.arbiter.Lock()
	CASVectorClock(vectorClock, msg)
	newMsg := generateChittyMsg(vectorClock, msg.MsgAndVectorClock.GetMsg(), msg.GetUsername())
	Logger(newMsg.MsgAndVectorClock.GetMsg(), vectorClock, logFileName)
	s.arbiter.Unlock()

	clientStreams := s.clientChannels
	for _, msgChan := range clientStreams {
		if msgChan != nil {
			msgChan <- newMsg
		}
	}

}

// helper methods

func generateChittyMsg(vectorClock []int32, msg string, username string) *ChittyChat.Message {
	broadcastMsg := ChittyChat.BroadCastMessage{Msg: msg, VectorClock: vectorClock}
	chittyMsg := ChittyChat.Message{Username: username, MsgAndVectorClock: &broadcastMsg}
	return &chittyMsg
}

func newServer() *chittyChatServer {
	s := &chittyChatServer{
		clientChannels: make([]chan *ChittyChat.Message, 10),
	}
	return s
}

func CASVectorClock(vectorClock []int32, publishRequest *ChittyChat.Message) {
	user := publishRequest.GetUsername()
	userVectorClock := publishRequest.MsgAndVectorClock.GetVectorClock()
	index := clientsConnected[user]
	vectorClock[index] = Max(vectorClock[index], userVectorClock[index]) + 1
}

func Max(x, y int32) int32 {
	if x < y {
		return y
	}
	return x
}

func ValidateMessage(message string) (bool, error) {
	valid := utf8.Valid([]byte(message))
	if !valid {
		fmt.Println(message)
		return false, errors.New("not UTF-8")
	}
	if len(message) > 128 {
		return false, errors.New("too long")
	}
	return true, nil
}

func Logger(message string, vectorClock []int32, logFileName string) {
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message + ", VectorClock: " + FormatVectorClock(vectorClock))
}

func FormatVectorClock(clock []int32) string {
	var sb = strings.Builder{}
	sb.WriteString("<")
	for i := 0; i < len(clock); i++ {
		sb.WriteString(" ")
		sb.WriteString(strconv.Itoa(int(clock[i])))
		sb.WriteString(",")
	}
	sb.WriteString(" >")
	return sb.String()
}

func checkErr(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
