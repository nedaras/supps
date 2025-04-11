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

type State struct {
	Idx      uint8
	Idxs     uint8
	Flags    uint8
	Fetching bool // could make into flags but no real point
	Cursor   int
	Products []mrbiceps.Product
}

var state State = State{
	Idxs: 2,
}

const (
	MrBiceps  uint8 = 0
	MyProtein uint8 = 1
	Protein   uint8 = 2
	Creatine  uint8 = 3
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

	render(s)
	s.Show()

	type Event struct {
		product mrbiceps.Product
		err     error
	}

	eventCh := make(chan tcell.Event)
	productCh := make(chan Event)

	go func() {
		i := 0
		for {
			eventCh <- s.PollEvent()
			i++
		}
	}()

	fetchProducts := func() {
		products, err := mrbiceps.GetProductsData()
		if err != nil {
			productCh <- Event{
				product: mrbiceps.Product{},
				err:     err,
			}
		}

		err = products.Each(func(i int, p mrbiceps.Product) {
			productCh <- Event{
				product: p,
				err:     nil,
			}
		})

		if err != nil {
			productCh <- Event{
				product: mrbiceps.Product{},
				err:     err,
			}
		}

		productCh <- Event{
			product: mrbiceps.Product{},
			err:     nil,
		}
	}

	for {
		select {
		case product := <-productCh:
			if product.err != nil {
				panic(product.err)
			}

			if (product.product == mrbiceps.Product{}) {
				state.Fetching = false
				sort.Slice(state.Products, func(i, j int) bool {
					return state.Products[i].Value > state.Products[j].Value
				})
			} else {
				state.Products = append(state.Products, product.product)
			}
		case event := <-eventCh:
			switch ev := (event).(type) {
			case *tcell.EventKey:
				flag := state.Flags&0b10000000 != 0

				if ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
					return
				}

				if flag {
					if !state.Fetching && ev.Key() == tcell.KeyEnter && bits.OnesCount8(state.Flags&0b01111100) != 0 {
						state.Products = []mrbiceps.Product{}
						state.Fetching = true
						state.Cursor = 0

						go fetchProducts()
					}

					if (ev.Key() == tcell.KeyUp) || (ev.Key() == tcell.KeyRune && ev.Rune() == 'k') {
						state.Cursor = max(state.Cursor-1, 0)
					}

					if (ev.Key() == tcell.KeyDown) || (ev.Key() == tcell.KeyRune && ev.Rune() == 'j') {
						state.Cursor = min(state.Cursor+1, len(state.Products)-1)
					}

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
	products := bits.OnesCount8(state.Flags & 0b01111100)

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

	text(s, 53, 1, strconv.Itoa(products), white)

	text(s, 0, 18, "───────────────────────────────────────────────────────────", gray)

	text(s, 13, 19, "↑/↓", white.Bold(true))
	text(s, 17, 19, "items", gray)

	text(s, 25, 19, "enter", white.Bold(true))
	text(s, 31, 19, "select", gray)

	text(s, 40, 19, "q", white.Bold(true))
	text(s, 42, 19, "quit", gray)

	if flag {
		if products == 0 {
			text(s, 19, 9, "No products selected.", gray)
		}

		if len(state.Products) != 0 {
			if state.Fetching {
				product := state.Products[len(state.Products)-1]
				text(s, 1, 17, product.Name+"...", gray)
			} else {
				for i := state.Cursor; i < len(state.Products) && i < state.Cursor+3; i++ {
					product := state.Products[i]
					y := i - state.Cursor

					hi := gray
					if i == state.Cursor {
						hi = white
					}

					text(s, 0, 4+y*4, "┌─────────────────────────────────────────────────────────┐", hi)
					text(s, 0, 5+y*4, "│                                                         │", hi)
					text(s, 0, 6+y*4, "│                                                         │", hi)
					text(s, 0, 7+y*4, "└─────────────────────────────────────────────────────────┘", hi)

					text(s, 2, 5+y*4, product.Name, white) // add like ... if like u know like u know like aaaaaaaaaaaa... │
					text(s, 48, 5+y*4, fmt.Sprintf("%.2f €", product.Price), gray)

					text(s, 2, 6+y*4, fmt.Sprintf("%.2f g/€", product.Value), gray)
				}
			}
		} else if products != 0 {
			text(s, 19, 9, "Press       to query.", gray)
			text(s, 25, 9, "enter", white.Bold(true))
		}
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
