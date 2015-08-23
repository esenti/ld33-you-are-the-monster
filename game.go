package main

import (
	// "fmt"
	gc "github.com/rthornton128/goncurses"
	"log"
	"math"
	"math/rand"
	"time"
)

func Sqrt(x float64) float64 {
	z := 1.0

	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}

	return z
}

func Pic(dx, dy, step int) [][]uint8 {
	pic := make([][]uint8, dy)

	for x, _ := range pic {
		pic[x] = make([]uint8, dx)

		for y, _ := range pic[x] {
			// pic[x][y] = uint8((x + y) / (step + 1))
			pic[x][y] = uint8((x + y + step) / 2)
		}
	}

	return pic
}

func Normalize(pic [][]uint8, scale uint8) [][]uint8 {
	var max uint8 = 0

	for x, _ := range pic {
		for _, v := range pic[x] {
			if v > max {
				max = v
			}
		}
	}

	ratio := scale / max

	for x, _ := range pic {
		for y, v := range pic[x] {
			pic[x][y] = v * ratio
		}
	}

	return pic
}

func MapColor(x uint8) gc.Char {
	// fmt.Println(x)
	color := int16(math.Max(1, math.Min(8, float64(x/10))))
	// fmt.Println(color)

	var mapping = map[int16]gc.Char{
		1: gc.ColorPair(1),
		2: gc.ColorPair(1) | gc.A_BOLD,
		3: gc.ColorPair(2),
		4: gc.ColorPair(2) | gc.A_BOLD,
		5: gc.ColorPair(3),
		6: gc.ColorPair(3) | gc.A_BOLD,
		7: gc.ColorPair(4),
		8: gc.ColorPair(4) | gc.A_BOLD,
	}

	return mapping[color]
}

func GetMapper() func(rune) gc.Char {
	var mapping = map[rune]gc.Char{
		'|': gc.ACS_VLINE,
		'-': gc.ACS_HLINE,
		'<': gc.ACS_ULCORNER,
		'>': gc.ACS_URCORNER,
		'[': gc.ACS_LLCORNER,
		']': gc.ACS_LRCORNER,
		'X': gc.ACS_CKBOARD,
	}

	return func(r rune) gc.Char {
		v, p := mapping[r]
		if p {
			return v
		} else {
			return gc.Char(r)
		}
	}
}

var MapChar = GetMapper()

type Vector struct {
	X int
	Y int
}

type Kind interface {
	Draw(screen *gc.Window, y, x int)
	GetSize() Vector
	Update(state *State)
}

type House struct {
	Frames *map[string][][]string
}

func (h *House) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["House"]

	if !p {
		frames := [][]string{[]string{
			"<----->",
			"|X X X|",
			"|     |",
			"|X X X|",
			"[-----]"},
		}

		(*f)["House"] = frames
		v = frames
	}

	for ly, row := range v[0] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *House) GetSize() Vector {
	return Vector{7, 5}
}

func (h *House) Update(state *State) {
}

type SmallHouse struct {
	Frames *map[string][][]string
	Frame  int
}

func (h *SmallHouse) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["SmallHouse"]

	if !p {
		frames := [][]string{[]string{
			"<--->",
			"|X  |",
			"|  X|",
			"[---]"}, []string{
			"<--->",
			"|X  |",
			"|   |",
			"[---]"}, []string{
			"<--->",
			"|   |",
			"|  X|",
			"[---]"},
		}

		(*f)["SmallHouse"] = frames
		v = frames
	}

	for ly, row := range v[h.Frame] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *SmallHouse) GetSize() Vector {
	return Vector{5, 4}
}

func (h *SmallHouse) Update(state *State) {
	h.Frame = rand.Intn(3)
}

type Factory struct {
	Frames *map[string][][]string
}

func (h *Factory) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["Factory"]

	if !p {
		frames := [][]string{[]string{
			"||    ",
			"||    ",
			"|----|",
			"|    |",
			"[----]"},
		}

		(*f)["Factory"] = frames
		v = frames
	}

	for ly, row := range v[0] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *Factory) GetSize() Vector {
	return Vector{6, 5}
}

func (h *Factory) Update(state *State) {
	state.Money += 5
	state.Pollution += 1
}

type Thing struct {
	X    int
	Y    int
	kind Kind
}

func (t *Thing) Overlap(other *Thing) bool {
	tSize := t.kind.GetSize()
	otherSize := other.kind.GetSize()
	return t.X < other.X+otherSize.X+1 && t.X+tSize.X+1 > other.X && t.Y < other.Y+otherSize.Y+1 && t.Y+tSize.Y+1 > other.Y
}

func (t *Thing) Draw(screen *gc.Window) {
	t.kind.Draw(screen, t.Y, t.X)
}

func (t *Thing) Update(state *State) {
	t.kind.Update(state)
}

type State struct {
	Money      int
	Population int
	Pollution  int
}

func main() {
	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init", err)
	}
	defer gc.End()

	if err := gc.StartColor(); err != nil {
		log.Fatal(err)
	}

	lines, cols := stdscr.MaxYX()

	gc.Raw(true)   // turn on raw "uncooked" input
	gc.Echo(false) // turn echoing of typed characters off
	gc.Cursor(0)   // hide cursor
	stdscr.Timeout(30)
	stdscr.Keypad(true) // allow keypad input

	gc.InitPair(1, gc.C_YELLOW, gc.C_BLACK)
	gc.InitPair(2, gc.C_RED, gc.C_BLACK)
	gc.InitPair(3, gc.C_GREEN, gc.C_BLACK)
	gc.InitPair(4, gc.C_BLUE, gc.C_BLACK)
	stdscr.SetBackground(gc.ColorPair(3))

	rand.Seed(time.Now().UTC().UnixNano())

	state := State{Money: 1000}

	var build *Thing

	things := make([]Thing, 0, 20)
	var toUpdate int64 = 0
	lastTimestamp := int64(time.Nanosecond) * time.Now().UTC().UnixNano() / int64(time.Millisecond)

	frames := make(map[string][][]string)
	build = &Thing{10, 10, &Factory{&frames}}

	for i := 0; i < 20; i++ {
		for {
			var kind Kind
			var pop int
			r := rand.Intn(2)

			switch r {
			case 0:
				kind = &SmallHouse{Frames: &frames}
				pop = 10
			case 1:
				kind = &House{&frames}
				pop = 40
			}

			house := &Thing{2 + rand.Intn(cols-5), 2 + rand.Intn(lines-5), kind}

			overlap := false
			for _, thing := range things {
				if house.Overlap(&thing) {
					overlap = true
					break
				}
			}

			if !overlap {
				things = append(things, *house)
				state.Population += pop
				break
			}
		}
	}

	var overlap bool
	price := 600

	for {
		now := int64(time.Nanosecond) * time.Now().UTC().UnixNano() / int64(time.Millisecond)
		delta := now - lastTimestamp
		lastTimestamp = now
		toUpdate -= delta

		c := stdscr.GetChar()

		switch c {
		case 'q':
			break
		case 'a':
			if build != nil {
				build.X--
			}
		case 'd':
			if build != nil {
				build.X++
			}
		case 'w':
			if build != nil {
				build.Y--
			}
		case 's':
			if build != nil {
				build.Y++
			}
		case 'p':
			if !overlap && state.Money >= price {
				things = append(things, *build)
				state.Money -= price
				build = nil
			}
		case 'b':
			if build == nil {
				build = &Thing{10, 10, &Factory{&frames}}
			} else {
				build = nil
			}
		}

		if c == 'q' {
			break
		}

		stdscr.Clear()

		overlap = false

		for _, thing := range things {
			stdscr.AttrOn(gc.ColorPair(3))
			thing.Draw(stdscr)
			if toUpdate <= 0 {
				thing.Update(&state)
			}
			if build != nil && build.Overlap(&thing) {
				overlap = true
			}
		}

		if build != nil {
			if overlap {
				stdscr.AttrOn(gc.ColorPair(2))
			}
			build.Draw(stdscr)
		}

		if toUpdate <= 0 {
			toUpdate = 2000
		}

		stdscr.AttrOn(gc.ColorPair(1) | gc.A_BOLD)
		stdscr.MovePrintf(1, 1, "Cash:       $%d", state.Money)
		stdscr.AttrOff(gc.ColorPair(1) | gc.A_BOLD)
		stdscr.AttrOn(gc.ColorPair(3) | gc.A_BOLD)
		stdscr.MovePrintf(2, 1, "Population: %d", state.Population)
		stdscr.AttrOff(gc.ColorPair(3) | gc.A_BOLD)
		stdscr.AttrOn(gc.ColorPair(2) | gc.A_BOLD)
		stdscr.MovePrintf(3, 1, "Pollution:  %d", state.Pollution)
		stdscr.AttrOff(gc.ColorPair(2) | gc.A_BOLD)
		stdscr.Refresh()
	}

	stdscr.GetChar()
}
