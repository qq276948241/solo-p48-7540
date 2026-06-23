package main

import "math/rand"

var weaponNames = []string{"短刀", "长剑", "战斧", "铁锤", "长矛", "匕首", "大剑", "弯刀"}
var armorNames = []string{"布衣", "皮甲", "锁子甲", "铁甲", "板甲", "法袍", "藤甲", "战甲"}

func (inv Inventory) TotalAtk() int {
	sum := 0
	for _, e := range inv {
		if e != nil {
			sum += e.Atk
		}
	}
	return sum
}

func (inv Inventory) TotalDef() int {
	sum := 0
	for _, e := range inv {
		if e != nil {
			sum += e.Def
		}
	}
	return sum
}

func (p *Player) ATK() int { return p.BaseATK + p.Inv.TotalAtk() }
func (p *Player) DEF() int { return p.BaseDEF + p.Inv.TotalDef() }

func (p *Player) AddEquip(e *Equipment) (removed *Equipment) {
	if len(p.Inv) >= InvMax {
		removed = p.Inv[0]
		p.Inv = append(p.Inv[1:], e)
	} else {
		p.Inv = append(p.Inv, e)
	}
	return
}

func randomEquipment(level int) *Equipment {
	t := EquipType(rand.Intn(2))
	if t == EWeapon {
		name := weaponNames[rand.Intn(len(weaponNames))]
		bonus := 1 + level/2 + rand.Intn(2)
		return &Equipment{Name: name, Type: EWeapon, Atk: bonus}
	} else {
		name := armorNames[rand.Intn(len(armorNames))]
		bonus := 1 + level/2 + rand.Intn(2)
		return &Equipment{Name: name, Type: EArmor, Def: bonus}
	}
}
