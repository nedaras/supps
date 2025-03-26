package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type State = struct {
  Idx   int
  Idxs  int
  Flags uint8
}

var state State = State{
  Idx: 0,
  Idxs: 2,
  Flags: 0,
}

const (
    MrBiceps  int = 0
    MyProtein int = 1
    Protein  int = 2
    Creatine  int = 3
)
func main() {
  s, err := tcell.NewScreen()
  if err != nil {
    panic(err)
  }
  defer s.Fini()

  if err := s.Init(); err != nil {
    panic(err)
  }

  render(s);
  s.Show()

  for {
    switch ev := s.PollEvent().(type) {
    case *tcell.EventKey:
      if ev.Key() == tcell.KeyUp {
        state.Idx = (state.Idx + 1) % state.Idxs
      }

      if ev.Key() == tcell.KeyDown {
        state.Idx = (state.Idx + 1) % state.Idxs
      }

      if ev.Key() == tcell.KeyEnter {
        state.Flags ^= (1 << state.Idx)
      }

      if state.Flags & 0x3 != 0 {
        state.Idxs = 4
      } else {
        state.Idxs = 2
      }

      if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
        return
      }
    }
    s.Clear()
    render(s);
    s.Show()
  }
}

func render(s tcell.Screen) {
  white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
  gray := tcell.StyleDefault.Foreground(tcell.ColorGray)

  text(s, 0, 0, "┌────────────────┐", gray)
  text(s, 0, 1, "│    ", gray)
  text(s, 5, 1, "supps.go", white.Bold(true))
  text(s, 13, 1, "    │", gray)
  text(s, 0, 2, "└────────────────┘", gray)

  mrbiceps := gray
  myprotein := gray

  if state.Idx == MrBiceps {
    mrbiceps = white.Background(tcell.ColorDarkOrange)
  }

  if state.Idx == MyProtein {
    myprotein = white.Background(tcell.ColorBlue)
  }

  prefix := func(idx int) string {
    if state.Flags & (1 << idx) != 0 {
      return "◉"
    }
    return "○"
  }

	text(s, 0, 4, "~ select shops ~", white)
	text(s, 0, 5, prefix(MrBiceps) + " MrBiceps         ", mrbiceps)
	text(s, 0, 6, prefix(MyProtein) + " MyProtein        ", myprotein)

  if state.Flags & 0x3 != 0 {
    proteine := gray
    creteine := gray

    if state.Idx == Protein {
      proteine = white.Background(tcell.ColorBlack)
    }

    if state.Idx == Creatine {
      creteine = white.Background(tcell.ColorBlack)
    }

    text(s, 0, 8, "~ select products ~", white)
    text(s, 0, 9, prefix(Protein) + " protein          ", proteine)
    text(s, 0, 10, prefix(Creatine) + " creatine         ", creteine)
  }
}

func text(s tcell.Screen, x int, y int, str string, style tcell.Style) {
  for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
  }
}
