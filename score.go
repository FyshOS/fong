package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// digitFont is a 3x5 pixel bitmap for each digit, in the chunky style of the
// original's score readout. '1' marks a lit pixel.
var digitFont = map[rune][5]string{
	'0': {"111", "101", "101", "101", "111"},
	'1': {"010", "110", "010", "010", "111"},
	'2': {"111", "001", "111", "100", "111"},
	'3': {"111", "001", "111", "001", "111"},
	'4': {"101", "101", "111", "001", "001"},
	'5': {"111", "100", "111", "001", "111"},
	'6': {"111", "100", "111", "101", "111"},
	'7': {"111", "001", "010", "010", "010"},
	'8': {"111", "101", "111", "101", "111"},
	'9': {"111", "101", "111", "001", "111"},
}

const (
	digitCols = 3
	digitRows = 5
	digitGap  = 1 // blank columns between digits
)

// scoreboard renders a number as scaled-up pixels in a positionable container.
type scoreboard struct {
	cont  *fyne.Container
	col   color.Color
	pixel float32
}

func newScoreboard(col color.Color, pixel float32) *scoreboard {
	return &scoreboard{cont: container.NewWithoutLayout(), col: col, pixel: pixel}
}

// set rebuilds the pixels to show n.
func (s *scoreboard) set(n int) {
	objs := []fyne.CanvasObject{}
	x := float32(0)
	for _, r := range strconv.Itoa(n) {
		glyph := digitFont[r]
		for row := 0; row < digitRows; row++ {
			for col := 0; col < digitCols; col++ {
				if glyph[row][col] != '1' {
					continue
				}
				px := canvas.NewRectangle(s.col)
				px.Resize(fyne.NewSize(s.pixel, s.pixel))
				px.Move(fyne.NewPos(x+float32(col)*s.pixel, float32(row)*s.pixel))
				objs = append(objs, px)
			}
		}
		x += float32(digitCols+digitGap) * s.pixel
	}
	s.cont.Objects = objs
	s.cont.Refresh()
}

// size is the pixel footprint of the current contents.
func (s *scoreboard) size(n int) fyne.Size {
	digits := len(strconv.Itoa(n))
	w := float32(digits*(digitCols+digitGap)-digitGap) * s.pixel
	return fyne.NewSize(w, digitRows*s.pixel)
}
