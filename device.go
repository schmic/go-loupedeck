package loupedeck

import (
	"image"
	"image/color"
	"image/draw"
	"log/slog"

	"github.com/gorilla/websocket"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"sync"
)

type transactionCallback func(m *Message)

// Device describes a Device device.
type Device struct {
	Vendor               string
	Product              string
	Model                string
	Version              string
	SerialNo             string
	displays             map[string]*Display
	font                 *opentype.Font
	face                 font.Face
	fontdrawer           *font.Drawer
	serial               *SerialWebSockConn
	conn                 *websocket.Conn
	buttonBindings       map[Button]ButtonFunc
	buttonUpBindings     map[Button]ButtonFunc
	knobBindings         map[Knob]KnobFunc
	touchBindings        map[TouchButton]TouchFunc
	touchUpBindings      map[TouchButton]TouchFunc
	transactionID        uint8
	transactionMutex     sync.Mutex
	transactionCallbacks map[byte]transactionCallback
}

func CreateDevice(s *SerialWebSockConn) *Device {
	// TODO: add some tests if Vendor/Product is known

	d := &Device{
		Vendor:               s.Vendor,
		Product:              s.Product,
		buttonBindings:       make(map[Button]ButtonFunc),
		buttonUpBindings:     make(map[Button]ButtonFunc),
		knobBindings:         make(map[Knob]KnobFunc),
		touchBindings:        make(map[TouchButton]TouchFunc),
		touchUpBindings:      make(map[TouchButton]TouchFunc),
		transactionCallbacks: map[byte]transactionCallback{},
		displays:             map[string]*Display{},
	}

	d.SetDisplays()

	return d
}

// Close closes the connection to the Loupedeck.
func (d *Device) Close() {
	slog.Info("Closing connections")
	d.conn.Close()
	d.serial.Close()
}

// FontDrawer returns a font.Drawer object configured to
// writing text onto the Loupedeck's graphical buttons.
func (d *Device) FontDrawer() font.Drawer {
	return font.Drawer{
		Src:  d.fontdrawer.Src,
		Face: d.face,
	}
}

// Face returns the current font.Face in use for writing text
// onto the Loupedeck's graphical buttons.
func (d *Device) Face() font.Face {
	return d.face
}

// TextInBox writes a specified string into a x,y pixel
// image.Image, using the specified foreground and background colors.
// The font size used will be chosen to maximize the size of the text.
func (d *Device) TextInBox(x, y int, s string, fg, bg color.Color) (image.Image, error) {
	im := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	fd := d.FontDrawer()
	fd.Src = &image.Uniform{fg}
	fd.Dst = im

	size := 12.0
	x26 := fixed.I(x)
	y26 := fixed.I(y)

	mx26 := fixed.I(int(float64(x) * 0.85))
	my26 := fixed.I(int(float64(y) * 0.85))

	for {
		face, err := opentype.NewFace(d.font, &opentype.FaceOptions{
			Size: size,
			DPI:  150,
		})
		if err != nil {
			return nil, err
		}

		fd.Face = face

		bounds, _ := fd.BoundString(s)
		//fmt.Printf("Measured %q at %+v\n", s, bounds)
		width := bounds.Max.X - bounds.Min.X
		height := bounds.Max.Y - bounds.Min.Y

		if width > mx26 || height > my26 {
			size = size * 0.8
			//fmt.Printf("Reducing font size to %f\n", size)
			continue
		}

		centerx := (x26 - width) / 2
		centery := (y26-height)/2 - bounds.Min.Y

		//fmt.Printf("H: %v  H: %v  Center: %v/%v\n", height, width, centerx, centery)

		fd.Dot = fixed.Point26_6{X: centerx, Y: centery}
		fd.DrawString(s)
		return im, nil
	}

}

// SetDefaultFont sets the default font for drawing onto buttons.
//
// TODO(laird): Actually make it easy to override this default.
func (d *Device) SetDefaultFont() error {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return err
	}
	d.font = f

	d.face, err = opentype.NewFace(d.font, &opentype.FaceOptions{
		Size: 12,
		DPI:  150,
	})
	if err != nil {
		return err
	}

	d.fontdrawer = &font.Drawer{
		Src:  &image.Uniform{color.RGBA{255, 255, 255, 255}},
		Face: d.face,
	}
	return nil
}

// SetBrightness sets the overall brightness of the Loupedeck display.
func (d *Device) SetBrightness(b int) error {
	data := make([]byte, 1)
	data[0] = byte(b)
	m := d.NewMessage(SetBrightness, data)
	return d.Send(m)
}

func (d *Device) SetDefaultBrightness() error {
	data := make([]byte, 9)
	m := d.NewMessage(SetBrightness, data)
	return d.Send(m)
}

// SetButtonColor sets the color of a specific Button.  The
// Loupedeck Live allows the 8 buttons below the display to be set to
// specific colors, however the 'Circle' button's colors may be
// overridden to show the status of the Loupedeck Live's connection to
// the host.
func (d *Device) SetButtonColor(b Button, c color.RGBA) error {
	data := make([]byte, 4)
	data[0] = byte(b)
	data[1] = c.R
	data[2] = c.G
	data[3] = c.B
	m := d.NewMessage(SetColor, data)
	return d.Send(m)
}

func (d *Device) Reset() error {
	data := make([]byte, 0)
	m := d.NewMessage(Reset, data)
	return d.Send(m)
}

func (d *Device) ClearDisplay() {
	d.GetDisplay("all").Clear()
}
