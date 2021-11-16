package main

import (
	"context"
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

var services = make(map[uint16]Service)

func init() {
	services[0x17] = &AuthorizationRegistrationService{}
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
		if r := recover(); r != nil {
			log.Println("Error handling message: ", r.(error).Error())
		}
		conn.Close()
		log.Printf("Closed connection to %v", conn.RemoteAddr())
	})()

	buf := make([]byte, 1024)
	for {
		session := ctx.Value("session").(*Session)
		if !session.GreetedClient {
			// send a hello
			hello := NewFLAP(ctx, 1, []byte{0, 0, 0, 1})
			err := session.Send(hello)
			panicIfError(err)
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
	flap := &FLAP{}
	flap.UnmarshalBinary(buf)

	if flap.Header.Channel == 1 {

	} else if flap.Header.Channel == 2 {
		snac := &SNAC{}
		err := snac.UnmarshalBinary(flap.Data)
		panicIfError(err)

		if service, ok := services[snac.Header.Family]; ok {
			service.HandleSNAC(ctx, snac)
		}
	}
}
