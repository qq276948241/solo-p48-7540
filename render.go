package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/nsf/termbox-go"
)

func drawText(x, y int, s string, fg, bg termbox.Attribute) {
	for i, ch := range s {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	status := fmt.Sprintf(" 第%d层 | HP:%d/%d | ATK:%d | DEF:%d | 金币:%d ",
		g.Level, g.Player.HP, g.Player.MaxHP, g.Player.ATK(), g.Player.DEF(), g.Player.Gold)
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
				case IChest:
					ch = 'C'
					fg = termbox.ColorMagenta | termbox.AttrBold
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

	sideX := MapW + 1
	drawText(sideX, 1, "══════════════", termbox.ColorBlue, termbox.ColorDefault)
	drawText(sideX, 2, " 背  包 ", termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault)
	drawText(sideX, 3, "══════════════", termbox.ColorBlue, termbox.ColorDefault)

	for i := 0; i < InvMax; i++ {
		row := 4 + i
		line := fmt.Sprintf("[%d] ", i+1)
		drawText(sideX, row, line, termbox.ColorDarkGray, termbox.ColorDefault)
		if i < len(g.Player.Inv) {
			e := g.Player.Inv[i]
			if e.Type == EWeapon {
				drawText(sideX+4, row, fmt.Sprintf("%s", e.Name), termbox.ColorYellow, termbox.ColorDefault)
				drawText(sideX+4+len([]rune(e.Name)), row, fmt.Sprintf(" 攻+%d", e.Atk), termbox.ColorYellow, termbox.ColorDefault)
			} else {
				drawText(sideX+4, row, fmt.Sprintf("%s", e.Name), termbox.ColorGreen, termbox.ColorDefault)
				drawText(sideX+4+len([]rune(e.Name)), row, fmt.Sprintf(" 防+%d", e.Def), termbox.ColorGreen, termbox.ColorDefault)
			}
		} else {
			drawText(sideX+4, row, "(空)", termbox.ColorDarkGray, termbox.ColorDefault)
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
	help := "WASD移动 | $金币 | +血瓶 | C宝箱 | E怪物 | >楼梯"
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
