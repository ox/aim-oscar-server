package oscar

import (
	"aim-oscar/util"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

type HandlerFunc func(*Session, *FLAP)

type Handler struct{ fn HandlerFunc }

func NewHandler(fn HandlerFunc) *Handler {
	return &Handler{
		fn: fn,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	session := NewSession(conn)
	buf := make([]byte, 1024)
	for {
		if !session.GreetedClient {
			// send a hello
			hello := NewFLAP(1)
			hello.Data.Write([]byte{0, 0, 0, 1})
			err := session.Send(hello)
			util.PanicIfError(err)
			session.GreetedClient = true
		}

		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("Read Error: ", err.Error())
			return
		}

		if n == 0 {
			return
		}

		// Try to parse all of the FLAPs in the buffer if we have enough bytes to
		// fill a FLAP header
		for len(buf) >= 6 && buf[0] == 0x2a {
			dataLength := util.Word(buf[4:6])
			flapLength := int(dataLength) + 6
			if len(buf) < flapLength {
				log.Printf("not enough data, only %d bytes\n", len(buf))
				break
			}

			flap := &FLAP{}
			if err := flap.UnmarshalBinary(buf[:flapLength]); err != nil {
				util.PanicIfError(errors.Wrap(err, "could not unmarshal FLAP"))
			}
			buf = buf[flapLength:]
			fmt.Printf("%v ->\n%+v\n", conn.RemoteAddr(), flap)
			h.fn(session, flap)
		}
	}
}
