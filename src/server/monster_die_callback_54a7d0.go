package server

// MonsterDieCallbackKind54A7D0 identifies the five monster.bin DIE_FUNCTION
// callbacks implemented at GAME.EXE 0054A7D0-0054A950.
type MonsterDieCallbackKind54A7D0 uint8

const (
	MonsterDieCallbackSwordsman54A7D0 MonsterDieCallbackKind54A7D0 = iota + 1
	MonsterDieCallbackUrchinShaman54A850
	MonsterDieCallbackArcher54A890
	MonsterDieCallbackOgre54A900
	MonsterDieCallbackOgreWarlord54A950
)

// MonsterDieDrop54A390 is the pointer-width-independent input to the common
// cooperative item-drop helper at GAME.EXE 0054A390.
type MonsterDieDrop54A390 struct {
	TypeID    string
	Modifiers [4]string
	Charge    uint8
}

// MonsterDieCallbackRuntime54A7D0 supplies the random source, cooperative-mode
// query, and item creation used by the five restored DIE_FUNCTION callbacks.
type MonsterDieCallbackRuntime54A7D0 struct {
	RandomInt func(minimum, maximum int) int
	CoopMode  func() bool
	DropItem  func(*Object, MonsterDieDrop54A390)
}

func monsterDieDropForRoll54A7D0(kind MonsterDieCallbackKind54A7D0, roll int) (MonsterDieDrop54A390, bool) {
	switch kind {
	case MonsterDieCallbackSwordsman54A7D0:
		if roll <= 20 {
			return MonsterDieDrop54A390{}, false
		}
		if roll <= 50 {
			return MonsterDieDrop54A390{
				TypeID:    "Sword",
				Modifiers: [4]string{"WeaponPower1", "Material1"},
			}, true
		}
		return MonsterDieDrop54A390{
			TypeID:    "WoodenShield",
			Modifiers: [4]string{"", "Material1"},
		}, true
	case MonsterDieCallbackUrchinShaman54A850:
		if roll <= 25 {
			return MonsterDieDrop54A390{}, false
		}
		return MonsterDieDrop54A390{
			TypeID:    "StaffWooden",
			Modifiers: [4]string{"WeaponPower1"},
		}, true
	case MonsterDieCallbackArcher54A890:
		if roll <= 20 {
			return MonsterDieDrop54A390{}, false
		}
		if roll <= 50 {
			return MonsterDieDrop54A390{TypeID: "Bow"}, true
		}
		return MonsterDieDrop54A390{TypeID: "Quiver"}, true
	case MonsterDieCallbackOgre54A900:
		if roll <= 25 {
			return MonsterDieDrop54A390{}, false
		}
		return MonsterDieDrop54A390{
			TypeID:    "OgreAxe",
			Modifiers: [4]string{"WeaponPower1", "Material2"},
		}, true
	case MonsterDieCallbackOgreWarlord54A950:
		if roll <= 25 {
			return MonsterDieDrop54A390{}, false
		}
		return MonsterDieDrop54A390{TypeID: "FanChakram", Charge: 5}, true
	default:
		return MonsterDieDrop54A390{}, false
	}
}

func validMonsterDieCallbackKind54A7D0(kind MonsterDieCallbackKind54A7D0) bool {
	return kind >= MonsterDieCallbackSwordsman54A7D0 && kind <= MonsterDieCallbackOgreWarlord54A950
}

// MonsterDieCallbackNative54A7D0 restores the five monster.bin DIE_FUNCTION
// callbacks without passing the dying monster through a PE32-sized integer.
// The random roll always occurs, while the shared drop helper only checks
// cooperative mode when the roll selected an item, matching GAME.EXE.
func MonsterDieCallbackNative54A7D0(unit *Object, kind MonsterDieCallbackKind54A7D0, runtime MonsterDieCallbackRuntime54A7D0) bool {
	if unit == nil || !validMonsterDieCallbackKind54A7D0(kind) || runtime.RandomInt == nil {
		return false
	}
	drop, selected := monsterDieDropForRoll54A7D0(kind, runtime.RandomInt(0, 100))
	if !selected {
		return true
	}
	if runtime.CoopMode == nil {
		return false
	}
	if !runtime.CoopMode() {
		return true
	}
	if runtime.DropItem == nil {
		return false
	}
	runtime.DropItem(unit, drop)
	return true
}
