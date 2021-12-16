package main

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
)

var _ encoding.BinaryUnmarshaler = &FLAP{}
var _ encoding.BinaryMarshaler = &FLAP{}

type FLAPHeader struct {
	Channel        uint8
	SequenceNumber uint16
	DataLength     uint16
}

type FLAP struct {
	Header FLAPHeader
	Data   []byte
}

func NewFLAP(session *Session, channel uint8, data []byte) *FLAP {
	session.SequenceNumber += 1

	return &FLAP{
		Header: FLAPHeader{
			Channel:        channel,
			SequenceNumber: uint16(session.SequenceNumber),
			DataLength:     uint16(len(data)),
		},
		Data: data,
	}
}

func (f *FLAP) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(0x2a)
	binary.Write(&buf, binary.BigEndian, f.Header)
	n, err := buf.Write(f.Data)
	if n != int(f.Header.DataLength) {
		return nil, fmt.Errorf("needed to write %d bytes to buffer but wrote %d", f.Header.DataLength, n)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (f *FLAP) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	start, err := buf.ReadByte()
	if err != nil {
		return err
	}
	if start != 0x2a {
		return fmt.Errorf("FLAP missing 0x2a header")
	}

	if err = binary.Read(buf, binary.BigEndian, &f.Header); err != nil {
		return err
	}

	f.Data = buf.Bytes()
	return nil
}

func (f *FLAP) Len() int {
	return 6 + int(f.Header.DataLength)
}

func (f *FLAP) String() string {
	return fmt.Sprintf("FLAP(CH:%d, SEQ:%d):\n%s", f.Header.Channel, f.Header.SequenceNumber, prettyBytes(f.Data))
}
