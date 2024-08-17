package loupedeck

// Button represents a physical button on the Loupedeck Live.  This
// includes the 8 buttons at the bottom of the device as well as the
// 'click' function of the 6 dials.
type Button uint16

// ButtonState represents the state of Buttons.
type ButtonState uint8

const (
	KnobButton1 Button = 1
	KnobButton2 Button = 2
	KnobButton3 Button = 3
	KnobButton4 Button = 4
	KnobButton5 Button = 5
	KnobButton6 Button = 6

	ButtonCircle Button = 7
	Button0      Button = 7
	Button1      Button = 8
	Button2      Button = 9
	Button3      Button = 10
	Button4      Button = 11
	Button5      Button = 12
	Button6      Button = 13
	Button7      Button = 14

	// CT-specific buttons.
	CTCircle Button = 15
	Undo     Button = 16
	Keyboard Button = 17
	Enter    Button = 18
	Save     Button = 19
	LeftFn   Button = 20
	Up       Button = 21
	A        Button = 21
	Left     Button = 22
	C        Button = 22
	RightFn  Button = 23
	Down     Button = 24
	B        Button = 24
	Right    Button = 25
	D        Button = 25
	E        Button = 26
)

const (
	ButtonDown ButtonState = 0
	ButtonUp   ButtonState = 1
)

// ButtonFunc is a function signature used for callbacks on Button
// events.  When a specified event happens, the ButtonFunc is called
// with parameters specifying which button was pushed and what its
// current state is.
type ButtonFunc func(Button, ButtonState)

// BindButton sets a callback for actions on a specific
// button.  When the Button is pushed down, then the provided
// ButtonFunc is called.
func (d *Device) BindButton(b Button, f ButtonFunc) {
	d.buttonBindings[b] = f
}

// BindButtonUp sets a callback for actions on a specific
// button.  When the Button is released, then the provided
// ButtonFunc is called.
func (d *Device) BindButtonUp(b Button, f ButtonFunc) {
	d.buttonUpBindings[b] = f
}
