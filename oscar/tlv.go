package oscar

import (
	"aim-oscar/util"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var _ encoding.BinaryUnmarshaler = &TLV{}
var _ encoding.BinaryMarshaler = &TLV{}

type TLV struct {
	Type       uint16
	DataLength uint16
	Data       []byte
}

func NewTLV(tlvType uint16, data []byte) *TLV {
	return &TLV{
		Type:       tlvType,
		DataLength: uint16(len(data)),
		Data:       data,
	}
}

func (t *TLV) Len() int {
	return 4 + int(t.DataLength)
}

func (t *TLV) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 4+t.DataLength)
	binary.BigEndian.PutUint16(buf[:2], t.Type)
	binary.BigEndian.PutUint16(buf[2:4], t.DataLength)
	copy(buf[4:], t.Data)
	return buf, nil
}

func (t *TLV) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return io.ErrUnexpectedEOF
	}
	t.Type = binary.BigEndian.Uint16(data[:2])
	t.DataLength = binary.BigEndian.Uint16(data[2:4])
	if len(data) < 4+int(t.DataLength) {
		return io.ErrUnexpectedEOF
	}
	t.Data = make([]byte, int(t.DataLength))
	copy(t.Data, data[4:4+int(t.DataLength)])
	return nil
}

func (t *TLV) String() string {
	return fmt.Sprintf("TLV(%#x):\n%s", t.Type, util.PrettyBytes(t.Data))
}

func UnmarshalTLVs(data []byte) ([]*TLV, error) {
	tlvs := make([]*TLV, 0)
	d := make([]byte, len(data))
	copy(d, data)

	for len(d) > 0 {
		tlv := &TLV{}
		if err := tlv.UnmarshalBinary(d); err != nil {
			return nil, errors.Wrap(err, "enexpected end to unmarshalling TLVs")
		}
		tlvs = append(tlvs, tlv)
		d = d[tlv.Len():]
	}
	return tlvs, nil
}

func FindTLV(tlvs []*TLV, tlvType uint16) *TLV {
	for _, tlv := range tlvs {
		if tlv.Type == tlvType {
			return tlv
		}
	}
	return nil
}

// type TLVReader struct {
// 	buf []byte
// 	pos int
// }

// func (r *TLVReader) ReadNextTLV() (*TLV, error) {
// 	if len(r.buf) < 4 {
// 		return nil, io.ErrUnexpectedEOF
// 	}

// 	t := &TLV{}
// 	t.Type = Word(r.buf[r.pos:r.pos+2])
// 	r.pos = r.pos + 2
// 	t.DataLength = Word(r.buf[r.pos : r.pos+2])
// 	r.pos = r.pos + 2
// 	copy(p[2:4], r.buf[r.pos:r.pos+2])
// 	r.pos = r.pos + 2

// 	// If there is not enough space to write the expected amount of data, error
// 	if dataLength > len(p)+4 {
// 		return 0, io.ErrUnexpectedEOF
// 	}

// 	n = n + copy(p[4:dataLength], r.buf[r.pos:r.pos+dataLength])
// 	r.pos = r.pos + dataLength
// 	return n, nil
// }

// func NewTLVReader(data []byte) *TLVReader {
// 	return &TLVReader{buf: data, pos: 0}
// }
