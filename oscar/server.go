package oscar

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

type Handler struct {
	services map[uint16]Service
}

func NewHandler() *Handler {
	return &Handler{
		services: make(map[uint16]Service),
	}
}

func (h *Handler) RegisterService(family uint16, service Service) {
	h.services[family] = service
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
			panicIfError(err)
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
			dataLength := Word(buf[4:6])
			flapLength := int(dataLength) + 6
			if len(buf) < flapLength {
				log.Printf("not enough data, only %d bytes\n", len(buf))
				break
			}

			flap := &FLAP{}
			if err := flap.UnmarshalBinary(buf[:flapLength]); err != nil {
				panicIfError(errors.Wrap(err, "could not unmarshal FLAP"))
			}
			buf = buf[flapLength:]
			fmt.Printf("%v ->\n%+v\n", conn.RemoteAddr(), flap)

			if flap.Header.Channel == 1 {

			} else if flap.Header.Channel == 2 {
				snac := &SNAC{}
				err := snac.UnmarshalBinary(flap.Data.Bytes())
				panicIfError(err)

				fmt.Printf("%+v\n", snac)
				if tlvs, err := UnmarshalTLVs(snac.Data.Bytes()); err == nil {
					for _, tlv := range tlvs {
						fmt.Printf("%+v\n", tlv)
					}
				} else {
					fmt.Printf("%s\n\n", prettyBytes(snac.Data.Bytes()))
				}

				if service, ok := h.services[snac.Header.Family]; ok {
					err = service.HandleSNAC(session, snac)
					panicIfError(err)
				}
			}
		}
	}
}
