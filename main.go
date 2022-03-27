package main

import (
	"fmt"
	// "github.com/hajimehoshi/ebiten"
	// "github.com/hajimehoshi/ebiten/ebitenutil"
	// "github.com/hajimehoshi/ebiten/inpututil"
	"github.com/jtestard/go-pong/pong"
	"time"
	"log"
	"net/http"
)

// Game is the structure of the game State
type Game struct {
	State    pong.GameState
	Ball     *pong.Ball
	Player1  *pong.Paddle
	Player2  *pong.Paddle
	SpaceBar int
	rally    int
	level    int
	maxScore int
}

var SpaceBar int = 1

const (
	initBallVelocity = 5.0
	initPaddleSpeed  = 6.0
	speedUpdateCount = 1
	speedIncrement   = 0.5
)

const (
	windowWidth  = 800
	windowHeight = 600
)

// NewGame creates an initializes a new game
func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) init() {
	g.State = pong.StartState
	g.maxScore = 11
	g.SpaceBar = 1

	g.Player1 = &pong.Paddle{
		Position: pong.Position{
			X: pong.InitPaddleShift,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  pong.InitPaddleWidth,
		Height: pong.InitPaddleHeight,
		Color:  pong.ObjColor,
		Up:     &pong.KeyPressType{
			Keyup: true,
			Keydown: false,
		},
		Down:   &pong.KeyPressType{
			Keyup: true,
			Keydown: false,
		},
	}
	g.Player2 = &pong.Paddle{
		Position: pong.Position{
			X: windowWidth - pong.InitPaddleShift,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  pong.InitPaddleWidth,
		Height: pong.InitPaddleHeight,
		Color:  pong.ObjColor,
		Up:     &pong.KeyPressType{
			Keyup: true,
			Keydown: false,
		},
		Down:   &pong.KeyPressType{
			Keyup: true,
			Keydown: false,
		},
	}
	g.Ball = &pong.Ball{
		Position: pong.Position{
			X: float32(windowWidth / 2),
			Y: float32(windowHeight / 2)},
		Radius:    pong.InitBallRadius,
		Color:     pong.ObjColor,
		XVelocity: initBallVelocity,
		YVelocity: initBallVelocity,
	}
	g.level = 0
}

func (g *Game) Reset(State pong.GameState) {
	w := windowWidth
	g.State = State
	g.SpaceBar = SpaceBar
	g.rally = 0
	g.level = 0
	if State == pong.StartState {
		g.Player1.Score = 0
		g.Player2.Score = 0
	}
	if State == pong.PlayState {
		g.State = pong.StartState
	}
	g.Player1.Position = pong.Position{
		X: pong.InitPaddleShift, Y: pong.GetCenter(windowWidth,windowHeight).Y}
	g.Player2.Position = pong.Position{
		X: float32(w - pong.InitPaddleShift), Y: pong.GetCenter(windowWidth,windowHeight).Y}
	g.Ball.Position = pong.GetCenter(windowWidth,windowHeight)
	g.Ball.XVelocity = initBallVelocity
	g.Ball.YVelocity = initBallVelocity
	g.Player1.Speed = initPaddleSpeed
	g.Player2.Speed = initPaddleSpeed
}

// Update updates the game State
func (g *Game) Update() error {
	switch g.State {
	case pong.StartState:
		if SpaceBar*g.SpaceBar < 0 {
			g.SpaceBar = SpaceBar
			g.State = pong.PlayState
		}

	case pong.PlayState:
		w := windowWidth

		g.Player1.Update(windowHeight)
		g.Player2.Update(windowHeight)

		xV := g.Ball.XVelocity
		g.Ball.Update(g.Player1, g.Player2, windowHeight)
		// rally count
		if xV*g.Ball.XVelocity < 0 {
			// score up when Ball touches human player's paddle
			if g.Ball.X < float32(w/2) {
				// g.Player1.Score++
				if (g.rally)%speedUpdateCount == 0 {
					g.level++
					g.Ball.XVelocity += speedIncrement
					g.Ball.YVelocity += speedIncrement
					g.Player1.Speed += speedIncrement
					g.Player2.Speed += speedIncrement
				}
			}

			g.rally++

			// spice things up
			
		}

		if g.Ball.X < 0 {
			g.Player2.Score++
			g.Reset(pong.PlayState)
		} else if g.Ball.X > float32(w) {
			g.Player1.Score++
			g.Reset(pong.PlayState)
		}

		if g.Player1.Score >= g.maxScore || g.Player2.Score >= g.maxScore {
			g.State = pong.GameOverState
		}

	case pong.GameOverState:
		if SpaceBar*g.SpaceBar < 0 {
			g.SpaceBar = SpaceBar
			g.Reset(pong.StartState)

		}
	}

	return nil
}

// Layout sets the screen layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}


func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	g := NewGame()

	ticker := time.NewTicker(16 * time.Millisecond)
    // done := make(chan bool)

	hub := NewHub(g)
	go hub.Run()

    go func() {
        for {
            select {
            case <-hub.done:
                SpaceBar = -SpaceBar
            case  <-ticker.C:
                g.Update()
                hub.broadcast <- g
            }
        }
    }()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})
	// go func() {
	// 	time.Sleep(15 * time.Second)
	//     g.State = pong.PlayState
	//     time.Sleep(45 * time.Second)
	//     ticker.Stop()
	//     done <- true
	// }()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}



	// if err := ebiten.RunGame(g); err != nil {
	// 	panic(err)
	// }
	fmt.Println("Hello World")
}
