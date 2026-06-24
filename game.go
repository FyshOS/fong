package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

const (
	paddleWidth  = 14
	paddleHeight = 90
	paddleInset  = 24
	ballSize     = 14
	paddleSpeed  = 7
	scoreHeight  = 70
)

type game struct {
	field *fyne.Container

	bg       *canvas.Rectangle
	net      []*canvas.Rectangle
	paddle1 *canvas.Rectangle
	paddle2 *canvas.Rectangle
	ball    *canvas.Rectangle
	board1  *scoreboard
	board2  *scoreboard

	// pause overlay: a dim wash and the "PAUSED" caption
	dim       *canvas.Rectangle
	pauseText *canvas.Text
	paused    bool

	// foreground colour for the greyscale pieces
	fg color.Color

	// logical state (top-left of paddles, centre of ball)
	p1y, p2y       float32
	ballX, ballY   float32
	velX, velY     float32
	score1, score2 int
	serveDir       float32
	serveWait      int // frames to hold before the ball launches

	up1, down1, up2, down2 bool
}

func newGame() *game {
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	var bgCol, fgCol color.Color
	if variant == theme.VariantLight {
		bgCol = color.NRGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0xff}
		fgCol = color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}
	} else {
		bgCol = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
		fgCol = color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff}
	}

	g := &game{fg: fgCol, serveDir: 1}

	g.bg = canvas.NewRectangle(bgCol)
	g.paddle1 = canvas.NewRectangle(fgCol)
	g.paddle2 = canvas.NewRectangle(fgCol)
	g.ball = canvas.NewRectangle(fgCol)

	g.board1 = newScoreboard(fgCol, 9)
	g.board1.set(0)
	g.board2 = newScoreboard(fgCol, 9)
	g.board2.set(0)

	objs := []fyne.CanvasObject{g.bg}
	// dashed centre net
	for i := 0; i < 16; i++ {
		dash := canvas.NewRectangle(fgCol)
		g.net = append(g.net, dash)
		objs = append(objs, dash)
	}
	objs = append(objs, g.paddle1, g.paddle2, g.ball, g.board1.cont, g.board2.cont)

	// pause overlay sits on top of everything and stays hidden until paused
	g.dim = canvas.NewRectangle(color.NRGBA{A: 0xb0})
	g.dim.Hide()
	g.pauseText = canvas.NewText("PAUSED", fgCol)
	g.pauseText.TextSize = 48
	g.pauseText.TextStyle = fyne.TextStyle{Bold: true}
	g.pauseText.Hide()
	objs = append(objs, g.dim, g.pauseText)

	g.field = container.NewWithoutLayout(objs...)
	return g
}

func (g *game) connectInput(w fyne.Window) {
	dc, ok := w.Canvas().(desktop.Canvas)
	if !ok {
		return
	}
	dc.SetOnKeyDown(func(e *fyne.KeyEvent) { g.setKey(e.Name, true) })
	dc.SetOnKeyUp(func(e *fyne.KeyEvent) { g.setKey(e.Name, false) })
}

func (g *game) setKey(name fyne.KeyName, down bool) {
	switch name {
	case fyne.KeyW:
		g.up1 = down
	case fyne.KeyS:
		g.down1 = down
	case fyne.KeyUp:
		g.up2 = down
	case fyne.KeyDown:
		g.down2 = down
	case fyne.KeySpace:
		if down {
			g.setPaused(!g.paused)
		}
	}
}

// setPaused freezes or resumes play and toggles the overlay to match.
func (g *game) setPaused(p bool) {
	if g.paused == p {
		return
	}
	g.paused = p
	if p {
		g.dim.Show()
		g.pauseText.Show()
	} else {
		g.dim.Hide()
		g.pauseText.Hide()
	}
}

func (g *game) start() {
	go func() {
		tick := time.NewTicker(time.Second / 60)
		defer tick.Stop()
		for range tick.C {
			fyne.Do(g.step)
		}
	}()
}

// resetBall serves the ball from the top centre, heading down the field,
// after a short pause like the original.
func (g *game) resetBall(w, h float32) {
	g.ballX = w / 2
	g.ballY = float32(scoreHeight) + ballSize
	g.velX = 6 * g.serveDir
	g.velY = 3.5
	g.serveDir = -g.serveDir
	g.serveWait = 60 // ~1s at 60fps
}

func (g *game) step() {
	size := g.field.Size()
	w, h := size.Width, size.Height
	if w <= 0 || h <= 0 {
		return
	}

	top := float32(scoreHeight)
	playH := h - top

	// initialise paddle / ball positions once we know the size
	if g.velX == 0 && g.velY == 0 {
		g.p1y = top + (playH-paddleHeight)/2
		g.p2y = g.p1y
		g.resetBall(w, h)
	}

	// while paused nothing moves, but keep things laid out for resizes
	if g.paused {
		g.layout(w, h, top)
		return
	}

	// move paddles
	if g.up1 {
		g.p1y -= paddleSpeed
	}
	if g.down1 {
		g.p1y += paddleSpeed
	}
	if g.up2 {
		g.p2y -= paddleSpeed
	}
	if g.down2 {
		g.p2y += paddleSpeed
	}
	g.p1y = clamp(g.p1y, top, h-paddleHeight)
	g.p2y = clamp(g.p2y, top, h-paddleHeight)

	// hold the ball at the serve spot during the pause; paddles still move
	if g.serveWait > 0 {
		g.serveWait--
		g.layout(w, h, top)
		return
	}

	// move ball
	g.ballX += g.velX
	g.ballY += g.velY

	half := float32(ballSize) / 2

	// bounce off top / bottom of the play area
	if g.ballY-half < top {
		g.ballY = top + half
		g.velY = -g.velY
		soundWall()
	}
	if g.ballY+half > h {
		g.ballY = h - half
		g.velY = -g.velY
		soundWall()
	}

	// paddle collisions
	p1x := float32(paddleInset)
	p2x := w - paddleInset - paddleWidth
	if g.velX < 0 && g.ballX-half <= p1x+paddleWidth && g.ballX-half >= p1x &&
		g.ballY >= g.p1y && g.ballY <= g.p1y+paddleHeight {
		g.ballX = p1x + paddleWidth + half
		g.bounce(g.p1y)
	}
	if g.velX > 0 && g.ballX+half >= p2x && g.ballX+half <= p2x+paddleWidth &&
		g.ballY >= g.p2y && g.ballY <= g.p2y+paddleHeight {
		g.ballX = p2x - half
		g.bounce(g.p2y)
	}

	// scoring
	if g.ballX < 0 {
		g.score2++
		g.board2.set(g.score2)
		soundScore()
		g.resetBall(w, h)
	} else if g.ballX > w {
		g.score1++
		g.board1.set(g.score1)
		soundScore()
		g.resetBall(w, h)
	}

	g.layout(w, h, top)
}

// bounce reflects the ball off a paddle, adding spin based on where it hit.
func (g *game) bounce(paddleY float32) {
	soundPaddle()
	g.velX = -g.velX
	// offset in [-1, 1] from the paddle centre
	offset := (g.ballY - (paddleY + paddleHeight/2)) / (paddleHeight / 2)
	g.velY = offset * 6
	// nudge the pace up a touch on every hit
	if g.velX > 0 {
		g.velX += 0.3
	} else {
		g.velX -= 0.3
	}
}

func (g *game) layout(w, h, top float32) {
	g.bg.Resize(fyne.NewSize(w, h))
	g.bg.Move(fyne.NewPos(0, 0))

	// centre net dashes, spanning the play area only
	dashH := (h - top) / float32(len(g.net)*2)
	for i, dash := range g.net {
		dash.Resize(fyne.NewSize(4, dashH))
		dash.Move(fyne.NewPos(w/2-2, top+float32(i*2)*dashH))
	}

	g.paddle1.Resize(fyne.NewSize(paddleWidth, paddleHeight))
	g.paddle1.Move(fyne.NewPos(paddleInset, g.p1y))
	g.paddle2.Resize(fyne.NewSize(paddleWidth, paddleHeight))
	g.paddle2.Move(fyne.NewPos(w-paddleInset-paddleWidth, g.p2y))

	g.ball.Resize(fyne.NewSize(ballSize, ballSize))
	g.ball.Move(fyne.NewPos(g.ballX-ballSize/2, g.ballY-ballSize/2))

	s1 := g.board1.size(g.score1)
	g.board1.cont.Move(fyne.NewPos(w/2-w/4-s1.Width/2, (top-s1.Height)/2))
	s2 := g.board2.size(g.score2)
	g.board2.cont.Move(fyne.NewPos(w/2+w/4-s2.Width/2, (top-s2.Height)/2))

	g.dim.Resize(fyne.NewSize(w, h))
	g.dim.Move(fyne.NewPos(0, 0))
	ts := g.pauseText.MinSize()
	g.pauseText.Move(fyne.NewPos((w-ts.Width)/2, (h-ts.Height)/2))
}

func clamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
