package loupedeck

import (
	"encoding/binary"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
)

// Function Listen waits for events from the Loupedeck and calls
// callbacks as configured.
func (l *Loupedeck) Listen() {
	slog.Info("Listening")
	for {
		websocketMsgType, message, err := l.conn.ReadMessage()

		if err != nil {
			slog.Warn("Read error", "error", err)
		}

		if len(message) == 0 {
			slog.Warn("Received a 0-byte message.  Skipping")
			continue
		}

		if websocketMsgType != websocket.BinaryMessage {
			slog.Warn("Unknown websocket message type received", "type", websocketMsgType)
		}

		m, _ := l.parseMessage(message)
		slog.Info("Read", "message", m.String())

		if m.transactionID != 0 {
			if c := l.transactionCallbacks[m.transactionID]; c != nil {
				slog.Info("Callback found, calling")
				c(m)
				l.transactionCallbacks[m.transactionID] = nil
			}
		} else {

			switch m.messageType {
			// Status messages in response to previous commands?
			case SetColor:
			case SetBrightness:
			case SetVibration:
			case Draw:
			case ConfirmFramebuff:
				// nothing
			case Version:
				l.Version = fmt.Sprintf("%d.%d.%d", m.data[0], m.data[1], m.data[2])
				slog.Info("Received 'Version' response", "version", l.Version)
			case Serial:
				l.SerialNo = string(m.data)
				slog.Info("Received 'Serial' response", "serial", l.SerialNo)
			case ButtonPress:
				button := Button(binary.BigEndian.Uint16(message[2:]))
				upDown := ButtonStatus(message[4])
				if upDown == ButtonDown && l.buttonBindings[button] != nil {
					l.buttonBindings[button](button, upDown)
				} else if upDown == ButtonUp && l.buttonUpBindings[button] != nil {
					l.buttonUpBindings[button](button, upDown)
				} else {
					slog.Info("Received uncaught button press message", "button", button, "upDown", upDown, "message", message)
				}
			case KnobRotate:
				knob := Knob(binary.BigEndian.Uint16(message[2:]))
				value := int(message[4])
				if l.knobBindings[knob] != nil {
					v := value
					if value == 255 {
						v = -1
					}
					l.knobBindings[knob](knob, v)
				} else {
					slog.Debug("Received knob rotate message", "knob", knob, "value", value, "message", message)
				}
			case Touch:
				x := binary.BigEndian.Uint16(message[4:])
				y := binary.BigEndian.Uint16(message[6:])
				id := message[8] // Not sure what this is for
				b := touchCoordToButton(x, y)

				if l.touchBindings[b] != nil {
					l.touchBindings[b](b, ButtonDown, x, y)
				} else {
					slog.Debug("Received touch message", "x", x, "y", y, "id", id, "b", b, "message", message)
				}
			case TouchEndCT, TouchEnd:
				x := binary.BigEndian.Uint16(message[4:])
				y := binary.BigEndian.Uint16(message[6:])
				id := message[8] // Not sure what this is for
				b := touchCoordToButton(x, y)

				if l.touchUpBindings[b] != nil {
					l.touchUpBindings[b](b, ButtonUp, x, y)
				} else {
					slog.Debug("Received touch end message", "x", x, "y", y, "id", id, "b", b, "message", message)
				}
			default:
				slog.Info("Received unknown message", "message", m.String())

			}
		}
	}
}
