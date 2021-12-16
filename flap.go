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
	Data   Buffer
}

func NewFLAP(channel uint8) *FLAP {
	return &FLAP{
		Header: FLAPHeader{
			Channel:        channel,
			SequenceNumber: 0,
			DataLength:     0,
		},
	}
}

func (f *FLAP) MarshalBinary() ([]byte, error) {
	buf := Buffer{}
	buf.WriteUint8(0x2a)

	f.Header.DataLength = uint16(len(f.Data.Bytes()))

	binary.Write(&buf, binary.BigEndian, f.Header)
	n, err := buf.Write(f.Data.Bytes())
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

	f.Data.Write(buf.Bytes())
	return nil
}

func (f *FLAP) Len() int {
	return 6 + int(f.Header.DataLength)
}

func (f *FLAP) String() string {
	return fmt.Sprintf("FLAP(CH:%d, SEQ:%d):\n%s", f.Header.Channel, f.Header.SequenceNumber, prettyBytes(f.Data.Bytes()))
}
