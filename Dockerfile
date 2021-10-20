FROM golang:latest

RUN mkdir /build
WORKDIR /build

RUN export GO111MODULE=on

COPY go.mod /build
COPY go.sum /build

RUN cd /build/ && git clone https://github.com/lottejd/DISYSMP2.git
RUN cd /build/DISYSMP2/Server && go build ./...

EXPOSE 9080

ENTRYPOINT [ "/build/DISYSMP2/Server/server" ]