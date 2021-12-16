package main

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
)

var _ encoding.BinaryUnmarshaler = &SNAC{}
var _ encoding.BinaryMarshaler = &SNAC{}

type SNACHeader struct {
	Family    uint16
	Subtype   uint16
	Flags     uint16
	RequestID uint32
}

type SNAC struct {
	Header SNACHeader
	Data   []byte
}

func NewSNAC(family uint16, subtype uint16, data []byte) *SNAC {
	d := make([]byte, 0, len(data))
	copy(d, data)

	return &SNAC{
		Header: SNACHeader{
			Family:    family,
			Subtype:   subtype,
			Flags:     0,
			RequestID: 0,
		},
		Data: d,
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
	if err := binary.Read(buf, binary.BigEndian, &s.Header); err != nil {
		return err
	}

	s.Data = make([]byte, buf.Len())
	copy(s.Data, buf.Bytes())

	return nil
}

func (s *SNAC) String() string {
	return fmt.Sprintf("SNAC(%#x, %#x)", s.Header.Family, s.Header.Subtype)
}
