# About

wolrepeater is service to send Wake-On-LAN packets across networks/subnets.

*Warning:* The server running wolrepeater must have network interfaces connected to both the `-listen-addr` and `-dest-addr` networks.

# Install

```bash
GOBIN=`pwd` go install github.com/korylprince/wolrepeater/cmd/wolrepeater@latest
```
# Usage

```bash
./wolrepeater -h
Usage of ./wolrepeater:
  -dest-addr string
    	Address to send WOL packets to. Required.
  -dest-mac string
    	MAC Address to send WOL packets for. Defaults to value passed to -listen-mac.
  -dest-port int
    	Port to send WOL packets to. (default 9)
  -listen-addr string
    	Address to listen for WOL packets on. Required.
  -listen-mac string
    	MAC Address to listen for WOL packets on. Required.
  -listen-port int
    	Port to listen for WOL packets on. (default 9)
  -log-file string
    	Location of log file. Defaults to stdout (default "-")
  -log-level string
    	Logging level. (default "info")
```

```bash
# repeat WOL packets for 12:34:56:78:9a:bc from 10.0.0.255 network to 10.0.1.255
sudo ./wolrepeater -listen-addr 10.0.0.255 -listen-mac 12:34:56:78:9a:bc -dest-addr 10.0.1.255
{"time":"2025-08-29T15:56:25.354329-05:00","level":"INFO","msg":"starting listener","listen-addr":"10.0.0.255","listen-port":9,"listen-mac":"12:34:56:78:9a:bc"}
{"time":"2025-08-29T15:56:30.243388-05:00","level":"INFO","msg":"WOL packet received","mac-address":"12:34:56:78:9a:bc"}
{"time":"2025-08-29T15:56:30.244118-05:00","level":"INFO","msg":"WOL packet sent","dest-addr":"10.0.1.255","dest-port":9,"dest-mac":"12:34:56:78:9a:bc"}
...
```
