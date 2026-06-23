package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	MapW = 40
	MapH = 20
)

type Tile int

const (
	TWall Tile = iota
	TFloor
)

type ItemType int

const (
	INone ItemType = iota
	IGold
	IPotion
	IStairs
)

type Monster struct {
	X, Y  int
	HP    int
	MaxHP int
	ATK   int
	DEF   int
}

type Player struct {
	X, Y int
	HP   int
	MaxHP int
	ATK  int
	DEF  int
	Gold int
}

type Room struct {
	X, Y, W, H int
}

type Game struct {
	Level    int
	Grid     [][]Tile
	Items    map[[2]int]ItemType
	Monsters []*Monster
	Player   Player
	Over     bool
	Message  string
	MsgTime  time.Time
}

func NewGame() *Game {
	g := &Game{
		Level: 1,
		Player: Player{
			HP: 20, MaxHP: 20,
			ATK: 3, DEF: 1,
		},
	}
	g.GenerateLevel()
	return g
}

func (g *Game) GenerateLevel() {
	g.Grid = make([][]Tile, MapH)
	for y := range g.Grid {
		g.Grid[y] = make([]Tile, MapW)
		for x := range g.Grid[y] {
			g.Grid[y][x] = TWall
		}
	}
	g.Items = make(map[[2]int]ItemType)
	g.Monsters = nil

	rooms := g.placeRooms()
	g.connectRooms(rooms)

	startRoom := rooms[0]
	g.Player.X = startRoom.X + startRoom.W/2
	g.Player.Y = startRoom.Y + startRoom.H/2

	floors := g.getFloors()
	rand.Shuffle(len(floors), func(i, j int) { floors[i], floors[j] = floors[j], floors[i] })

	playerKey := [2]int{g.Player.X, g.Player.Y}
	idx := 0
	used := map[[2]int]bool{playerKey: true}

	pickFloor := func() (int, int) {
		for idx < len(floors) {
			k := [2]int{floors[idx][0], floors[idx][1]}
			idx++
			if !used[k] {
				used[k] = true
				return k[0], k[1]
			}
		}
		return -1, -1
	}

	numGold := 3 + rand.Intn(3)
	for i := 0; i < numGold; i++ {
		x, y := pickFloor()
		if x >= 0 {
			g.Items[[2]int{x, y}] = IGold
		}
	}

	numPotions := 1 + rand.Intn(2)
	for i := 0; i < numPotions; i++ {
		x, y := pickFloor()
		if x >= 0 {
			g.Items[[2]int{x, y}] = IPotion
		}
	}

	numMonsters := 2 + g.Level + rand.Intn(2)
	for i := 0; i < numMonsters; i++ {
		x, y := pickFloor()
		if x >= 0 {
			m := &Monster{
				X: x, Y: y,
				HP:    3 + g.Level*2,
				MaxHP: 3 + g.Level*2,
				ATK:   1 + g.Level,
				DEF:   g.Level / 2,
			}
			g.Monsters = append(g.Monsters, m)
		}
	}
}

func (g *Game) placeRooms() []Room {
	var rooms []Room
	attempts := 0
	for len(rooms) < 6 && attempts < 200 {
		w := 4 + rand.Intn(5)
		h := 3 + rand.Intn(4)
		x := 1 + rand.Intn(MapW-w-2)
		y := 1 + rand.Intn(MapH-h-2)

		overlap := false
		for _, r := range rooms {
			if x-1 < r.X+r.W+1 && x+w+1 > r.X-1 && y-1 < r.Y+r.H+1 && y+h+1 > r.Y-1 {
				overlap = true
				break
			}
		}
		if !overlap {
			rooms = append(rooms, Room{X: x, Y: y, W: w, H: h})
			for ry := y; ry < y+h; ry++ {
				for rx := x; rx < x+w; rx++ {
					g.Grid[ry][rx] = TFloor
				}
			}
		}
		attempts++
	}
	if len(rooms) == 0 {
		r := Room{X: 2, Y: 2, W: 6, H: 4}
		rooms = append(rooms, r)
		for ry := r.Y; ry < r.Y+r.H; ry++ {
			for rx := r.X; rx < r.X+r.W; rx++ {
				g.Grid[ry][rx] = TFloor
			}
		}
	}
	return rooms
}

func (g *Game) connectRooms(rooms []Room) {
	for i := 1; i < len(rooms); i++ {
		r1 := rooms[i-1]
		r2 := rooms[i]
		x1 := r1.X + r1.W/2
		y1 := r1.Y + r1.H/2
		x2 := r2.X + r2.W/2
		y2 := r2.Y + r2.H/2

		if rand.Intn(2) == 0 {
			g.carveHCorridor(x1, x2, y1)
			g.carveVCorridor(y1, y2, x2)
		} else {
			g.carveVCorridor(y1, y2, x1)
			g.carveHCorridor(x1, x2, y2)
		}
	}
}

func (g *Game) carveHCorridor(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if y >= 0 && y < MapH && x >= 0 && x < MapW {
			g.Grid[y][x] = TFloor
		}
	}
}

func (g *Game) carveVCorridor(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if y >= 0 && y < MapH && x >= 0 && x < MapW {
			g.Grid[y][x] = TFloor
		}
	}
}

func (g *Game) getFloors() [][2]int {
	var floors [][2]int
	for y := 0; y < MapH; y++ {
		for x := 0; x < MapW; x++ {
			if g.Grid[y][x] == TFloor {
				floors = append(floors, [2]int{x, y})
			}
		}
	}
	return floors
}

func (g *Game) allMonstersDead() bool {
	for _, m := range g.Monsters {
		if m.HP > 0 {
			return false
		}
	}
	return true
}

func (g *Game) monsterAt(x, y int) *Monster {
	for _, m := range g.Monsters {
		if m.HP > 0 && m.X == x && m.Y == y {
			return m
		}
	}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (g *Game) combat(m *Monster) {
	pDmg := maxInt(1, g.Player.ATK-m.DEF)
	mDmg := maxInt(1, m.ATK-g.Player.DEF)

	for m.HP > 0 && g.Player.HP > 0 {
		m.HP -= pDmg
		if m.HP <= 0 {
			g.SetMsg("击败怪物! 受到%d伤害", mDmg)
			g.Player.HP = maxInt(0, g.Player.HP-mDmg)
			if g.allMonstersDead() {
				sx, sy := g.findStairsPos()
				if sx >= 0 {
					g.Items[[2]int{sx, sy}] = IStairs
				}
			}
			return
		}
		g.Player.HP -= mDmg
		if g.Player.HP <= 0 {
			g.Over = true
			g.SetMsg("你被怪物杀死了...")
			return
		}
	}
}

func (g *Game) findStairsPos() (int, int) {
	occupied := map[[2]int]bool{
		{g.Player.X, g.Player.Y}: true,
	}
	for k := range g.Items {
		occupied[k] = true
	}
	for _, m := range g.Monsters {
		occupied[[2]int{m.X, m.Y}] = true
	}

	for y := 0; y < MapH; y++ {
		for x := 0; x < MapW; x++ {
			if g.Grid[y][x] == TFloor && !occupied[[2]int{x, y}] {
				return x, y
			}
		}
	}
	return -1, -1
}

func (g *Game) SetMsg(format string, args ...interface{}) {
	g.Message = fmt.Sprintf(format, args...)
	g.MsgTime = time.Now()
}

func (g *Game) Move(dx, dy int) {
	if g.Over {
		return
	}
	nx := g.Player.X + dx
	ny := g.Player.Y + dy
	if nx < 0 || nx >= MapW || ny < 0 || ny >= MapH {
		return
	}
	if g.Grid[ny][nx] == TWall {
		return
	}

	if m := g.monsterAt(nx, ny); m != nil {
		g.combat(m)
		if !g.Over {
			g.SetMsg("战斗! 你HP:%d 怪物HP:%d", g.Player.HP, m.HP)
		}
		return
	}

	g.Player.X = nx
	g.Player.Y = ny

	key := [2]int{nx, ny}
	if item, ok := g.Items[key]; ok {
		delete(g.Items, key)
		switch item {
		case IGold:
			g.Player.ATK++
			g.Player.Gold++
			g.SetMsg("拾取金币! 攻击力+1 (ATK:%d)", g.Player.ATK)
		case IPotion:
			heal := 5
			if g.Player.HP+heal > g.Player.MaxHP {
				heal = g.Player.MaxHP - g.Player.HP
			}
			g.Player.HP += heal
			g.SetMsg("饮用药水! 恢复%dHP (HP:%d/%d)", heal, g.Player.HP, g.Player.MaxHP)
		case IStairs:
			g.Level++
			g.SetMsg("进入第%d层!", g.Level)
			g.GenerateLevel()
		}
	}
}

func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	status := fmt.Sprintf(" 第%d层 | HP:%d/%d | ATK:%d | DEF:%d | 金币:%d ", g.Level, g.Player.HP, g.Player.MaxHP, g.Player.ATK, g.Player.DEF, g.Player.Gold)
	for i, ch := range status {
		termbox.SetCell(i, 0, ch, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlue)
	}
	for i := len(status); i < MapW; i++ {
		termbox.SetCell(i, 0, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}

	for y := 0; y < MapH; y++ {
		for x := 0; x < MapW; x++ {
			ch := ' '
			fg := termbox.ColorDefault
			bg := termbox.ColorDefault

			switch g.Grid[y][x] {
			case TWall:
				ch = '#'
				fg = termbox.ColorWhite
			case TFloor:
				ch = '.'
				fg = termbox.ColorDarkGray
			}

			if item, ok := g.Items[[2]int{x, y}]; ok {
				switch item {
				case IGold:
					ch = '$'
					fg = termbox.ColorYellow | termbox.AttrBold
				case IPotion:
					ch = '+'
					fg = termbox.ColorGreen | termbox.AttrBold
				case IStairs:
					ch = '>'
					fg = termbox.ColorCyan | termbox.AttrBold
				}
			}

			for _, m := range g.Monsters {
				if m.HP > 0 && m.X == x && m.Y == y {
					ch = 'E'
					fg = termbox.ColorRed | termbox.AttrBold
				}
			}

			if g.Player.X == x && g.Player.Y == y {
				ch = '@'
				fg = termbox.ColorYellow | termbox.AttrBold
			}

			termbox.SetCell(x, y+1, ch, fg, bg)
		}
	}

	msgY := MapH + 1
	msg := g.Message
	if time.Since(g.MsgTime) > 3*time.Second {
		msg = ""
	}
	if g.Over {
		msg = fmt.Sprintf("游戏结束! 得分:%d 到达层数:%d 按Q退出", g.Player.Gold+g.Level*10, g.Level)
	}
	for i, ch := range msg {
		termbox.SetCell(i, msgY, ch, termbox.ColorWhite, termbox.ColorDefault)
	}

	helpY := msgY + 1
	help := "WASD移动 | $金币(攻击+1) | +血瓶(回复5HP) | E怪物 | >楼梯"
	if g.Over {
		help = "按Q退出游戏"
	}
	for i, ch := range help {
		termbox.SetCell(i, helpY, ch, termbox.ColorDarkGray, termbox.ColorDefault)
	}

	termbox.Flush()
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[2J\033[H")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := termbox.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "终端初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	game := NewGame()
	game.Render()

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			default:
				switch ev.Ch {
				case 'w', 'W':
					game.Move(0, -1)
				case 's', 'S':
					game.Move(0, 1)
				case 'a', 'A':
					game.Move(-1, 0)
				case 'd', 'D':
					game.Move(1, 0)
				case 'q', 'Q':
					if game.Over {
						break mainloop
					}
				}
			}
		case termbox.EventError:
			break mainloop
		}
		game.Render()
	}
}
