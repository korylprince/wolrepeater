package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/korylprince/wolrepeater"
)

func run() error {
	// parse flags
	listenAddrStr := flag.String("listen-addr", "", "Address to listen for WOL packets on. Defaults to all addresses.")
	listenPort := flag.Int("listen-port", 9, "Port to listen for WOL packets on.")
	listenMACStr := flag.String("listen-mac", "", "MAC Address to listen for WOL packets on. Required.")
	destAddrStr := flag.String("dest-addr", "", "Address to send WOL packets to. Required.")
	destPort := flag.Int("dest-port", 9, "Port to send WOL packets to.")
	destMACStr := flag.String("dest-mac", "", "MAC Address to send WOL packets for. Defaults to value passed to -listen-mac.")
	logFileStr := flag.String("log-file", "-", "Location of log file. Defaults to stdout")
	logLevelStr := flag.String("log-level", "info", "Logging level.")
	flag.Parse()

	// parse and validate flag values
	var listenAddr net.IP
	if *listenAddrStr != "" {
		listenAddr = net.ParseIP(*listenAddrStr)
		if listenAddr == nil {
			return fmt.Errorf("could not parse -listen-addr %q as IP", *listenAddrStr)
		}
	}

	if *listenPort == 0 {
		return errors.New("-listen-port cannot be 0")
	}

	var listenMAC net.HardwareAddr
	var err error
	if *listenMACStr != "" {
		listenMAC, err = net.ParseMAC(*listenMACStr)
		if err != nil {
			return fmt.Errorf("could not parse -listen-mac %q: %w", *listenMACStr, err)
		}
	} else {
		return errors.New("-listen-mac is required")
	}

	var destAddr net.IP
	if *destAddrStr != "" {
		destAddr = net.ParseIP(*destAddrStr)
		if destAddr == nil {
			return fmt.Errorf("could not parse -dest-addr %q as IP", *destAddrStr)
		}
	} else {
		return errors.New("-dest-addr is required")
	}

	if *destPort == 0 {
		return errors.New("-dest-port cannot be 0")
	}

	var destMAC net.HardwareAddr
	if *destMACStr != "" {
		destMAC, err = net.ParseMAC(*destMACStr)
		if err != nil {
			return fmt.Errorf("could not parse -dest-mac %q: %w", *destMACStr, err)
		}
	} else {
		destMAC = listenMAC
	}

	logLevel := new(slog.Level)
	if err := logLevel.UnmarshalText([]byte(*logLevelStr)); err != nil {
		return fmt.Errorf("could not parse -log-level %q: %w", *logLevelStr, err)
	}

	// open log file
	logFileHandle := os.Stdout
	if *logFileStr != "-" {
		logFile, err := os.OpenFile(*logFileStr, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("could not open log file %q: %w", *logFileStr, err)
		}
		defer logFile.Close()
	}

	// sanity checks to prevent loopbacks
	if bytes.Equal(listenMAC, destMAC) {
		if listenAddr == nil && listenPort == destPort {
			return fmt.Errorf("listener (*:%d) and destination (%s:%d) will cause a loopback when listen and destination MACs are the same",
				listenPort, destAddr, destPort)
		} else if listenAddr.Equal(destAddr) && listenPort == destPort {
			return fmt.Errorf("listener (%s:%d) and destination (%s:%d) will cause a loopback when listen and destination MACs are the same",
				listenAddr, listenPort, destAddr, destPort)
		}
	}

	// create and run repeater
	r := &wolrepeater.Repeater{
		ListenAddr:         listenAddr,
		ListenPort:         *listenPort,
		ListenHardwareAddr: listenMAC,
		DestAddr:           destAddr,
		DestPort:           *destPort,
		DestHardwareAddr:   destMAC,

		Logger: slog.New(slog.NewJSONHandler(logFileHandle, &slog.HandlerOptions{Level: logLevel})),
	}

	return r.Listen()
}

func main() {
	if err := run(); err != nil {
		flag.Usage()
		fmt.Println(err)
	}
}
