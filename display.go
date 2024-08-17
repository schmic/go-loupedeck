package loupedeck

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"log/slog"

	"maze.io/x/pixel/pixelcolor"
)

type Display struct {
	device    *Device
	id        byte
	width     int
	height    int
	offsetx   int // used for mapping legacy left/center/right screens onto unified devices.
	offsety   int // used for mapping legacy left/center/right screens onto unified devices.
	Name      string
	bigEndian bool
}

func (d *Device) GetDisplay(name string) *Display {
	return d.displays[name]
}

func (d *Device) addDisplay(name string, id byte, width, height, offsetx, offsety int, bigEndian bool) {
	d.displays[name] = &Display{
		device:    d,
		Name:      name,
		id:        id,
		width:     width,
		height:    height,
		offsetx:   offsetx,
		offsety:   offsety,
		bigEndian: bigEndian,
	}
}

func (d *Device) SetDisplays() {
	switch d.Product {
	case "0003":
		slog.Info("Using Loupedeck CT v1 display settings.")
		d.addDisplay("left", 'L', 60, 270, 0, 0, false)
		d.addDisplay("main", 'A', 360, 270, 60, 0, false)
		d.addDisplay("right", 'R', 60, 270, 420, 0, false)
		d.addDisplay("dial", 'W', 240, 240, 0, 0, true)
	case "0007":
		slog.Info("Using Loupedeck CT v2 display settings.")
		d.addDisplay("left", 'M', 60, 270, 0, 0, false)
		d.addDisplay("main", 'M', 360, 270, 60, 0, false)
		d.addDisplay("right", 'M', 60, 270, 420, 0, false)
		d.addDisplay("all", 'M', 480, 270, 0, 0, false)
		d.addDisplay("dial", 'W', 240, 240, 0, 0, true)
	case "0004":
		slog.Info("Using Loupedeck Live display settings.")
		d.addDisplay("left", 'M', 60, 270, 0, 0, false)
		d.addDisplay("main", 'M', 360, 270, 60, 0, false)
		d.addDisplay("right", 'M', 60, 270, 420, 0, false)
		d.addDisplay("all", 'M', 480, 270, 0, 0, false)
	case "0006", "0d06":
		slog.Info("Using Loupedeck Live S/Razor Stream Controller display settings.")
		d.addDisplay("left", 'M', 60, 270, 0, 0, false)
		d.addDisplay("main", 'M', 360, 270, 60, 0, false)
		d.addDisplay("right", 'M', 60, 270, 420, 0, false)
		d.addDisplay("all", 'M', 480, 270, 0, 0, false)
	default:
		panic("Unknown device type: " + d.Product)
	}
}

func (d *Display) Height() int {
	return d.height
}

func (d *Display) Width() int {
	return d.width
}

func (d *Display) Draw(im image.Image, xoff, yoff int) {
	slog.Info("Draw called", "Display", d.Name, "xoff", xoff, "yoff", yoff, "width", im.Bounds().Dx(), "height", im.Bounds().Dy())

	x := xoff + d.offsetx
	y := yoff + d.offsety
	width := im.Bounds().Dx()
	height := im.Bounds().Dy()
	slog.Info("Draw parameters", "x", x, "y", y, "width", width, "height", height)

	// Call 'WriteFramebuff'
	data := make([]byte, 10)
	binary.BigEndian.PutUint16(data[0:], uint16(d.id))
	binary.BigEndian.PutUint16(data[2:], uint16(x))
	binary.BigEndian.PutUint16(data[4:], uint16(y))
	binary.BigEndian.PutUint16(data[6:], uint16(width))
	binary.BigEndian.PutUint16(data[8:], uint16(height))

	b := im.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			pixel := pixelcolor.ToRGB565(im.At(x, y))
			lowByte := byte(pixel & 0xff)
			highByte := byte(pixel >> 8)

			// The Loupedeck CT's center knob screen wants
			// images fed to it big endian; all other
			// displays are little endian.
			if d.bigEndian {
				data = append(data, highByte, lowByte)
			} else {
				data = append(data, lowByte, highByte)
			}
		}
	}

	m := d.device.NewMessage(WriteFramebuff, data)
	err := d.device.Send(m)
	if err != nil {
		slog.Warn("Send failed", "err", err)
	}

	// I'd love to watch the return code for WriteFramebuff, but
	// it doesn't seem to come back until after Draw, below.

	//resp, err := d.loupedeck.SendAndWait(m, 50*time.Millisecond)
	//if err != nil {
	//	slog.Warn("Received error on draw", "message", resp)
	//}

	// Call 'Draw'.  The screen isn't actually updated until
	// 'draw' arrives.  Unclear if we should wait for the previous
	// Framebuffer transaction to complete first, but adding a
	// giant sleep here doesn't seem to change anything.
	//
	// Ideally, we'd batch these and only call Draw when we're
	// doing with multiple FB updates.

	d.Refresh()
}

func (d *Display) Clear() {
	dw := d.width
	dh := d.height
	im := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.Draw(im, im.Bounds(), &image.Uniform{&color.Black}, image.ZP, draw.Src)
	d.Draw(im, 0, 0)
}

func (d *Display) Refresh() {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data[0:], uint16(d.id))
	msg := d.device.NewMessage(Draw, data)
	err := d.device.Send(msg)
	if err != nil {
		slog.Warn("Send failed", "err", err)
	}
}
