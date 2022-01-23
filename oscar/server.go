package oscar

import (
	"aim-oscar/util"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/pkg/errors"
)

type HandlerFunc func(context.Context, *FLAP) context.Context
type HandleCloseFn func(context.Context, *Session)

type Handler struct {
	handle      HandlerFunc
	handleClose HandleCloseFn
}

func NewHandler(fn HandlerFunc, handleClose HandleCloseFn) *Handler {
	return &Handler{
		handle:      fn,
		handleClose: handleClose,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	ctx := NewContextWithSession(context.Background(), conn)
	session, _ := SessionFromContext(ctx)

	buf := make([]byte, 2048)
	for {
		if !session.GreetedClient {
			// send a hello
			hello := NewFLAP(1)
			hello.Data.Write([]byte{0, 0, 0, 1})
			session.Send(hello)
			session.GreetedClient = true
		}

		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "use of closed network connection") {
				session.Disconnect()
				h.handleClose(ctx, session)
				return
			}

			log.Println("OSCAR Read Error: ", err.Error())
			return
		}

		if n == 0 {
			return
		}

		// Try to parse all of the FLAPs in the buffer if we have enough bytes to
		// fill a FLAP header
		for len(buf) >= 6 && buf[0] == 0x2a {
			dataLength := binary.BigEndian.Uint16(buf[4:6])
			flapLength := int(dataLength) + 6
			if len(buf) < flapLength {
				log.Printf("not enough data, only %d bytes\n", len(buf))
				fmt.Printf("%s\n", util.PrettyBytes(buf))
				break
			}

			flap := &FLAP{}
			if err := flap.UnmarshalBinary(buf[:flapLength]); err != nil {
				util.PanicIfError(errors.Wrap(err, "could not unmarshal FLAP"))
			}
			buf = buf[flapLength:]
			ctx = h.handle(ctx, flap)
		}
	}
}
