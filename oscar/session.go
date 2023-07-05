package oscar

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

type sessionKey string

func (s sessionKey) String() string {
	return "oscar-" + string(s)
}

var (
	currentSession = sessionKey("session")
)

type Session struct {
	conn           net.Conn
	SequenceNumber uint16
	GreetedClient  bool
	ScreenName     string
	Logger         *slog.Logger
}

func NewSession(conn net.Conn, logger *slog.Logger) *Session {
	return &Session{
		conn:           conn,
		SequenceNumber: 0,
		GreetedClient:  false,
		ScreenName:     "",
		Logger:         logger,
	}
}

func NewContextWithSession(ctx context.Context, conn net.Conn, logger *slog.Logger) context.Context {
	session := NewSession(conn, logger)
	return context.WithValue(ctx, currentSession, session)
}

func SessionFromContext(ctx context.Context) (session *Session, err error) {
	s := ctx.Value(currentSession)
	if s == nil {
		return nil, errors.New("no session in context")
	}
	return s.(*Session), nil
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Session) Send(flap *FLAP) error {
	s.SequenceNumber += 1
	flap.Header.SequenceNumber = s.SequenceNumber
	bytes, err := flap.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "could not marshal message")
	}

	if s.Logger != nil {
		if s.ScreenName != "" {
			s.Logger.Debug("SEND",
				slog.String("screen_name", s.ScreenName),
				"flap",
				flap,
			)
		} else {
			s.Logger.Debug("SEND", "flap", flap)
		}
	}

	_, err = s.conn.Write(bytes)
	return errors.Wrap(err, "could not write to client connection")
}

func (s *Session) Disconnect() error {
	return s.conn.Close()
}
