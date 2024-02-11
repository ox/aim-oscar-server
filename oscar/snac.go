package oscar

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
	Data   Buffer
}

func NewSNAC(family uint16, subtype uint16) *SNAC {
	return &SNAC{
		Header: SNACHeader{
			Family:    family,
			Subtype:   subtype,
			Flags:     0,
			RequestID: 0,
		},
	}
}

func (s *SNAC) MarshalBinary() ([]byte, error) {
	buf := Buffer{}

	binary.Write(&buf, binary.BigEndian, s.Header)
	b := s.Data.Bytes()
	n, err := buf.Write(b)
	if n != len(b) {
		return nil, fmt.Errorf("needed to write %d bytes to buffer but wrote %d", len(b), n)
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

	s.Data.Write(buf.Bytes())

	return nil
}

func (s *SNAC) String() string {
	return fmt.Sprintf("SNAC(%#x, %#x)", s.Header.Family, s.Header.Subtype)
}

func (s *SNAC) WriteTLV(tlv *TLV) {
	s.Data.WriteBinary(tlv)
}

func (s *SNAC) AppendTLVs(tlvs []*TLV) {
	s.Data.WriteUint16(uint16(len(tlvs))) // number of TLVs
	for _, tlv := range tlvs {
		s.Data.WriteBinary(tlv)
	}
}
