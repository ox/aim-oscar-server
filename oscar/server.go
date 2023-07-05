package oscar

import (
	"aim-oscar/util"
	"bytes"
	"context"

	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
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

func (h *Handler) Handle(conn net.Conn, logger *slog.Logger) {
	connLogger := logger.With("session_id", uuid.New(), "ip", conn.RemoteAddr().String())
	connLogger.Info("New Connection")

	ctx := NewContextWithSession(context.Background(), conn, connLogger)
	session, _ := SessionFromContext(ctx)

	var buf bytes.Buffer
	for {
		if !session.GreetedClient {
			// send a hello
			hello := NewFLAP(1)
			hello.Data.Write([]byte{0, 0, 0, 1})
			session.Send(hello)
			session.GreetedClient = true
		}

		// Wait for some data to read
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		incoming := make([]byte, 512)
		n, err := conn.Read(incoming)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "use of closed network connection") {
				session.Disconnect()
				h.handleClose(ctx, session)
				return
			}

			// If the read timed out, just try reading again
			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			connLogger.Error("OSCAR Read Error", "err", err.Error())
			return
		}

		if n == 0 {
			return
		}

		buf.Write(incoming[:n])

		// Try to parse all of the FLAPs in the buffer if we have enough bytes to
		// fill a FLAP header
		for buf.Len() >= 6 && buf.Bytes()[0] == 0x2a {
			bufBytes := buf.Bytes()
			dataLength := binary.BigEndian.Uint16(bufBytes[4:6])
			flapLength := int(dataLength) + 6
			if len(bufBytes) < flapLength {
				connLogger.Error(fmt.Sprintf("not enough data, expected %d bytes but have %d bytes", flapLength, len(bufBytes)))
				fmt.Printf("%s\n", util.PrettyBytes(bufBytes))
				break
			}

			flap := &FLAP{}
			flapBuf := make([]byte, flapLength)
			buf.Read(flapBuf)
			if err := flap.UnmarshalBinary(flapBuf); err != nil {
				connLogger.Error("could not unmarshal FLAP", "err", err)
				// Toss out everything
				buf.Reset()
				break
			}

			ctx = h.handle(ctx, flap)
		}
	}
}
