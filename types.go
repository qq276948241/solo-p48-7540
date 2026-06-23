package main

import "time"

const (
	MapW     = 40
	MapH     = 20
	InvMax   = 3
	SideBarW = 24
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
	IChest
)

type EquipType int

const (
	EWeapon EquipType = iota
	EArmor
)

type Equipment struct {
	Name string
	Type EquipType
	Atk  int
	Def  int
}

type Inventory []*Equipment

type Monster struct {
	X, Y  int
	HP    int
	MaxHP int
	ATK   int
	DEF   int
}

type Player struct {
	X, Y   int
	HP     int
	MaxHP  int
	BaseATK int
	BaseDEF int
	Gold    int
	Inv     Inventory
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
