package main

import (
	// "fmt"
	gc "github.com/rthornton128/goncurses"
	"log"
	"math/rand"
	"strings"
	"time"
)

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
	GetColor() gc.Char
}

type House struct {
	Frames *map[string][][]string
	Frame  int
	Dead   bool
}

func (h *House) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	var v [][]string
	var p bool

	if h.Dead {
		h.Frame = 0
		v, p = (*f)["DeadHouse"]

		if !p {
			frames := [][]string{[]string{
				"       ",
				"|      ",
				"|  |  >",
				"| ||  |",
				"[-----]"},
			}

			(*f)["DeadHouse"] = frames
			v = frames
		}
	} else {

		v, p = (*f)["House"]

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

func (h *House) GetColor() gc.Char {
	if !h.Dead {
		return gc.ColorPair(3)
	} else {
		return gc.ColorPair(6) | gc.A_BOLD
	}
}

func (h *House) Update(state *State) {
	if !h.Dead {
		h.Frame = rand.Intn(3)
		if state.Pollution > 1000+50*state.LeaveCooldown {
			if rand.Intn(12000) <= state.Pollution {
				h.Dead = true
				state.Population -= 40
				state.LeaveCooldown += 1
			}
		}
	}
}

type SmallHouse struct {
	Frames *map[string][][]string
	Frame  int
	Dead   bool
}

func (h *SmallHouse) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	var v [][]string
	var p bool

	if h.Dead {
		h.Frame = 0
		v, p = (*f)["DeadSmallHouse"]

		if !p {
			frames := [][]string{[]string{
				"    ",
				"|   ",
				"|   |",
				"[---]"},
			}

			(*f)["DeadSmallHouse"] = frames
			v = frames
		}
	} else {

		v, p = (*f)["SmallHouse"]

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
	if !h.Dead {
		h.Frame = rand.Intn(3)
		if state.Pollution > 300+20*state.LeaveCooldown {
			if rand.Intn(6000) <= state.Pollution {
				h.Dead = true
				state.Population -= 10
				state.LeaveCooldown += 1
			}
		}
	}
}

func (h *SmallHouse) GetColor() gc.Char {
	if !h.Dead {
		return gc.ColorPair(3)
	} else {
		return gc.ColorPair(6) | gc.A_BOLD
	}
}

type Factory struct {
	Frames *map[string][][]string
	Frame  int
}

func (h *Factory) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["Factory"]

	if !p {
		frames := [][]string{[]string{
			"      ",
			"||    ",
			"||    ",
			"|----|",
			"|    |",
			"[----]"}, []string{
			"xx    ",
			"||    ",
			"||    ",
			"|----|",
			"|    |",
			"[----]"}, []string{
			"      ",
			"||    ",
			"||    ",
			"|----|",
			"|  XX|",
			"[----]"},
		}

		(*f)["Factory"] = frames
		v = frames
	}

	for ly, row := range v[h.Frame] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *Factory) GetSize() Vector {
	return Vector{6, 6}
}

func (h *Factory) Update(state *State) {
	state.Money += 5 * (1.0 + state.Boost)
	state.MoneyDelta += 5 * (1.0 + state.Boost)
	state.Pollution += 1
	state.PollutionDelta += 1
	h.Frame = rand.Intn(3)
}

func (h *Factory) GetColor() gc.Char {
	return gc.ColorPair(7) | gc.A_BOLD
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

func (t *Thing) GetColor() gc.Char {
	return t.kind.GetColor()
}

func (t *Thing) GetSize() Vector {
	return t.kind.GetSize()
}

type BigFactory struct {
	Frames *map[string][][]string
	Frame  int
}

func (h *BigFactory) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["BigFactory"]

	if !p {
		frames := [][]string{[]string{
			"             ",
			"|| ||        ",
			"|| ||   <->  ",
			"|-------| |--",
			"|           |",
			"| XXX   XXX |",
			"[-----------]"}, []string{
			"xx xx        ",
			"|| ||        ",
			"|| ||   <->  ",
			"|-------| |--",
			"|           |",
			"| XXX   XXX |",
			"[-----------]"}, []string{
			"             ",
			"|| ||        ",
			"|| ||   |    ",
			"|-------| |--",
			"|           |",
			"| XXX       |",
			"[-----------]"}}

		(*f)["BigFactory"] = frames
		v = frames
	}

	for ly, row := range v[h.Frame] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *BigFactory) GetSize() Vector {
	return Vector{13, 7}
}

func (h *BigFactory) Update(state *State) {
	state.Money += 10 * (1.0 + state.Boost)
	state.MoneyDelta += 10 * (1.0 + state.Boost)
	state.Pollution += 3
	state.PollutionDelta += 3
	h.Frame = rand.Intn(3)
}

func (h *BigFactory) GetColor() gc.Char {
	return gc.ColorPair(7) | gc.A_BOLD
}

type Shop struct {
	Frames *map[string][][]string
	Frame  int
}

func (h *Shop) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["Shop"]

	if !p {
		frames := [][]string{[]string{
			"<------>",
			"| XX   |",
			"[------]"}, []string{
			"<------>",
			"|      |",
			"[------]"}}

		(*f)["Shop"] = frames
		v = frames
	}

	for ly, row := range v[h.Frame] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *Shop) GetSize() Vector {
	return Vector{8, 3}
}

func (h *Shop) Update(state *State) {
	state.Money += 2 * (1.0 + state.Boost)
	state.MoneyDelta += 2 * (1.0 + state.Boost)
	h.Frame = rand.Intn(2)
}

func (h *Shop) GetColor() gc.Char {
	return gc.ColorPair(7) | gc.A_BOLD
}

type Office struct {
	Frames  *map[string][][]string
	Frame   int
	Boosted bool
}

func (h *Office) Draw(screen *gc.Window, y, x int) {
	f := h.Frames
	v, p := (*f)["Office"]

	if !p {
		frames := [][]string{[]string{
			"<----->",
			"|XXXXX|",
			"|     |",
			"|XXXXX|",
			"|     |",
			"|XXXXX|",
			"[-----]"}, []string{
			"<----->",
			"|XXXXX|",
			"|     |",
			"|     |",
			"|     |",
			"|XXXXX|",
			"[-----]"}}

		(*f)["Office"] = frames
		v = frames
	}

	for ly, row := range v[h.Frame] {
		for lx, c := range row {
			screen.MoveAddChar(y+ly, x+lx, MapChar(c))
		}
	}
}

func (h *Office) GetSize() Vector {
	return Vector{7, 7}
}

func (h *Office) Update(state *State) {
	h.Frame = rand.Intn(2)
	if !h.Boosted {
		state.Boost += 0.2
		h.Boosted = true
	}
}

func (h *Office) GetColor() gc.Char {
	return gc.ColorPair(7) | gc.A_BOLD
}

type State struct {
	Money          float32
	Population     int
	Pollution      int
	Boost          float32
	LeaveCooldown  int
	MoneyDelta     float32
	PollutionDelta int
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
	stdscr.Timeout(50)
	stdscr.Keypad(true) // allow keypad input

	gc.InitPair(1, gc.C_YELLOW, gc.C_BLACK)
	gc.InitPair(2, gc.C_RED, gc.C_BLACK)
	gc.InitPair(3, gc.C_GREEN, gc.C_BLACK)
	gc.InitPair(4, gc.C_BLUE, gc.C_BLACK)
	gc.InitPair(5, gc.C_WHITE, gc.C_BLACK)
	gc.InitPair(6, gc.C_BLACK, gc.C_BLACK)
	gc.InitPair(7, gc.C_MAGENTA, gc.C_BLACK)

	stdscr.SetBackground(gc.ColorPair(3))

	rand.Seed(time.Now().UTC().UnixNano())

	state := State{Money: 150}

	var build *Thing

	things := make([]Thing, 0, 20)
	var toUpdate int64 = 0
	lastTimestamp := int64(time.Nanosecond) * time.Now().UTC().UnixNano() / int64(time.Millisecond)

	frames := make(map[string][][]string)

	count := rand.Intn(15) + 25

	for i := 0; i < count; i++ {
		for {
			var kind Kind
			var pop int
			r := rand.Intn(2)

			switch r {
			case 0:
				kind = &SmallHouse{Frames: &frames}
				pop = 10
			case 1:
				kind = &House{Frames: &frames}
				pop = 40
			}

			house := &Thing{2 + rand.Intn(cols-8), 2 + rand.Intn(lines-8), kind}

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
	var price float32
	showInfo := true

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
			if build != nil && build.X > 0 {
				build.X--
			}
		case 'd':
			if build != nil && build.X+build.GetSize().X < cols {
				build.X++
			}
		case 'w':
			if build != nil && build.Y > 0 {
				build.Y--
			}
		case 's':
			if build != nil && build.Y+build.GetSize().Y < lines {
				build.Y++
			}
		case ' ':
			if build != nil && !overlap && state.Money >= price {
				things = append(things, *build)
				state.Money -= price
				build = nil
			}
		case '1':
			if state.Money >= 100 {
				build = &Thing{cols / 2, lines / 2, &Shop{Frames: &frames}}
				price = 100
			}
		case '2':
			if state.Money >= 600 {
				build = &Thing{cols / 2, lines / 2, &Factory{Frames: &frames}}
				price = 600
			}
		case '3':
			if state.Money >= 1100 {
				build = &Thing{cols / 2, lines / 2, &BigFactory{Frames: &frames}}
				price = 1100
			}
		case '4':
			if state.Money >= 2000 {
				build = &Thing{cols / 2, lines / 2, &Office{Frames: &frames}}
				price = 2000
			}
		}

		if c == 'q' {
			break
		}

		stdscr.Erase()

		if state.Population == 0 {
			stdscr.AttrOn(gc.ColorPair(3) | gc.A_BOLD)
			stdscr.MovePrint(lines/2, cols/2-4, "You won!")
			continue
		}

		if showInfo && c != 0 {
			showInfo = false
		}

		if showInfo {
			info := `
  __  __         _____       _     _ _
 |  \/  |       / ____|     | |   | | |
 | \  / |_ __  | |  __  ___ | | __| | |__   ___ _ __ __ _
 | |\/| | '__| | | |_ |/ _ \| |/ _  | '_ \ / _ \ '__/ _  |
 | |  | | |_   | |__| | (_) | | (_| | |_) |  __/ | | (_| |
 |_|  |_|_(_)   \_____|\___/|_|\__,_|_.__/ \___|_|  \__, |
                                                     __/ |
                                                     |___/



			You are Mr. Goldberg, a famous and well respected businessman.
			You're looking to expand your business to a new city.
			It looks promising, but has one serious problem - too many
			people living in it, getting in the way!
			Maybe if you build some factories near their homes they will
			eventually move out and stop bothering you?

			Instructions:
			 * earn money to build stuff
			 * pollution will cause people to move out
			 * you win when city population reaches 0!
			 * 1/2/3/4 to choose building, WASD to move, SPACE to build it, Q to quit

			                        [ any key to start ]
			`

			stdscr.AttrOn(gc.ColorPair(5) | gc.A_BOLD)
			for i, s := range strings.Split(info, "\n") {
				stdscr.MovePrint(lines/2+i-16, cols/2-30, strings.Trim(s, "\t"))
			}
			continue
		}

		overlap = false
		if toUpdate <= 0 {
			state.MoneyDelta = 0
			state.PollutionDelta = 0
		}

		for _, thing := range things {
			stdscr.AttrOn(thing.GetColor())
			thing.Draw(stdscr)
			stdscr.AttrOff(thing.GetColor())
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
			} else {
				stdscr.AttrOn(gc.ColorPair(7))
			}
			build.Draw(stdscr)
		}

		if toUpdate <= 0 {
			toUpdate = 1000
		}

		stdscr.AttrOn(gc.ColorPair(1) | gc.A_BOLD)
		stdscr.MovePrintf(1, 1, "Cash:       $%.2f (+$%.2f/s)", state.Money, state.MoneyDelta)
		stdscr.AttrOff(gc.ColorPair(1) | gc.A_BOLD)
		stdscr.AttrOn(gc.ColorPair(3) | gc.A_BOLD)
		stdscr.MovePrintf(2, 1, "Population: %d", state.Population)
		stdscr.AttrOff(gc.ColorPair(3) | gc.A_BOLD)
		stdscr.AttrOn(gc.ColorPair(2) | gc.A_BOLD)
		stdscr.MovePrintf(3, 1, "Pollution:  %d (+%d/s)", state.Pollution, state.PollutionDelta)
		stdscr.AttrOff(gc.ColorPair(2) | gc.A_BOLD)

		if state.Money >= 100 {
			stdscr.AttrOn(gc.ColorPair(5) | gc.A_BOLD)
		} else {
			stdscr.AttrOn(gc.ColorPair(6) | gc.A_BOLD)
		}

		stdscr.MovePrint(1, cols-50, "1: Shop ($100) [2$/s]")

		if state.Money >= 600 {
			stdscr.AttrOn(gc.ColorPair(5) | gc.A_BOLD)
		} else {
			stdscr.AttrOn(gc.ColorPair(6) | gc.A_BOLD)
		}

		stdscr.MovePrint(3, cols-50, "2: Factory ($600) [5$/s, 1 pollution/s]")

		if state.Money >= 1100 {
			stdscr.AttrOn(gc.ColorPair(5) | gc.A_BOLD)
		} else {
			stdscr.AttrOn(gc.ColorPair(6) | gc.A_BOLD)
		}

		stdscr.MovePrint(5, cols-50, "3: Big factory ($1100) [10$/s, 3 pollution/s]")
		stdscr.AttrOff(gc.A_BOLD)

		if state.Money >= 2000 {
			stdscr.AttrOn(gc.ColorPair(5) | gc.A_BOLD)
		} else {
			stdscr.AttrOn(gc.ColorPair(6) | gc.A_BOLD)
		}

		stdscr.MovePrint(7, cols-50, "4: Office ($2000) [+20% money generated]")
		stdscr.AttrOff(gc.A_BOLD)

		stdscr.Refresh()
	}

	stdscr.GetChar()
}
