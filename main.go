package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
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

	handler := oscar.NewHandler()
	handler.RegisterService(0x17, &AuthorizationRegistrationService{db: db})

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
