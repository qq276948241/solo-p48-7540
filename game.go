package main

import (
	"fmt"
	"time"
)

func NewGame() *Game {
	g := &Game{
		Level: 1,
		Player: Player{
			HP: 20, MaxHP: 20,
			BaseATK: 3, BaseDEF: 1,
			Inv: Inventory{},
		},
	}
	g.GenerateLevel()
	return g
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
			g.Player.BaseATK++
			g.Player.Gold++
			g.SetMsg("拾取金币! 攻击+1 (ATK:%d)", g.Player.ATK())
		case IPotion:
			heal := 5
			if g.Player.HP+heal > g.Player.MaxHP {
				heal = g.Player.MaxHP - g.Player.HP
			}
			g.Player.HP += heal
			g.SetMsg("饮用药水! 恢复%dHP (HP:%d/%d)", heal, g.Player.HP, g.Player.MaxHP)
		case IChest:
			eq := randomEquipment(g.Level)
			removed := g.Player.AddEquip(eq)
			if eq.Type == EWeapon {
				if removed != nil {
					g.SetMsg("★宝箱开出[%s](ATK+%d) 背包已满! 丢弃[%s]装入新武器!", eq.Name, eq.Atk, removed.Name)
				} else {
					g.SetMsg("★宝箱开出武器[%s]! ATK+%d (总ATK:%d)", eq.Name, eq.Atk, g.Player.ATK())
				}
			} else {
				if removed != nil {
					g.SetMsg("★宝箱开出[%s](DEF+%d) 背包已满! 丢弃[%s]装入新护甲!", eq.Name, eq.Def, removed.Name)
				} else {
					g.SetMsg("★宝箱开出护甲[%s]! DEF+%d (总DEF:%d)", eq.Name, eq.Def, g.Player.DEF())
				}
			}
		case IStairs:
			g.Level++
			g.SetMsg("进入第%d层!", g.Level)
			g.GenerateLevel()
		}
	}
}
