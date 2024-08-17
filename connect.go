package loupedeck

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func ConnectWebsocket(s *SerialWebSockConn, d *Device) error {
	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			slog.Info("Dialing...")
			return s, nil
		},
		HandshakeTimeout: 1 * time.Second,
	}

	header := http.Header{}

	slog.Info("Attempting to open websocket connection")
	conn, resp, err := dialer.Dial("ws://fake", header)
	if err != nil {
		slog.Warn("dial failed", "err", err)
		return err
	}
	d.serial = s
	d.conn = conn

	slog.Info("Connect successful", "resp", resp)
	slog.Info("Found Loupedeck", "vendor", d.Vendor, "product", d.Product)

	err = d.Reset()
	if err != nil {
		panic("unable to reset the device, aborting ...")
	}

	// Ask the device about itself.  The responses come back
	// asynchronously, so we need to provide a callback.  Since
	// `listen()` hasn't been called yet, we *have* to use
	// callbacks, blocking via 'sendAndWait' isn't going to work.
	data := make([]byte, 0)
	m := d.NewMessage(Version, data)
	err = d.SendWithCallback(m, func(m *Message) {
		d.Version = fmt.Sprintf("%d.%d.%d", m.data[0], m.data[1], m.data[2])
		slog.Info("Received 'Version' response", "version", d.Version)
	})
	if err != nil {
		return fmt.Errorf("unable to send: %v", err)
	}

	m = d.NewMessage(Serial, data)
	err = d.SendWithCallback(m, func(m *Message) {
		d.SerialNo = string(m.data)
		slog.Info("Received 'Serial' response", "serial", d.SerialNo)
	})
	if err != nil {
		return fmt.Errorf("unable to send: %v", err)
	}

	err = d.SetDefaultFont()
	if err != nil {
		return fmt.Errorf("unable to set default font: %v", err)
	}

	err = d.SetDefaultBrightness()
	if err != nil {
		return fmt.Errorf("unable to set default brightness: %v", err)
	}

	return nil
}

// ConnectSerialAuto connects to the first compatible Loupedeck in the
// system.  To connect to a specific Loupedeck, use ConnectSerialPath.
func ConnectSerialAuto() (*SerialWebSockConn, error) {
	slog.Info("Enumerating ports")

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no serial ports found")
	}

	for _, port := range ports {
		slog.Info("Trying to open port", "port", port.Name)
		if port.IsUSB && (port.VID == "2ec2" || port.VID == "1532") {
			p, err := serial.Open(port.Name, &serial.Mode{})
			if err != nil {
				return nil, fmt.Errorf("unable to open port %q", port.Name)
			}
			conn := &SerialWebSockConn{
				Name:    port.Name,
				Port:    p,
				Vendor:  port.VID,
				Product: port.PID,
			}
			return conn, nil
		}
	}

	return nil, fmt.Errorf("no Loupedeck devices found")
}

// ConnectSerialPath connects to a specific Loupedeck, using the path
// to the USB serial device as a key.
func ConnectSerialPath(serialPath string) (*SerialWebSockConn, error) {
	p, err := serial.Open(serialPath, &serial.Mode{})
	if err != nil {
		return nil, fmt.Errorf("unable to open serial device %q", serialPath)
	}
	conn := &SerialWebSockConn{
		Name: serialPath,
		Port: p,
	}

	return conn, nil
}

// type connectResult struct {
// 	l   *Loupedeck
// 	err error
// }

// ConnectAuto connects to a Loupedeck Live by automatically locating
// the first USB Loupedeck device in the system.  If you have more
// than one device and want to connect to a specific one, then use
// ConnectPath().
// func ConnectAuto() (*Loupedeck, error) {
// 	c, err := ConnectSerialAuto()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return tryConnect(c)
// }

// ConnectPath connects to a Loupedeck Live via a specified serial
// device.  If successful it returns a new Loupedeck.
// func ConnectPath(serialPath string) (*Loupedeck, error) {
// 	c, err := ConnectSerialPath(serialPath)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return tryConnect(c)
// }

// tryConnect helps make connections to USB devices more reliable by
// adding timeout and retry logic.
//
// Without this, 50% of the time my LoupeDeck fails to connect the
// HTTP link for the websocket.  We send the HTTP headers to request a
// websocket connection, but the LoupeDeck never returns.
//
// This is a painful workaround for that.  It uses the generic Go
// pattern for implementing a timeout (do the "real work" in a
// goroutine, feeding answers to a channel, and then add a timeout on
// select).  If the timeout triggers, then it tries a second time to
// connect.  This has a 100% success rate for me.
//
// The actual connection logic is all in doConnect(), below.
// func tryConnect(c *SerialWebSockConn) (*Loupedeck, error) {
// 	result := make(chan connectResult, 1)
// 	go func() {
// 		r := connectResult{}
// 		r.l, r.err = ConnectWebsocket(c)
// 		result <- r
// 	}()
//
// 	select {
// 	case <-time.After(2 * time.Second):
// 		// timeout
// 		slog.Info("Timeout! Trying again without timeout.")
// 		return ConnectWebsocket(c)
//
// 	case result := <-result:
// 		return result.l, result.err
// 	}
// }
