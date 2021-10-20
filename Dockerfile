FROM golang:latest

RUN mkdir /build
WORKDIR /build

RUN export GO111MODULE=on

COPY go.mod /build
COPY go.sum /build

RUN cd /build/ && git clone https://github.com/lottejd/DISYSMP2.git

RUN cd /build/DISYSMP2
RUN mv DISYSMP2 disysmp2
RUN cd /build/disysmp2/Server
RUN mv Server server
RUN cd /build/disysmp2/Server && go build ./...



EXPOSE 8080


ENTRYPOINT [ "/build/disysmp2/Server/Server" ]