package main

import "math/rand"

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

	numChests := 1 + rand.Intn(2)
	for i := 0; i < numChests; i++ {
		x, y := pickFloor()
		if x >= 0 {
			g.Items[[2]int{x, y}] = IChest
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
