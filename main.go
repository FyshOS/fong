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
	g.start()

	w.ShowAndRun()
}
