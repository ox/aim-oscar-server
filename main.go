package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	SRV_HOST    = ""
	SRV_PORT    = "5190"
	SRV_ADDRESS = SRV_HOST + ":" + SRV_PORT
)

type Session struct {
	Conn          net.Conn
	GreetedClient bool
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn:          conn,
		GreetedClient: false,
	}
}

func (s *Session) Send(bytes []byte) {
	fmt.Printf("-> %v\n%s\n\n", s.Conn.RemoteAddr(), prettyBytes(bytes))
	s.Conn.Write(bytes)
}

func main() {
	listener, err := net.Listen("tcp", SRV_ADDRESS)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-exitChan
		fmt.Println("Shutting down")
		os.Exit(1)
	}()

	fmt.Println("OSCAR listening on " + SRV_ADDRESS)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		ctx := context.WithValue(context.Background(), "session", NewSession(conn))
		log.Printf("Connection from %v", conn.RemoteAddr())

		go handleTCPConnection(ctx, conn)
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn) {
	defer (func() {
		recover()
		conn.Close()
		log.Printf("Closed connection to %v", conn.RemoteAddr())
	})()

	buf := make([]byte, 1024)
	for {
		session := ctx.Value("session").(*Session)
		if !session.GreetedClient {
			// send a hello
			hello := []byte{0x2a, 1, 0, 0, 0, 4, 0, 0, 0, 1}
			session.Send(hello)
			session.GreetedClient = true
		}

		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("Read Error: ", err.Error())
			return
		}

		if n == 0 {
			return
		}

		fmt.Printf("%v ->\n%s\n\n", conn.RemoteAddr(), prettyBytes(buf[:n]))
		handleMessage(ctx, buf[:n])
	}
}

func handleMessage(ctx context.Context, buf []byte) {
	messageBuf := bytes.NewBuffer(buf)

	start, err := messageBuf.ReadByte()
	panicIfError(err)
	if start != 0x2a {
		log.Println("FLAP message missing leading 0x2a")
		return
	}

	// Start parsing FLAP header
	channel := mustReadNBytes(messageBuf, 1)[0]
	log.Println("Message for channel: ", channel)

	datagramSeqNum := mustReadNBytes(messageBuf, 2)
	log.Println("Datagram Sequence Number: ", binary.BigEndian.Uint16(datagramSeqNum))

	dataLength := mustReadNBytes(messageBuf, 2)
	log.Println("Data Length: ", binary.BigEndian.Uint16(dataLength))

}
