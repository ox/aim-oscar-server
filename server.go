package main

import (
	"aim-oscar/config"
	"aim-oscar/models"
	"aim-oscar/oscar"
	"aim-oscar/services"
	"aim-oscar/util"
	"bytes"
	"context"

	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slog"
)

type HandlerFunc func(context.Context, *oscar.FLAP) context.Context
type HandleCloseFn func(context.Context, *oscar.Session)

type Handler struct {
	conf           *config.AppConfig
	db             *bun.DB
	logger         *slog.Logger
	sessionManager *SessionManager
	serviceManager *ServiceManager
	onlineCh       chan *models.User
}

func NewHandler(conf *config.AppConfig, db *bun.DB, logger *slog.Logger, sm *SessionManager, svm *ServiceManager, onlineCh chan *models.User) *Handler {
	return &Handler{
		conf, db, logger, sm, svm, onlineCh,
	}
}

func (h *Handler) Handle(conn net.Conn, logger *slog.Logger) {
	connLogger := logger.With("session_id", uuid.New(), "ip", conn.RemoteAddr().String())
	connLogger.Info("New Connection")

	ctx := oscar.NewContextWithSession(context.Background(), conn, connLogger)
	session, err := oscar.SessionFromContext(ctx)
	if err != nil {
		connLogger.Error("could not create session for context", "err", err)
	}

	var buf bytes.Buffer
	for {
		if !session.GreetedClient {
			// send a hello
			hello := oscar.NewFLAP(1)
			hello.Data.Write([]byte{0, 0, 0, 1})
			session.Send(hello)
			session.GreetedClient = true
		}

		// Wait for some data to read
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		incoming := make([]byte, 512)
		n, err := conn.Read(incoming)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "use of closed network connection") {
				session.Disconnect()
				h.handleCloseFn(ctx, session)
				return
			}

			// If the read timed out, just try reading again
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			connLogger.Error("OSCAR Read Error", "err", err.Error())
			return
		}

		if n == 0 {
			return
		}

		buf.Write(incoming[:n])

		// Try to parse all of the FLAPs in the buffer if we have enough bytes to
		// fill a FLAP header
		for buf.Len() >= 6 && buf.Bytes()[0] == 0x2a {
			bufBytes := buf.Bytes()
			dataLength := binary.BigEndian.Uint16(bufBytes[4:6])
			flapLength := int(dataLength) + 6
			if len(bufBytes) < flapLength {
				connLogger.Error(fmt.Sprintf("not enough data, expected %d bytes but have %d bytes", flapLength, len(bufBytes)))
				fmt.Printf("%s\n", util.PrettyBytes(bufBytes))
				break
			}

			flap := &oscar.FLAP{}
			flapBuf := make([]byte, flapLength)
			buf.Read(flapBuf)
			if err := flap.UnmarshalBinary(flapBuf); err != nil {
				connLogger.Error("could not unmarshal FLAP", "err", err)
				// Toss out everything
				buf.Reset()
				break
			}

			ctx = h.handleFn(ctx, flap)
		}
	}
}

func (h *Handler) handleFn(ctx context.Context, flap *oscar.FLAP) context.Context {
	session, err := oscar.SessionFromContext(ctx)
	if err != nil {
		// TODO
		h.logger.Error("no session in context", "err", err, "flap", flap)
		return ctx
	}

	if user := models.UserFromContext(ctx); user != nil {
		if h.conf.LogLevel == slog.LevelDebug.String() {
			session.Logger.Debug("RECV",
				slog.String("screen_name", user.ScreenName),
				slog.String("ip", session.RemoteAddr().String()),
				"flap", flap,
			)
		}
		user.LastActivityAt = time.Now()
		ctx = models.NewContextWithUser(ctx, user)
		session.ScreenName = user.ScreenName
		h.sessionManager.SetSession(user.ScreenName, session)
	} else {
		if h.conf.LogLevel == slog.LevelDebug.String() {
			session.Logger.Debug("RECV",
				slog.String("ip", session.RemoteAddr().String()),
				"flap", flap,
			)
		}
	}

	if flap.Header.Channel == 1 {
		// Is this a hello?
		if bytes.Equal(flap.Data.Bytes(), []byte{0, 0, 0, 1}) {
			return ctx
		}

		user, screenName, err := services.AuthenticateFLAPCookie(ctx, h.db, flap)
		if err != nil {
			session.Logger.Error("Could not authenticate user cookie", "screen_name", screenName, slog.String("err", err.Error()))
			return ctx
		}

		session.Logger.Info("Authenticated user", "screen_name", user.ScreenName)

		session.ScreenName = user.ScreenName
		ctx = models.NewContextWithUser(ctx, user)

		// Send available services
		servicesSnac := oscar.NewSNAC(0x1, 0x3)
		for _, service := range services.ServiceVersions {
			servicesSnac.Data.WriteUint16(service.Family)
		}

		servicesFlap := oscar.NewFLAP(2)
		servicesFlap.Data.WriteBinary(servicesSnac)
		session.Send(servicesFlap)

		return ctx
	} else if flap.Header.Channel == 2 {
		snac := &oscar.SNAC{}
		if err := snac.UnmarshalBinary(flap.Data.Bytes()); err != nil {
			session.Logger.Error("could not unmarshal FLAP data", "err", err)
			session.Disconnect()
			h.handleCloseFn(ctx, session)
			return ctx
		}

		if service, ok := h.serviceManager.GetService(snac.Header.Family); ok {
			newCtx, err := service.HandleSNAC(ctx, h.db, snac)
			if err != nil {
				session.Logger.Error("error handling SNAC", slog.String("err", err.Error()))
				session.Disconnect()
				h.handleCloseFn(ctx, session)
			}

			return newCtx
		}
	} else if flap.Header.Channel == 4 {
		h.handleCloseFn(ctx, session)
	} else if flap.Header.Channel == 5 {
		// User is still connected
		// TODO: handle when user stops sending these messages?
		return ctx
		// session.Logger.Debug(fmt.Sprintf("%s is still connected", session.ScreenName))
	} else {
		session.Logger.Info("unhandled channel message", "channel", flap.Header.Channel, "flap", flap)
	}

	return ctx
}

func (h *Handler) handleCloseFn(ctx context.Context, session *oscar.Session) {
	session.Logger.Info("Disconnected")

	user := models.UserFromContext(ctx)
	if user != nil {
		if err := user.SetAway(ctx, h.db); err != nil {
			h.logger.Error("Could not set user as away", slog.String("err", err.Error()))
		}

		h.logger.Info("Disconnecting user", slog.String("screen_name", user.ScreenName))

		h.onlineCh <- user
		if session, err := oscar.SessionFromContext(ctx); err == nil {
			session.Disconnect()
			h.sessionManager.RemoveSession(user.ScreenName)
		}
	}
}
