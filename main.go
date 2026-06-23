package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

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
