package main

import (
	"aim-oscar/models"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
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
	// Set up the DB
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Print all queries to stdout.
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	db.RegisterModel((*models.User)(nil))

	fixture := dbfixture.New(db, dbfixture.WithRecreateTables())
	err = fixture.Load(context.Background(), os.DirFS("models"), "fixtures.yml")
	if err != nil {
		panic(err)
	}

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

		session := NewSession(conn)
		log.Printf("Connection from %v", conn.RemoteAddr())

		go handleTCPConnection(db, session, conn)
	}
}

func handleTCPConnection(db *bun.DB, session *Session, conn net.Conn) {
	// defer (func() {
	// 	if err := recover(); err != nil {
	// 		log.Printf("Error handling message: %+v\n", err.(error))
	// 	}
	// 	conn.Close()
	// 	log.Printf("Closed connection to %v", conn.RemoteAddr())
	// })()

	buf := make([]byte, 1024)
	for {
		if !session.GreetedClient {
			// send a hello
			hello := NewFLAP(1)
			hello.Data.Write([]byte{0, 0, 0, 1})
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

		// Try to parse all of the FLAPs in the buffer if we have enough bytes to
		// fill a FLAP header
		for len(buf) >= 6 && buf[0] == 0x2a {
			dataLength := Word(buf[4:6])
			flapLength := int(dataLength) + 6
			if len(buf) < flapLength {
				log.Printf("not enough data, only %d bytes\n", len(buf))
				break
			}

			flap := &FLAP{}
			if err := flap.UnmarshalBinary(buf[:flapLength]); err != nil {
				panicIfError(errors.Wrap(err, "could not unmarshal FLAP"))
			}
			buf = buf[flapLength:]
			fmt.Printf("%v ->\n%+v\n", conn.RemoteAddr(), flap)
			handleMessage(db, session, flap)
		}
	}
}

func handleMessage(db *bun.DB, session *Session, flap *FLAP) {
	if flap.Header.Channel == 1 {

	} else if flap.Header.Channel == 2 {
		snac := &SNAC{}
		err := snac.UnmarshalBinary(flap.Data.Bytes())
		panicIfError(err)

		fmt.Printf("%+v\n", snac)
		if tlvs, err := UnmarshalTLVs(snac.Data.Bytes()); err == nil {
			for _, tlv := range tlvs {
				fmt.Printf("%+v\n", tlv)
			}
		} else {
			fmt.Printf("%s\n\n", prettyBytes(snac.Data.Bytes()))
		}

		if service, ok := services[snac.Header.Family]; ok {
			err = service.HandleSNAC(db, session, snac)
			panicIfError(err)
		}
	}
}
