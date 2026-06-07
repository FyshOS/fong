package main

import (
	"math"
	"time"

	"fyne.io/fyne/v2"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

const audioRate beep.SampleRate = 44100

var audioReady bool

// initAudio sets up the speaker once. If it fails the game runs silently.
func initAudio() {
	if err := speaker.Init(audioRate, audioRate.N(time.Second/30)); err != nil {
		fyne.LogError("audio unavailable, running silently", err)
		return
	}
	audioReady = true
}

// tone is a one-shot square-wave streamer, the kind of blip the original used.
type tone struct {
	freq      float64
	amp       float64
	pos       int
	remaining int
}

func (t *tone) Stream(samples [][2]float64) (n int, ok bool) {
	if t.remaining <= 0 {
		return 0, false
	}
	for i := range samples {
		if t.remaining <= 0 {
			return i, true
		}
		phase := math.Mod(float64(t.pos)*t.freq/float64(audioRate), 1)
		v := t.amp
		if phase >= 0.5 {
			v = -t.amp
		}
		samples[i][0] = v
		samples[i][1] = v
		t.pos++
		t.remaining--
	}
	return len(samples), true
}

func (t *tone) Err() error { return nil }

func playTone(freq float64, d time.Duration) {
	if !audioReady {
		return
	}
	speaker.Play(&tone{freq: freq, amp: 0.2, remaining: audioRate.N(d)})
}

// The three classic Pong blips: paddle, wall, and a lower point tone.
func soundPaddle() { playTone(480, 90*time.Millisecond) }
func soundWall()   { playTone(240, 90*time.Millisecond) }
func soundScore()  { playTone(130, 250*time.Millisecond) }
