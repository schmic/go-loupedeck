package loupedeck

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType is a uint16 used to identify various commands and
// actions needed for the Loupedeck protocod.
type MessageType byte

// See 'COMMANDS' in https://github.com/foxxyz/loupedeck/blob/master/constants.js
const (
	ButtonPress    MessageType = 0x00
	KnobRotate     MessageType = 0x01
	SetColor       MessageType = 0x02
	Serial         MessageType = 0x03
	Reset          MessageType = 0x06
	Version        MessageType = 0x07
	SetBrightness  MessageType = 0x09
	MCU            MessageType = 0x0d
	Draw           MessageType = 0x0f
	WriteFramebuff MessageType = 0x10
	SetVibration   MessageType = 0x1b
	Touch          MessageType = 0x4d
	TouchCT        MessageType = 0x52
	TouchEnd       MessageType = 0x6d
	TouchEndCT     MessageType = 0x72
)

// Message defines a message for communicating with the Loupedeck
// over USB.  All communication with the Loupedeck occurs via
// Messages, but most application software can use higher-level
// functions in this library and never touch messages directly.
type Message struct {
	messageType   MessageType
	transactionID byte
	length        byte
	data          []byte
}

// NewMessage creates a new low-level Loupedeck message with
// a specified type and data.  This isn't generally needed for
// end-use.
func (d *Device) NewMessage(messageType MessageType, data []byte) *Message {
	dataLen := float64(len(data))
	length := math.Min(dataLen+3, 255)

	m := Message{
		transactionID: d.newTransactionID(),
		messageType:   messageType,
		length:        byte(length),
		data:          data,
	}

	return &m
}

// ParseMessage creates a Loupedeck Message from a block of
// bytes.  This is used to decode incoming messages from a Loupedeck,
// and shouldn't generally be needed outside of this library.
func (d *Device) ParseMessage(b []byte) (*Message, error) {
	m := Message{
		length:        b[0],
		messageType:   MessageType(b[1]),
		transactionID: b[2],
		data:          b[3:],
	}
	return &m, nil
}

// function asBytes() returns the wire-format form of the message.
func (m *Message) asBytes() []byte {
	b := make([]byte, 3)
	b[0] = m.length
	b[1] = byte(m.messageType)
	b[2] = m.transactionID
	b = append(b, m.data...)

	return b
}

// function String() returns a human-readable form of the message for
// debugging use.
func (m *Message) String() string {
	d := m.data

	if len(d) > 16 {
		d = d[0:16]
		return fmt.Sprintf("{len: %d, type: %02x, txn: %02x, data: %v..., actual_len: %d}", m.length, m.messageType, m.transactionID, d, len(m.data))
	}
	return fmt.Sprintf("{len: %d, type: %02x, txn: %02x, data: %v}", m.length, m.messageType, m.transactionID, d)
}

// newTransactionId picks the next 8-bit transaction ID
// number.  This is used as part of the Loupedeck protocol and used to
// match results with specific queries.  The transaction ID
// incrememnts per call and rolls over back to 1 (not 0).
func (d *Device) newTransactionID() uint8 {
	d.transactionMutex.Lock()
	t := d.transactionID
	t++
	if t == 0 {
		t = 1
	}
	d.transactionID = t
	d.transactionMutex.Unlock()

	return t
}

// Send sends a message to the specified device.
func (d *Device) Send(m *Message) error {
	// slog.Info("Sending", "message", m.String())
	d.transactionCallbacks[m.transactionID] = nil

	return d.send(m)
}

// SendWithCallback sends a message to the specified device
// and registers a callback.  When (or if) the Loupedeck sends a
// response to the message, the callback function will be called and
// provided with the response message.
func (d *Device) SendWithCallback(m *Message, c transactionCallback) error {
	slog.Info("Setting callback", "message", m.String())
	d.transactionCallbacks[m.transactionID] = c

	return d.send(m)
}

// SendAndWait sends a message and then waits for a response, returning the response message.
func (d *Device) SendAndWait(m *Message, timeout time.Duration) (*Message, error) {
	ch := make(chan *Message)
	defer close(ch)
	// TODO(scottlaird): actually implement the timeout.
	err := d.SendWithCallback(m, func(m2 *Message) {
		defer func() {
			_ = recover()
		}()
		slog.Info("sendAndWait callback received, sending to channel")
		ch <- m2
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send: %v", err)
	}

	// Trying SendAndWait with Draw() usually fails, because it
	// doesn't get a response back until after the following
	// message has been sent.  Trying to figure out if this is a
	// weird Loupedeck thing or a protocol issue or what.  Try
	// sending a ping over WS, just to see if anything shakes
	// loose.

	//	slog.Info("Sending ping.")
	//	err = d.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
	//	slog.Info("Ping send", "err", err)

	select {
	case resp := <-ch:
		slog.Info("sendAndWait received ok")
		return resp, nil
	case <-time.After(timeout):
		slog.Warn("sendAndWait timeout")
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// send sends a message to the specified device.
func (d *Device) send(m *Message) error {
	b := m.asBytes()
	return d.conn.WriteMessage(websocket.BinaryMessage, b)
}
