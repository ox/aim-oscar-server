package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type SNACHeader struct {
	Family    uint16
	Subtype   uint16
	Flags     uint16
	RequestID uint16
}

type SNAC struct {
	Header SNACHeader
	Data   []byte
}

func NewSNAC(family uint16, subtype uint16, data []byte) *SNAC {
	return &SNAC{
		Header: SNACHeader{
			Family:    family,
			Subtype:   subtype,
			Flags:     0,
			RequestID: 0,
		},
		Data: data,
	}
}

func (s *SNAC) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, s.Header)
	n, err := buf.Write(s.Data)
	if n != len(s.Data) {
		return nil, fmt.Errorf("needed to write %d bytes to buffer but wrote %d", len(s.Data), n)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *SNAC) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)

	if err := binary.Read(buf, binary.BigEndian, &s.Header); s != nil {
		return err
	}

	s.Data = buf.Bytes()

	return nil
}
