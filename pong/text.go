package pong

import (
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"image/color"
	"log"
	"math"
)

const (
	fontSize      = 30
	smallFontSize = int(fontSize / 2)
)

var (
	ArcadeFont      font.Face
	SmallArcadeFont font.Face
)

func InitFonts() {
	tt, err := truetype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	var dpi float64 = 72
	ArcadeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    float64(fontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	SmallArcadeFont = truetype.NewFace(tt, &truetype.Options{
		Size:    float64(smallFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func DrawCaption(state GameState, color color.Color, screen *ebiten.Image) {
	w, h := screen.Size()
	msg := []string{}
	switch state {
	case StartState, PlayState:
		msg = append(msg, "Player 1: ↑ is UP, ↓ is Down\nPlayer 2: W is UP, S is Down")
	}
	for i, l := range msg {
		n := len(l)*smallFontSize
		x := int(math.Abs(float64(w - n)) / 2)
		text.Draw(screen, l, SmallArcadeFont, x, h-4+(i-2)*smallFontSize, color)
	}
}

func DrawBigText(state GameState, color color.Color, screen *ebiten.Image) {
	w, _ := screen.Size()
	var texts []string
	switch state {
	case StartState:
		texts = []string{
			"",
			"PONG",
			"",
			"SPACE -> START GAME",
		}
	case GameOverState:
		texts = []string{
			"",
			"GAME OVER!",
			"SPACE -> RESET",
		}
	}
	for i, l := range texts {
		x := (w - len(l)*fontSize) / 2
		text.Draw(screen, l, ArcadeFont, x, (i+4)*fontSize, color)
	}
}
