package main

import (
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/services"
	"aim-oscar/util"
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var (
	OSCAR_HOST        = "0.0.0.0"
	OSCAR_PORT        = "5190"
	OSCAR_ADDRESS     = OSCAR_HOST + ":" + OSCAR_PORT
	OSCAR_BOS_HOST    = OSCAR_HOST
	OSCAR_BOS_PORT    = OSCAR_PORT
	OSCAR_BOS_ADDRESS = OSCAR_BOS_HOST + ":" + OSCAR_BOS_PORT
	DB_URL            = ""
	DB_USER           = ""
	DB_PASSWORD       = ""
)

func init() {
	if oscarHost, ok := os.LookupEnv("OSCAR_HOST"); ok {
		OSCAR_HOST = oscarHost
	}
	var oscarHost string
	flag.StringVar(&oscarHost, "host", OSCAR_HOST, "Server hostname")

	if oscarPort, ok := os.LookupEnv("OSCAR_PORT"); ok {
		OSCAR_PORT = oscarPort
	}
	var oscarPort string
	flag.StringVar(&oscarPort, "port", OSCAR_PORT, "Server port")

	if oscarBOSHost, ok := os.LookupEnv("OSCAR_BOS_HOST"); ok {
		OSCAR_BOS_HOST = oscarBOSHost
	}
	var oscarBOSHost string
	flag.StringVar(&oscarBOSHost, "boshost", OSCAR_BOS_HOST, "BOS hostname")

	if oscarBOSPort, ok := os.LookupEnv("OSCAR_BOS_PORT"); ok {
		OSCAR_BOS_PORT = oscarBOSPort
	}
	var oscarBOSPort string
	flag.StringVar(&oscarBOSPort, "bosport", OSCAR_BOS_PORT, "BOS port")

	if dbUrl, ok := os.LookupEnv("DB_URL"); ok {
		DB_URL = strings.TrimSpace(dbUrl)
	}

	if dbUser, ok := os.LookupEnv("DB_USER"); ok {
		DB_USER = strings.TrimSpace(dbUser)
	}

	if dbPassword, ok := os.LookupEnv("DB_PASSWORD"); ok {
		DB_PASSWORD = strings.TrimSpace(dbPassword)
	}

	flag.Parse()

	OSCAR_HOST = oscarHost
	OSCAR_PORT = oscarPort
	OSCAR_ADDRESS = OSCAR_HOST + ":" + OSCAR_PORT

	OSCAR_BOS_HOST = oscarBOSHost
	OSCAR_BOS_PORT = oscarBOSPort
	OSCAR_BOS_ADDRESS = OSCAR_BOS_HOST + ":" + OSCAR_BOS_PORT

	if DB_URL == "" {
		log.Fatalln("DB Url not specified")
	}

	if DB_USER == "" {
		log.Fatalln("DB User not specified")
	}

	if DB_PASSWORD == "" {
		log.Fatalln("DB password not specified")
	}
}

func main() {
	pgconn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(DB_URL),
		pgdriver.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		pgdriver.WithUser(DB_USER),
		pgdriver.WithPassword(DB_PASSWORD),
		pgdriver.WithDatabase("postgres"),
		pgdriver.WithInsecure(true),
		pgdriver.WithTimeout(5*time.Second),
		pgdriver.WithDialTimeout(5*time.Second),
		pgdriver.WithReadTimeout(5*time.Second),
		pgdriver.WithWriteTimeout(5*time.Second),
	)

	log.Printf("DB URL: %s", DB_URL)

	// Set up the DB
	sqldb := sql.OpenDB(pgconn)
	db := bun.NewDB(sqldb, pgdialect.New())
	db.SetConnMaxIdleTime(15 * time.Second)
	db.SetConnMaxLifetime(1 * time.Minute)

	// Print all queries to stdout.
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Register our DB models
	db.RegisterModel((*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil))

	listener, err := net.Listen("tcp", OSCAR_ADDRESS)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	sessionManager := NewSessionManager()

	// Goroutine that listens for messages to deliver and tries to find a user socket to push them to
	commCh, messageRoutine := MessageDelivery(sessionManager)
	go messageRoutine(db)

	// Goroutine that listens for users who change their online status and notifies their buddies
	onlineCh, onlineRoutine := OnlineNotification(sessionManager)
	go onlineRoutine(db)

	serviceManager := NewServiceManager()
	serviceManager.RegisterService(0x01, &services.GenericServiceControls{OnlineCh: onlineCh, ServerHostname: OSCAR_ADDRESS})
	serviceManager.RegisterService(0x02, &services.LocationServices{OnlineCh: onlineCh})
	serviceManager.RegisterService(0x03, &services.BuddyListManagement{})
	serviceManager.RegisterService(0x04, &services.ICBM{CommCh: commCh})
	serviceManager.RegisterService(0x17, &services.AuthorizationRegistrationService{BOSAddress: OSCAR_BOS_ADDRESS})

	handleCloseFn := func(ctx context.Context, session *oscar.Session) {
		log.Printf("%v disconnected", session.RemoteAddr())

		user := models.UserFromContext(ctx)
		if user != nil {
			user.Status = models.UserStatusAway
			user.Cipher = ""
			if err := user.Update(ctx, db); err != nil {
				log.Print(errors.Wrap(err, "could not set user as inactive"))
			}

			onlineCh <- user
			sessionManager.RemoveSession(user.ScreenName)
		}
	}

	handleFn := func(ctx context.Context, flap *oscar.FLAP) context.Context {
		session, err := oscar.SessionFromContext(ctx)
		if err != nil {
			// TODO
			log.Printf("no session in context. FLAP dump:\n%s\n", flap)
			return ctx
		}

		if user := models.UserFromContext(ctx); user != nil {
			fmt.Printf("%s (%v) ->\n%+v\n", user.ScreenName, session.RemoteAddr(), flap)
			user.LastActivityAt = time.Now()
			ctx = models.NewContextWithUser(ctx, user)
			sessionManager.SetSession(user.ScreenName, session)
		} else {
			fmt.Printf("%v ->\n%+v\n", session.RemoteAddr(), flap)
		}

		if flap.Header.Channel == 1 {
			// Is this a hello?
			if bytes.Equal(flap.Data.Bytes(), []byte{0, 0, 0, 1}) {
				return ctx
			}

			user, err := services.AuthenticateFLAPCookie(ctx, db, flap)
			if err != nil {
				log.Printf("Could not authenticate cookie: %s", err)
				return ctx
			}
			ctx = models.NewContextWithUser(ctx, user)

			// Send available services
			servicesSnac := oscar.NewSNAC(1, 3)
			for family := range services.ServiceVersions {
				servicesSnac.Data.WriteUint16(family)
			}

			servicesFlap := oscar.NewFLAP(2)
			servicesFlap.Data.WriteBinary(servicesSnac)
			session.Send(servicesFlap)

			return ctx
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

			if service, ok := serviceManager.GetService(snac.Header.Family); ok {
				newCtx, err := service.HandleSNAC(ctx, db, snac)
				if err != nil {
					log.Printf("encountered error: %s", err)
					session.Disconnect()
					handleCloseFn(ctx, session)
				}

				return newCtx
			}
		} else if flap.Header.Channel == 4 {
			session.Disconnect()
			handleCloseFn(ctx, session)
		}

		return ctx
	}

	handler := oscar.NewHandler(handleFn, handleCloseFn)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-exitChan
		close(commCh)
		close(onlineCh)
		fmt.Println("Shutting down")
		os.Exit(1)
	}()

	fmt.Println("OSCAR listening on " + OSCAR_ADDRESS)
	fmt.Println("OSCAR BOS set to " + OSCAR_BOS_ADDRESS)
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
