package main

import (
	"aim-oscar/oscar"
	"aim-oscar/util"
	"context"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/exp/slog"
)

type OSCARLogHandler struct {
	logger    *log.Logger
	level     slog.Level
	attrs     []slog.Attr
	openGroup string
	lock      *sync.Mutex
}

func (h *OSCARLogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.lock.Lock()
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	h.logger.Println(timeStr, level, msg)

	for _, attr := range h.attrs {
		h.logger.Printf("  %s=%s\n", color.YellowString(h.openGroup+attr.Key), color.WhiteString("%v", attr.Value.Any()))
	}

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "flap" {
			return true
		} else {
			h.logger.Printf("  %s=%s\n", color.YellowString(h.openGroup+a.Key), color.WhiteString("%v", a.Value.Any()))
		}

		return true
	})

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "flap" {
			flap := a.Value.Any().(*oscar.FLAP)
			snac := &oscar.SNAC{}
			if err := snac.UnmarshalBinary(flap.Data.Bytes()); err == nil {
				h.logger.Printf("  FLAP(CH:%d, SEQ:%d):", flap.Header.Channel, flap.Header.SequenceNumber)
				h.logger.Printf("    SNAC(%#x, %#x):", snac.Header.Family, snac.Header.Subtype)
				tlvs, err := oscar.UnmarshalTLVs(snac.Data.Bytes())
				if err == nil {
					for _, tlv := range tlvs {
						h.logger.Printf("      TLV(%#x):", tlv.Type)
						tlvLines := strings.Split(util.PrettyBytes(tlv.Data), "\n")
						for _, line := range tlvLines {
							h.logger.Printf("        %s\n", line)
						}
					}
				} else {
					snaclines := strings.Split(util.PrettyBytes(snac.Data.Bytes()), "\n")
					for _, line := range snaclines {
						h.logger.Printf("      %s\n", line)
					}
				}
			} else {
				flapLines := strings.Split(flap.String(), "\n")
				for _, line := range flapLines {
					h.logger.Printf("  %s\n", line)
				}
			}
		}
		return true
	})

	h.lock.Unlock()
	return nil
}

func (h *OSCARLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *OSCARLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OSCARLogHandler{
		attrs:     append(h.attrs, attrs...),
		logger:    h.logger,
		level:     h.level,
		lock:      h.lock,
		openGroup: h.openGroup,
	}
}

func (h *OSCARLogHandler) WithGroup(name string) slog.Handler {
	return &OSCARLogHandler{
		attrs:     h.attrs,
		logger:    h.logger,
		level:     h.level,
		lock:      h.lock,
		openGroup: h.openGroup + name + ".",
	}
}

func NewOSCARLogHandler(
	out io.Writer,
	opts *slog.HandlerOptions,
) *OSCARLogHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	return &OSCARLogHandler{
		level:  opts.Level.Level(),
		logger: log.New(out, "", 0),
		lock:   &sync.Mutex{},
	}
}
