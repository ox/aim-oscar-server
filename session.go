package main

import (
	"encoding"
	"fmt"
	"net"
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

func (s *Session) Send(m encoding.BinaryMarshaler) error {
	bytes, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	fmt.Printf("-> %v\n%s\n\n", s.Conn.RemoteAddr(), prettyBytes(bytes))
	_, err = s.Conn.Write(bytes)
	return err
}
