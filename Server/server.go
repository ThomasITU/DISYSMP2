package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/lottejd/DISYSMP2/ChittyChat"
	"google.golang.org/grpc"
)

const (
	port          = ":8080"
	serverLogFile = "serverLog"
)

var (
	clientsConnectedVectorClocks []int32
	broadCastBuffer              chan (bufferedMessage)
	clientCount                  int
	lock                         sync.Mutex
)

type Server struct {
	ChittyChat.UnimplementedChittyChatServiceServer
}

type bufferedMessage struct {
	message         string
	vectorTimeStamp []int32
	clientId        int32
}

func main() {

	//init
	clientsConnectedVectorClocks = make([]int32, 0, 0)
	broadCastBuffer = make(chan bufferedMessage, 10)
	lock = sync.Mutex{}

	//setup listener on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	go EvalLatestBroadCast(broadCastBuffer)

	ChittyChat.RegisterChittyChatServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
	}

}

func (s *Server) GetBroadcast(ctx context.Context, _ *ChittyChat.GetBroadcastRequest) (*ChittyChat.Response, error) {
	latestBroadcast := <-broadCastBuffer
	if len(latestBroadcast.vectorTimeStamp) < 0 {
		return nil, errors.New("no recent broadcasts")
	} else {
		broadCastBuffer <- latestBroadcast
	}
	return &ChittyChat.Response{Msg: latestBroadcast.message, ClientsConnected: latestBroadcast.vectorTimeStamp, ClientId: latestBroadcast.clientId}, nil
}

func (s *Server) Publish(ctx context.Context, message *ChittyChat.PublishRequest) (*ChittyChat.Response, error) {
	validateMessage, err := ValidateMessage(message.GetRequest())
	if validateMessage {
		Logger("Request was accepted, and is being broadcasted, "+strconv.Itoa(int(message.GetClientId())), serverLogFile)
		Broadcast(message.GetRequest(), int(message.GetClientId()))
		return &ChittyChat.Response{Msg: "Request was accepted, and is being broadcasted"}, nil
	} else {
		Logger(err.Error(), "errorLog")
		return &ChittyChat.Response{Msg: err.Error()}, err
	}
}

func (s *Server) JoinChat(ctx context.Context, _ *ChittyChat.JoinChatRequest) (*ChittyChat.JoinResponse, error) {

	clientsConnectedVectorClocks = append(clientsConnectedVectorClocks, 0)
	clientId := clientCount
	clientCount++

	msg := "client: " + strconv.Itoa(clientId) + " succesfully joined the chat"
	Logger(msg+", vectorClock: "+FormatVectorClock(clientsConnectedVectorClocks), serverLogFile)

	return &ChittyChat.JoinResponse{ClientId: int32(clientId)}, nil
}

func (s *Server) LeaveChat(ctx context.Context, request *ChittyChat.LeaveChatRequest) (*ChittyChat.LeaveResponse, error) {

	msg := "clientId succesfully left the chat"
	Logger(msg+", vectorClock: "+FormatVectorClock(clientsConnectedVectorClocks), serverLogFile)

	// "remove client" setting vectorClock at clientId to 0
	clientId := request.GetClientId()
	clientsConnectedVectorClocks[clientId] = 0
	return &ChittyChat.LeaveResponse{Msg: msg}, nil
}

func Broadcast(msg string, clientId int) {
	Logger(msg+", "+strconv.Itoa(clientId), serverLogFile)

	lock.Lock()
	clientsConnectedVectorClocks[clientId]++
	vectorClock := clientsConnectedVectorClocks
	lock.Unlock()

	broadCastBuffer <- bufferedMessage{message: msg, vectorTimeStamp: vectorClock, clientId: int32(clientId)}
}

// help method
func EvalLatestBroadCast(broadCastBuffer chan (bufferedMessage)) {
	for {
	Select:
		select {
		case recent := <-broadCastBuffer:
			broadCastIsNewer := false
			broadCastIsLatest := true
			if len(clientsConnectedVectorClocks) > len(recent.vectorTimeStamp) {
				break Select
			}
			for i := 0; i < len(clientsConnectedVectorClocks); i++ {
				if clientsConnectedVectorClocks[i] < recent.vectorTimeStamp[i] {
					broadCastIsNewer = true
					clientsConnectedVectorClocks[i] = recent.vectorTimeStamp[i]
				}
				if clientsConnectedVectorClocks[i] != recent.vectorTimeStamp[i] {
					broadCastIsLatest = false
				}
			}
			if broadCastIsNewer || broadCastIsLatest {
				broadCastBuffer <- recent
			}
		default:
			// If nothing is in the buffer sleep 0.25 seconds
			time.Sleep(time.Millisecond * 250)
		}
	}
}

func Logger(message string, logFileName string) {
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

func Check(err error, message string) {
	if err != nil {
		panic(err)
	}
	log.Printf("%s\n", message)
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
