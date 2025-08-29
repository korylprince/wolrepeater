package wolrepeater

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
)

const (
	wolPacketLen = 102
	wolMACRepeat = 16
	macLen       = 6
)

var magicHeader = []byte{255, 255, 255, 255, 255, 255}

type Repeater struct {
	ListenAddr         net.IP
	ListenPort         int
	ListenHardwareAddr net.HardwareAddr
	DestAddr           net.IP
	DestPort           int
	DestHardwareAddr   net.HardwareAddr

	Logger *slog.Logger
}

func verifyWOLHardwareAddress(buf []byte, mac net.HardwareAddr) bool {
	for i := range wolMACRepeat {
		if !bytes.Equal(buf[macLen*i:macLen*(i+1)], mac) {
			return false
		}
	}
	return true
}

// Listen listens for incoming WOL packets matching the listener and sends WOL packets to the destination
func (r *Repeater) Listen() error {
	r.Logger.Info("starting listener",
		slog.String("listen-addr", r.ListenAddr.String()),
		slog.Int("listen-port", r.ListenPort),
		slog.String("listen-mac", r.ListenHardwareAddr.String()),
	)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: r.ListenAddr, Port: r.ListenPort})
	if err != nil {
		return fmt.Errorf("could not open udp listener: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("could not read udp packet: %w", err)
		}
		// validate correct length for WOL packet
		if n != wolPacketLen {
			r.Logger.Debug("received non-WOL packet", slog.Int("invalid-packet-length", n))
			continue
		}

		// validate packet has WOL header
		if !bytes.Equal(buf[:len(magicHeader)], magicHeader) {
			r.Logger.Debug("received non-WOL packet", slog.Any("invalid-header", buf[:len(magicHeader)]))
			continue
		}

		// check if packet hardware address matches
		pktHardwareAddr := net.HardwareAddr(buf[len(magicHeader) : len(magicHeader)+macLen])
		if !bytes.Equal(pktHardwareAddr, r.ListenHardwareAddr) {
			// if debug logging is enabled check if this is a valid WOL packet for another mac address
			if r.Logger.Handler().Enabled(context.Background(), slog.LevelDebug) {
				if verifyWOLHardwareAddress(buf[len(magicHeader):], pktHardwareAddr) {
					r.Logger.Debug("received WOL packet for another MAC address",
						slog.String("mac-address", pktHardwareAddr.String()),
					)
				} else {
					r.Logger.Debug("received invalid WOL packet")
				}
			}
			continue
		}

		// check hardware address is repeated 16 times
		if !verifyWOLHardwareAddress(buf[len(magicHeader):], r.ListenHardwareAddr) {
			r.Logger.Debug("received invalid WOL packet: incomplete match")
			continue
		}

		r.Logger.Info("WOL packet received", slog.String("mac-address", r.ListenHardwareAddr.String()))

		// send WOL packet to destination
		if err := r.sendWOL(); err != nil {
			r.Logger.Warn("sending WOL packet failed",
				slog.String("dest-addr", r.DestAddr.String()),
				slog.Int("dest-port", r.DestPort),
				slog.String("dest-mac", r.DestHardwareAddr.String()),
				slog.String("send-error", err.Error()),
			)
			continue
		}

		r.Logger.Info("WOL packet sent",
			slog.String("dest-addr", r.DestAddr.String()),
			slog.Int("dest-port", r.DestPort),
			slog.String("dest-mac", r.DestHardwareAddr.String()),
		)
	}
}

func genWOLPacket(mac net.HardwareAddr) []byte {
	pkt := make([]byte, 0, wolPacketLen)
	pkt = append(pkt, magicHeader...)
	for range wolMACRepeat {
		pkt = append(pkt, mac...)
	}
	return pkt
}

func (r *Repeater) sendWOL() error {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: r.DestAddr, Port: r.DestPort})
	if err != nil {
		return fmt.Errorf("could not open UDP conn: %w", err)
	}
	defer conn.Close()

	pkt := genWOLPacket(r.DestHardwareAddr)

	if _, err := conn.Write(pkt); err != nil {
		return fmt.Errorf("could not send packet: %w", err)
	}

	return nil
}
