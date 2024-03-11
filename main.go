package main

import (
	"aim-oscar/config"
	"aim-oscar/db"
	"aim-oscar/models"
	"aim-oscar/services"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uptrace/bun/extra/bundebug"
	"golang.org/x/exp/slog"
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

	var level slog.Level = slog.LevelDebug
	if err := level.UnmarshalText([]byte(conf.AppConfig.LogLevel)); err != nil {
		log.Fatalf("invalid app.log_level: %s", err)
	}

	var logHandler slog.Handler = NewOSCARLogHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	if conf.AppConfig.LogStyle == "machine" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	db, err := db.Connect(&conf.DBConfig)
	if err != nil {
		logger.Error("could not connect to DB", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// Print all queries to stdout.
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(conf.AppConfig.LogLevel == slog.LevelDebug.String())))

	// Register our DB models
	db.RegisterModel((*models.User)(nil), (*models.Message)(nil), (*models.Buddy)(nil), (*models.EmailVerification)(nil))

	// On start, all users must be offline bc there are no connections (while this is a one-server operation)
	ctx := context.Background()
	if _, err := db.NewUpdate().Model(&models.User{}).Set("status = ?", models.UserStatusAway).Where("status != ?", models.UserStatusAway).Exec(ctx); err != nil {
		logger.Error("could not set all users as offline", "err", err.Error())
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", conf.OscarConfig.Addr)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	sessionManager := NewSessionManager()

	// Goroutine that listens for messages to deliver and tries to find a user socket to push them to
	commCh, messageRoutine := MessageDelivery(sessionManager, logger)
	go messageRoutine(db)

	// Goroutine that listens for users who change their online status and notifies their buddies
	onlineCh, onlineRoutine := OnlineNotification(sessionManager, logger)
	go onlineRoutine(db)

	serviceManager := NewServiceManager()
	serviceManager.RegisterService(0x01, &services.GenericServiceControls{OnlineCh: onlineCh, ServerHostname: conf.OscarConfig.Addr})
	serviceManager.RegisterService(0x02, &services.LocationServices{OnlineCh: onlineCh})
	serviceManager.RegisterService(0x03, &services.BuddyListManagement{OnlineCh: onlineCh})
	serviceManager.RegisterService(0x04, &services.ICBM{CommCh: commCh})
	// serviceManager.RegisterService(0x0f, &services.DirectorySearchService{})
	// serviceManager.RegisterService(0x13, &services.FeedbagService{})
	serviceManager.RegisterService(0x17, &services.AuthorizationRegistrationService{BOSAddress: conf.OscarConfig.BOS})
	serviceManager.RegisterService(0x18, &services.AlertService{})

	handler := NewHandler(&conf.AppConfig, db, logger, sessionManager, serviceManager, onlineCh)

	var metricsServer *http.Server
	if conf.AppConfig.Metrics.Addr != "" {
		mux := http.NewServeMux()
		metricsHandler := promhttp.Handler()

		if conf.AppConfig.Metrics.User != "" && conf.AppConfig.Metrics.Password != "" {
			metricsHandler = BasicAuth(promhttp.Handler().ServeHTTP, conf.AppConfig.Metrics.User, conf.AppConfig.Metrics.Password, "identify yourself")
		}

		mux.Handle("/metrics", metricsHandler)
		metricsServer = &http.Server{
			Addr:    conf.AppConfig.Metrics.Addr,
			Handler: mux,
		}
		go func() {
			logger.Info("Metrics handler started", "metrics_server_addr", metricsServer.Addr)
			metricsServer.ListenAndServe()
		}()
	}

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		<-exitChan
		close(commCh)
		close(onlineCh)

		if metricsServer != nil {
			metricsServer.Close()
		}

		logger.Info("Shutting down")
		os.Exit(1)
	}()

	logger.Info("Listening on " + conf.OscarConfig.Addr)
	logger.Info("BOS host " + conf.OscarConfig.BOS)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handler.Handle(conn, logger)
	}
}
