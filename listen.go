package loupedeck

import (
	"encoding/binary"
	"log/slog"

	"github.com/gorilla/websocket"
)

// Listen waits for events from the Loupedeck and calls
// callbacks as configured.
func (d *Device) Listen() error {
	slog.Info("Listening ...")
	for {
		websocketMsgType, data, err := d.conn.ReadMessage()
		if err != nil {
			slog.Warn("Read error, exiting", "error", err)
			return err
		}

		if len(data) == 0 {
			slog.Warn("Received a 0-byte message.  Skipping")
			continue
		}

		if websocketMsgType != websocket.BinaryMessage {
			slog.Warn("Unknown websocket message type received", "type", websocketMsgType)
			continue
		}

		msg, _ := d.ParseMessage(data)

		if msg.transactionID != 0 {
			if cb := d.transactionCallbacks[msg.transactionID]; cb != nil {
				slog.Info("Callback found with", "txid", msg.transactionID)
				cb(msg)
				d.transactionCallbacks[msg.transactionID] = nil
			}
			continue
		}

		switch msg.messageType {

		case ButtonPress:
			button := Button(binary.BigEndian.Uint16(data[2:]))
			upDown := ButtonState(data[4])

			slog.Info("Received button press message", "button", button, "upDown", upDown, "message", data)

			if upDown == ButtonDown && d.buttonBindings[button] != nil {
				d.buttonBindings[button](button, upDown)
			} else if upDown == ButtonUp && d.buttonUpBindings[button] != nil {
				d.buttonUpBindings[button](button, upDown)
			}

		case KnobRotate:
			knob := Knob(binary.BigEndian.Uint16(data[2:]))
			value := int(data[4])

			slog.Info("Received knob rotate message", "knob", knob, "value", value, "message", data)

			if d.knobBindings[knob] != nil {
				v := value
				if value == 255 {
					v = -1
				}
				d.knobBindings[knob](knob, v)
			}

		case Touch:
			x := binary.BigEndian.Uint16(data[4:])
			y := binary.BigEndian.Uint16(data[6:])
			id := data[8] // Not sure what this is for
			b := CoordToTouchButton(x, y)

			slog.Info("Received touch message", "x", x, "y", y, "id", id, "b", b, "message", data)

			if d.touchBindings[b] != nil {
				d.touchBindings[b](b, ButtonDown, x, y)
			}

		case TouchEnd:
			x := binary.BigEndian.Uint16(data[4:])
			y := binary.BigEndian.Uint16(data[6:])
			id := data[8] // Not sure what this is for
			b := CoordToTouchButton(x, y)

			slog.Info("Received touch end message", "x", x, "y", y, "id", id, "b", b, "message", data)

			if d.touchUpBindings[b] != nil {
				d.touchUpBindings[b](b, ButtonUp, x, y)
			}

		case 0x73:
			// seems to be some websocket information, we ignore it
			// fmt.Printf("%s \n", msg.data)

		default:
			slog.Info("Received unhandled", "message", msg)
			//slog.Info("Received unknown", "message", msg.String())
		}
	}
}
