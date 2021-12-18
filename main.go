package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/util"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

var services map[uint16]Service

func init() {
	services = make(map[uint16]Service)
}

func RegisterService(family uint16, service Service) {
	services[family] = service
}

func main() {
	// Set up the DB
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	db.SetConnMaxIdleTime(15 * time.Second)
	db.SetConnMaxLifetime(1 * time.Minute)

	// Print all queries to stdout.
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Register our DB models
	db.RegisterModel((*models.User)(nil))

	// dev: load in fixtures to test against
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

	handler := oscar.NewHandler(func(session *oscar.Session, flap *oscar.FLAP) {
		if flap.Header.Channel == 1 {
			// Is this a hello?
			if bytes.Equal(flap.Data.Bytes(), []byte{0, 0, 0, 1}) {
				return
			}
		} else if flap.Header.Channel == 2 {
			snac := &oscar.SNAC{}
			err := snac.UnmarshalBinary(flap.Data.Bytes())
			util.PanicIfError(err)

			fmt.Printf("%+v\n", snac)
			if tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes()); err == nil {
				for _, tlv := range tlvs {
					fmt.Printf("%+v\n", tlv)
				}
			} else {
				fmt.Printf("%s\n\n", util.PrettyBytes(snac.Data.Bytes()))
			}

			if service, ok := services[snac.Header.Family]; ok {
				err = service.HandleSNAC(db, session, snac)
				util.PanicIfError(err)
			}
		}
	})

	RegisterService(0x17, &AuthorizationRegistrationService{})

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

		log.Printf("Connection from %v", conn.RemoteAddr())
		go handler.Handle(conn)
	}
}
