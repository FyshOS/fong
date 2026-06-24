package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	initAudio()

	w := a.NewWindow("Fong")
	w.SetPadded(false)

	g := newGame()
	w.SetContent(g.field)
	w.Resize(fyne.NewSize(800, 500))

	g.connectInput(w)

	// pause automatically when the app is sent to the background
	a.Lifecycle().SetOnExitedForeground(func() {
		fyne.Do(func() { g.setPaused(true) })
	})

	g.start()

	w.ShowAndRun()
}
