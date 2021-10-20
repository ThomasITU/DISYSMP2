# DISYSMP2

ASK TA

- UTF-8
- Timestamp hos serveren
- how to broadcast
- skal hver client have en log?
- er der en smartere måde at broadcaste på end at clienten lytter i et for evigheds loop


# Commands

## proto:
- protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ChittyChat/ChittyChat.proto 

## docker:

- $env:DOCKER_BUILDKIT=1

- docker build -t test .

- docker run -p 8080:8080 -tid test

