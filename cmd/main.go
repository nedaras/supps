package main

import (
	"fmt"
	"math/bits"
	"sort"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"supps.go/pkg/mrbiceps"
)

type State = struct {
	Idx   uint8
	Idxs  uint8
	Flags uint8
}

var state State = State{
	Idx:   0,
	Idxs:  2,
	Flags: 0,
}

const (
	MrBiceps  uint8 = 0
	MyProtein uint8 = 1
	Protein   uint8 = 2
	Creatine  uint8 = 3
)

func main() {
	if true {
		p, err := mrbiceps.GetProductsData()
		if err != nil {
			panic(err)
		}
		fmt.Println("len:", p.Length())

		ps := make([]mrbiceps.Product, 0, p.Length())
		err = p.Each(func(i int, p mrbiceps.Product) {
			ps = append(ps, p)
		})
		if err != nil {
			panic(err)
		}

		sort.Slice(ps, func(i, j int) bool {
			return ps[i].Value > ps[j].Value
		})

		for _, p := range ps {
			fmt.Printf("p: %v\n", p)
		}
	}

	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	defer s.Fini()

	if err := s.Init(); err != nil {
		panic(err)
	}

	render(s)
	s.Show()

	for {
		switch ev := s.PollEvent().(type) {
		case *tcell.EventKey:
			flag := state.Flags&0b10000000 != 0

			if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
				return
			}

			if flag {
			} else {
				if (ev.Key() == tcell.KeyUp) || (ev.Key() == tcell.KeyRune && ev.Rune() == 'k') {
					state.Idx = (state.Idx + state.Idxs - 1) % state.Idxs
				}

				if (ev.Key() == tcell.KeyDown) || (ev.Key() == tcell.KeyRune && ev.Rune() == 'j') {
					state.Idx = (state.Idx + 1) % state.Idxs
				}

				if ev.Key() == tcell.KeyEnter {
					state.Flags ^= (1 << state.Idx)
				}

				if state.Flags&0x3 != 0 {
					state.Idxs = 4
				} else {
					state.Flags = state.Flags & 0b10000011
					state.Idxs = 2
				}
			}

			if ev.Key() == tcell.KeyRune && ev.Rune() == 's' {
				state.Flags = state.Flags & 0b01111111
			}

			if ev.Key() == tcell.KeyRune && ev.Rune() == 'p' {
				state.Flags = state.Flags | 0b10000000
			}
		}
		s.Clear()
		render(s)
		s.Show()
	}
}

func render(s tcell.Screen) {
	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	flag := state.Flags&0b10000000 != 0

	text(s, 0, 0, "┌────────────────┬───────────────────┬────────────────────┐", gray)
	text(s, 0, 1, "│                │                   │                    │", gray)
	text(s, 0, 2, "└────────────────┴───────────────────┴────────────────────┘", gray)

	text(s, 5, 1, "supps.go", white.Bold(true))
	text(s, 22, 1, "s", white)

	if flag {
		text(s, 24, 1, "selection", gray)
	} else {
		text(s, 24, 1, "selection", white)
	}

	text(s, 42, 1, "p", white)

	if flag {
		text(s, 44, 1, "products", white)
	} else {
		text(s, 44, 1, "products", gray)
	}

	text(s, 53, 1, strconv.Itoa(bits.OnesCount8(state.Flags&0b01111100)), white)

	text(s, 0, 16, "───────────────────────────────────────────────────────────", gray)

	if flag {

		// No products selected.
		println("Sigma\nnbohuafhgf\ngfregrrgerge\nijwheuegeurgiugrgeiygewr\n")

	} else {
		mrbiceps := gray
		myprotein := gray

		if state.Idx == MrBiceps {
			mrbiceps = white.Background(tcell.ColorDarkOrange)
		}

		if state.Idx == MyProtein {
			myprotein = white.Background(tcell.ColorBlue)
		}

		prefix := func(idx uint8) string {
			if state.Flags&(1<<idx) != 0 {
				return "◉"
			}
			return "○"
		}

		text(s, 2, 4, "~ select shops ~", white)
		text(s, 1, 5, " "+prefix(MrBiceps)+" MrBiceps          ", mrbiceps)
		text(s, 1, 6, " "+prefix(MyProtein)+" MyProtein         ", myprotein)

		if state.Flags&0x3 != 0 {
			proteine := gray
			creteine := gray

			if state.Idx == Protein {
				proteine = white.Background(tcell.ColorRed)
			}

			if state.Idx == Creatine {
				creteine = white.Background(tcell.ColorBlack)
			}

			text(s, 2, 8, "~ select products ~", white)
			text(s, 1, 9, " "+prefix(Protein)+" protein           ", proteine)
			text(s, 1, 10, " "+prefix(Creatine)+" creatine          ", creteine)
		}
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
