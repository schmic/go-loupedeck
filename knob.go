package loupedeck

// Knob represents the 6 knobs on the Loupedeck Live.
type Knob uint16

const (
	// CTKnob is the large knob in the center of the Loupedeck CT
	CTKnob Knob = 0
	// Knob1 is the upper left knob.
	Knob1 Knob = 1
	// Knob2 is the middle left knob.
	Knob2 Knob = 2
	// Knob3 is the bottom left knob.
	Knob3 Knob = 3
	// Knob4 is the upper right knob.
	Knob4 Knob = 4
	// Knob5 is the middle right knob.
	Knob5 Knob = 5
	// Knob6 is the bottom right knob.
	Knob6 Knob = 6
)

// KnobFunc is a function signature used for callbacks on Knob events,
// similar to ButtonFunc's use with Button events.  The exact use of
// the second parameter depends on the use; in some cases it's simply
// +1/-1 (for right/left button turns) and in other cases it's the
// current value of the dial.
type KnobFunc func(Knob, int)

// BindKnob sets a callback for actions on a specific
// knob.  When the Knob is turned then the provided
// KnobFunc is called.
func (d *Device) BindKnob(k Knob, f KnobFunc) {
	d.knobBindings[k] = f
}
