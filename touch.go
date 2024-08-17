package loupedeck

import (
	"log"
	"time"
)

// TouchButton represents the regions of the touchpad on the Loupedeck Live.
type TouchButton uint16

const (
	Touch1     TouchButton = 1
	Touch2     TouchButton = 2
	Touch3     TouchButton = 3
	Touch4     TouchButton = 4
	Touch5     TouchButton = 5
	Touch6     TouchButton = 6
	Touch7     TouchButton = 7
	Touch8     TouchButton = 8
	Touch9     TouchButton = 9
	Touch10    TouchButton = 10
	Touch11    TouchButton = 11
	Touch12    TouchButton = 12
	TouchLeft  TouchButton = 101 // TouchLeft indicates that the left touchscreen area has been touched
	TouchRight TouchButton = 102 // TouchRight indicates that hte right touchscreen area has been touched
)

// TouchFunc is a function signature used for callbacks on TouchButton
// events, similar to ButtonFunc and KnobFunc.  The parameters are:
//
//   - The TouchButton touched
//   - The ButtonStatus (down/up)
//   - The X location touched (relative to the whole display)
//   - The Y location touched (relative to the whole display)
type TouchFunc func(TouchButton, ButtonState, uint16, uint16)

// BindTouch sets a callback for actions on a specific
// TouchButton.  When the TouchButton is pushed down, then the
// provided TouchFunc is called.
func (d *Device) BindTouch(b TouchButton, f TouchFunc) {
	d.touchBindings[b] = f
}

// BindTouchUp sets a callback for actions on a specific
// TouchButton.  When the TouchButton is released, then the
// provided TouchFunc is called.
func (d *Device) BindTouchUp(b TouchButton, f TouchFunc) {
	d.touchUpBindings[b] = f
}

// Quick hack to decide if a touch is a click or a drag.  In an ideal
// world, we'd also support double-click, but that either requires
// knowing the future or delaying click messages until after a
// specified time has passed without a second click, and there's no
// room in the code for either today.
func isClick(duration time.Duration, x, y int) bool {
	if duration > 500*time.Millisecond {
		return false
	}

	if x > 20 || x < -20 {
		return false
	}

	if y > 20 || y < -20 {
		return false
	}

	return true
}

// touchCoordToButton translates an x,y coordinate on the
// touchscreen to a TouchButton.
func CoordToTouchButton(x, y uint16) TouchButton {
	switch {
	case x < 60:
		return TouchLeft
	case x >= 420:
		return TouchRight
	}

	x -= 60
	x /= 90
	y /= 90

	return TouchButton(uint16(Touch1) + x + 4*y)
}

// touchToXY turns a specific TouchButton into a set of x,y +
// Display addresses, for use with the Draw function.
func (b *TouchButton) ToCoord() (int, int) {
	switch *b {
	case Touch1:
		return 0, 0
	case Touch2:
		return 90, 0
	case Touch3:
		return 180, 0
	case Touch4:
		return 270, 0
	case Touch5:
		return 0, 90
	case Touch6:
		return 90, 90
	case Touch7:
		return 180, 90
	case Touch8:
		return 270, 90
	case Touch9:
		return 0, 180
	case Touch10:
		return 90, 180
	case Touch11:
		return 180, 180
	case Touch12:
		return 270, 180
	}

	log.Fatalln("Unknown TouchButton", b)
	return -1, -1
}
