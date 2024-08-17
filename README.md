# loupedeck

This provides somewhat minimal support for talking to a [Loupedeck
Live](https://loupedeck.com/us/products/loupedeck-live/) from Go.
Supported features:

- Reacting to button, knob, and touchscreen events.
- Displaying images on any of the 3 displays.

This code talks _directly_ to the Loupedeck hardware, and doesn't go
through Loupedeck's Windows/Mac software. It's largely intended for
using the Loupedeck as a controller for semi-embedded devices, like a
DMX lighting controller. This code _should_ work under
Linux/Windows/Mac/etc, although only Linux on a Raspberry Pi has been
tested. It talks to the Loupedeck via the
[`go.bug.st/serial`](http://go.bug.st/serial) library ; any platform
with working USB serial support in the library will likely work just
fine.

This is only tested with a Loupedeck Live; other Loupedeck models use
the same protocol but have different numbers of displays and controls,
and will need minor updates to work correctly.

## Sample code

```
	l, err := loupedeck.ConnectAuto()
	if err != nil { ... }

	// Create 3 variables for holding dial positions, and add a callback for whenever they change.
	light1 := loupedeck.NewWatchedInt(0)
	light1.AddWatcher(func (i int) { fmt.Printf("DMX 1->%d\n", i) })
	light2 := loupedeck.NewWatchedInt(0)
	light2.AddWatcher(func (i int) { fmt.Printf("DMX 3->%d\n", i) })
	light3 := loupedeck.NewWatchedInt(0)
	light3.AddWatcher(func (i int) { fmt.Printf("DMX 5->%d\n", i) })

        // Use the left display and the 3 left knobs to adjust 3 independent lights between 0 and 100.
	// Whenever these change, the callbacks from 'AddWatcher' (above) will be called.
	d.NewTouchDial(loupedeck.DisplayLeft, light1, light2, light3, 0, 100)

	// Define the 'Circle' button (bottom left) to function as an "off" button for lights 1-3.
	// Similar to NewTouchDial, the callbacks from `AddWatcher` will be called.  This
	// includes an implicit call to the TouchDial's Draw() function, so just calling 'Set'
	// will update the values, the lights (if the callbacks above actually did anything useful),
	// and the Loupedeck.

	d.BindButton(loupedeck.Circle, func (b loupedeck.Button, s loupedeck.ButtonStatus){
		light1.Set(0)
		light2.Set(0)
		light3.Set(0)
	})

	d.Listen()
```

## Disclaimer

This is not an official Google project.
