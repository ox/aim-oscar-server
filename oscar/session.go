package oscar

import (
	"aim-oscar/util"
	"context"
	"log"
	"net"

	"github.com/pkg/errors"
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
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:           conn,
		SequenceNumber: 0,
		GreetedClient:  false,
		ScreenName:     "",
	}
}

func NewContextWithSession(ctx context.Context, conn net.Conn) context.Context {
	session := NewSession(conn)
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

	if s.ScreenName != "" {
		log.Printf("SEND TO %s (%v)\n%s\n\n", s.ScreenName, s.conn.RemoteAddr(), util.PrettyBytes(bytes))
	} else {
		log.Printf("SEND TO %v\n%s\n\n", s.conn.RemoteAddr(), util.PrettyBytes(bytes))
	}

	_, err = s.conn.Write(bytes)
	return errors.Wrap(err, "could not write to client connection")
}

func (s *Session) Disconnect() error {
	return s.conn.Close()
}
