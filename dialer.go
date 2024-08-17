package loupedeck

import (
	"net"
	"time"

	"go.bug.st/serial"
)

// SerialWebSockConn implements an external dialer interface for the
// Gorilla that allows it to talk to Loupedeck's weird
// websockets-over-serial-over-USB setup.
//
// The Gorilla websockets library can use an external dialer
// interface, which means that we can use it *mostly* unmodified to
// talk to a serial device instead of a network device.  We just need
// to provide something that matches the net.Conn interface.  Here's a
// minimal implementation.
type SerialWebSockConn struct {
	Name            string
	Port            serial.Port
	Vendor, Product string
}

// Read reads bytes from the connected serial port.
func (s *SerialWebSockConn) Read(b []byte) (n int, err error) {
	//slog.Info("Reading", "limit_bytes", len(b))
	n, err = s.Port.Read(b)
	//slog.Info("Read", "bytes", n, "err", err, "data", fmt.Sprintf("%v", b[:n]))
	return n, err
}

// Write sends bytes to the connected serial port.
func (s *SerialWebSockConn) Write(b []byte) (n int, err error) {
	//slog.Info("Writing", "bytes", len(b), "message", fmt.Sprintf("%v", b))
	return s.Port.Write(b)
}

// Close closed the connection.
func (s *SerialWebSockConn) Close() error {
	return nil // d.Port.Close()
}

// LocalAddr is needed for Gorilla compatibility, but doesn't actually
// make sense with serial ports.
func (s *SerialWebSockConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (s *SerialWebSockConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (s *SerialWebSockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (s *SerialWebSockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is needed for Gorilla compatibility, but doesn't
// actually make sense with serial ports.
func (s *SerialWebSockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
