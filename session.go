package main

import (
	"context"
	"encoding"
	"fmt"
	"net"
)

type sessionKey string

func (s sessionKey) String() string {
	return "oscar-" + string(s)
}

var (
	currentSession = sessionKey("session")
)

type Session struct {
	Conn           net.Conn
	SequenceNumber uint16
	GreetedClient  bool
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn:           conn,
		SequenceNumber: 0,
		GreetedClient:  false,
	}
}

func NewContextWithSession(ctx context.Context, conn net.Conn) context.Context {
	session := NewSession(conn)
	return context.WithValue(ctx, currentSession, session)
}

func CurrentSession(ctx context.Context) (session *Session, err error) {
	session, ok := ctx.Value(currentSession).(*Session)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}
	return
}

func (s *Session) Send(m encoding.BinaryMarshaler) error {
	bytes, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	fmt.Printf("-> %v\n%s\n\n", s.Conn.RemoteAddr(), prettyBytes(bytes))
	_, err = s.Conn.Write(bytes)
	return err
}
