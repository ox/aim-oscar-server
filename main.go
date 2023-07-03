package main

import (
	"aim-oscar/config"
	"aim-oscar/db"
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/services"
	"aim-oscar/util"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uptrace/bun/extra/bundebug"
)

const (
	LogLevelDebug = "debug"
)

func main() {
	configPath := flag.String("config", "", "Path to app config")
	flag.Parse()

	if configPath == nil || *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	conf, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("could not parse config: %s", err)
	}

	db, err := db.Connect(&conf.DBConfig)
	if err != nil {
		log.Fatalf("could not connect to DB: %s", err)
	}

	// Print all queries to stdout.
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(conf.AppConfig.LogLevel == "debug")))

	// Register our DB models
	db.RegisterModel((*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil))

	listener, err := net.Listen("tcp", conf.OscarConfig.Addr)
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
	serviceManager.RegisterService(0x01, &services.GenericServiceControls{OnlineCh: onlineCh, ServerHostname: conf.OscarConfig.Addr})
	serviceManager.RegisterService(0x02, &services.LocationServices{OnlineCh: onlineCh})
	serviceManager.RegisterService(0x03, &services.BuddyListManagement{OnlineCh: onlineCh})
	serviceManager.RegisterService(0x04, &services.ICBM{CommCh: commCh})
	serviceManager.RegisterService(0x17, &services.AuthorizationRegistrationService{BOSAddress: conf.OscarConfig.Addr})

	handleCloseFn := func(ctx context.Context, session *oscar.Session) {
		log.Printf("%v disconnected", session.RemoteAddr())

		user := models.UserFromContext(ctx)
		if user != nil {
			if err := user.SetAway(ctx, db); err != nil {
				log.Printf("could not set user as away: %s", err)
			}

			log.Printf("closing down user %s\n", user.ScreenName)

			onlineCh <- user
			if session, err := oscar.SessionFromContext(ctx); err == nil {
				session.Disconnect()
				sessionManager.RemoveSession(user.ScreenName)
			}
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
			if conf.AppConfig.LogLevel == LogLevelDebug {
				log.Printf("FROM %s (%v)\n%+v\n", user.ScreenName, session.RemoteAddr(), flap)
			}
			user.LastActivityAt = time.Now()
			ctx = models.NewContextWithUser(ctx, user)
			session.ScreenName = user.ScreenName
			sessionManager.SetSession(user.ScreenName, session)
		} else {
			if conf.AppConfig.LogLevel == LogLevelDebug {
				log.Printf("FROM %v\n%+v\n", session.RemoteAddr(), flap)
			}
		}

		if flap.Header.Channel == 1 {
			// Is this a hello?
			if bytes.Equal(flap.Data.Bytes(), []byte{0, 0, 0, 1}) {
				log.Println("this is a hello")
				return ctx
			}

			user, err := services.AuthenticateFLAPCookie(ctx, db, flap)
			if err != nil {
				log.Printf("Could not authenticate cookie: %s", err)
				return ctx
			}

			session.ScreenName = user.ScreenName
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
			if err := snac.UnmarshalBinary(flap.Data.Bytes()); err != nil {
				log.Println("could not unmarshal FLAP data:", err)
				session.Disconnect()
				handleCloseFn(ctx, session)
				return ctx
			}

			if conf.AppConfig.LogLevel == LogLevelDebug {
				fmt.Printf("%s\n", snac)
				if tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes()); err == nil {
					for _, tlv := range tlvs {
						fmt.Printf("%s\n", tlv)
					}
				} else {
					fmt.Printf("%s\n\n", util.PrettyBytes(snac.Data.Bytes()))
				}
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

		log.Println("Shutting down")
		os.Exit(1)
	}()

	log.Println("OSCAR listening on " + conf.OscarConfig.Addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		log.Printf("Connection from %v", conn.RemoteAddr())
		go handler.Handle(conn)
	}
}
