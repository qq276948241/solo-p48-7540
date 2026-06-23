package main

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

func (g *Game) combat(m *Monster) {
	pDmg := maxInt(1, g.Player.ATK()-m.DEF)
	mDmg := maxInt(1, m.ATK-g.Player.DEF())

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
