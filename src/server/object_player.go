package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

type PlayerState byte

const (
	PlayerState0         = PlayerState(0)
	PlayerState1         = PlayerState(1)
	PlayerState2         = PlayerState(2)
	PlayerState3         = PlayerState(3)
	PlayerState4         = PlayerState(4)
	PlayerState5         = PlayerState(5)
	PlayerState6         = PlayerState(6)
	PlayerState7         = PlayerState(7)
	PlayerState8         = PlayerState(8)
	PlayerState9         = PlayerState(9)
	PlayerState10        = PlayerState(10)
	PlayerState11        = PlayerState(11)
	PlayerState12        = PlayerState(12)
	PlayerState13        = PlayerState(13)
	PlayerState14        = PlayerState(14)
	PlayerState15        = PlayerState(15)
	PlayerState16        = PlayerState(16)
	PlayerState17        = PlayerState(17)
	PlayerState18        = PlayerState(18)
	PlayerState19        = PlayerState(19)
	PlayerState20        = PlayerState(20)
	PlayerState21        = PlayerState(21)
	PlayerState22        = PlayerState(22)
	PlayerState23        = PlayerState(23)
	PlayerState24        = PlayerState(24)
	PlayerStateShakeFist = PlayerState(25)
	PlayerStateLaugh     = PlayerState(26)
	PlayerState27        = PlayerState(27)
	PlayerStatePoint     = PlayerState(28)
	PlayerState29        = PlayerState(29)
	PlayerState30        = PlayerState(30)
	PlayerState31        = PlayerState(31)
	PlayerState32        = PlayerState(32)
	PlayerState33        = PlayerState(33)
)

// QuestSoulGate keeps the domain meaning of PlayerUpdateData.SoulGate while
// retaining Object's native pointer representation.
type QuestSoulGate = Object

type PlayerUpdateData struct {
	Field0              uint32             // 0, 0
	ManaCur             uint16             // 1, 4
	ManaPrev            uint16             // 1, 6
	ManaMax             uint16             // 2, 8
	Field2_1            uint16             // 2, 10
	HealthSamples       [32]uint16         // 3..18, 12..75; initialized by GAME.EXE 004EE730
	HealthSampleCur     uint16             // 19, 76; separately reloaded trailing sample
	Field19_1           uint16             // 19, 78
	Field20_0           uint16             // 20, 80
	Field20_1           uint16             // 20, 82
	Field21             uint32             // 21, 84
	State               PlayerState        // 22, 88
	State2              PlayerState        // 22, 89
	Field22_2           uint8              // 22, 90
	Stamina             uint8              // 22, 91; current player stamina
	Field23             uint32             // 23, 92
	Field24             uint32             // 24, 96
	Field25             uint32             // 25, 100
	EquippedWeapon      *Object            // 26, 104; current player weapon
	Field27             uint32             // 27, 108
	Field28             uint32             // 28, 112
	Field29             [4]*Object         // 29, 116, TODO: teleport markers? traps?
	HarpoonTarg         *Object            // 33, 132
	HarpoonBolt         *Object            // 34, 136
	Harpoon35           uint32             // 35, 140
	HarpoonTargX        float32            // 36, 144
	HarpoonTargY        float32            // 37, 148
	HarpoonFrame        uint32             // 38, 152
	Field39             uint32             // 39, 156
	Field40_0           uint16             // 40, 160
	Field40_1           uint16             // 40, 162
	Field41             uint32             // 41, 164
	CustomWaypoints     [3]*Object         // 42..44, 168; PlayerWaypoint objects
	CustomWaypointWrite uint8              // 45, 180
	CustomWaypointRead  uint8              // 45, 181
	Field45_2           uint16             // 45, 182
	SpellPhonemeLeaf    *PhonemeLeaf       // 46, 184
	Field47_0           uint8              // 47, 188
	Field47_1           uint8              // 47, 189
	Field47_2           uint16             // 47, 190
	TrapSpells          [5]uint32          // 48, 192
	TrapSpellsCnt       uint32             // 53, 212
	SpellCastStart      uint32             // 54, 216
	Field55             int                // 55, 220, TODO: spell-related? x coord?
	Field56             int                // 56, 224, TODO: spell-related? y coord?
	Field57             uint32             // 57, 228
	Field58             uint32             // 58, 232
	Field59_0           uint8              // 59, 236, TODO: frame index?
	Field59_1           uint8              // 59, 237
	Field59_2           uint16             // 59, 238
	MovementFlags       uint32             // 60, 240; movement modifier and direction bits
	CurTraps            uint32             // 61, 244
	Field62             uint32             // 62, 248
	Field63             uint32             // 63, 252
	Field64             uint32             // 64, 256
	IsCamping           uint32             // 65, 260
	Field66             uint32             // 66, 264
	Field67             uint32             // 67, 268
	Field68             uint32             // 68, 272
	Player              *Player            // 69, 276
	Trade70             *TradeSession      // 70, 280
	DialogWith          *Object            // 71, 284
	CursorObj           *Object            // 72, 288
	Field73             *MonsterUpdateData // 73, 292; original bot's monster update record
	CollisionWall       *Wall              // 74, 296; last wall recorded by movement collision
	Field75             uint32             // 75, 300
	Field76             uint32             // 76, 304
	SoulGate            *QuestSoulGate     // 77, 308
	QuestExit           *Object            // 78, 312; Quest exit currently occupied by the player
	QuestWarpGate       *Object            // 79, 316; Quest warp gate currently occupied by the player
	ExtraLives          uint32             // 80, 320; tradable Ankhs currently held
	QuestPlayerState    [32]uint32         // 81..112, 324..451; indexed by PlayerInd
	RespawnMarkers      [32]byte           // 113..120, 452..483; indexed by PlayerInd in Quest respawn
	QuestPlayerFlagsA   [32]byte           // 121..128, 484..515; indexed by PlayerInd
	QuestPlayerFlagsB   [32]byte           // 129..136, 516..547; indexed by PlayerInd
	Field137            uint32             // 137, 548, TODO: some timestamp
	Field138            uint32             // 138, 552
}

func (obj *Object) ChangeScore(val int) {
	if !obj.Class().Has(object.ClassPlayer) {
		return
	}
	obj.changeScore(val)
	s := obj.Server()
	if tm := obj.Team(); tm != nil {
		s.TeamChangeLessons(tm, val+int(tm.Lessons))
	}
	s.Nox_xxx_netReportLesson_4D8EF0(obj)
}

func (obj *Object) changeScore(val int) { // nox_xxx_playerSubLessons_4D8EC0(-v), nox_xxx_changeScore_4D8E90(v)
	if !obj.Class().Has(object.ClassPlayer) {
		return
	}
	pl := obj.ControllingPlayer()
	pl.Lessons += int32(val)
}

func (obj *Object) PlayerSpellPhoneme(ph spell.Phoneme, aud sound.ID, resetTimer bool) {
	s := obj.Server()
	ud := obj.UpdateDataPlayer()
	if ud.SpellCastStart == 0 {
		ud.SpellPhonemeLeaf = s.Spells.PhonemeTree()
		ud.SpellCastStart = s.Frame()
	} else if resetTimer {
		ud.SpellCastStart = s.Frame()
	}
	ud.SpellPhonemeLeaf = ud.SpellPhonemeLeaf.Next(ph)
	s.Audio.EventObj(aud, obj, 0, 0)
	ud.Field47_0 = 0
}

func (obj *Object) PlayerActionPhoneme(cc player.CtrlCode, resetTimer bool) {
	if noxflags.HasGame(noxflags.GameModeChat) {
		return
	}
	switch cc {
	case player.CCSpellGestureUp:
		obj.PlayerSpellPhoneme(spell.PhonUN, sound.SoundSpellPhonemeUp, resetTimer)
	case player.CCSpellGestureDown:
		obj.PlayerSpellPhoneme(spell.PhonZO, sound.SoundSpellPhonemeDown, resetTimer)
	case player.CCSpellGestureLeft:
		obj.PlayerSpellPhoneme(spell.PhonET, sound.SoundSpellPhonemeLeft, resetTimer)
	case player.CCSpellGestureRight:
		obj.PlayerSpellPhoneme(spell.PhonCHA, sound.SoundSpellPhonemeRight, resetTimer)
	case player.CCSpellGestureUpperRight:
		obj.PlayerSpellPhoneme(spell.PhonIN, sound.SoundSpellPhonemeUpRight, resetTimer)
	case player.CCSpellGestureUpperLeft:
		obj.PlayerSpellPhoneme(spell.PhonKA, sound.SoundSpellPhonemeUpLeft, resetTimer)
	case player.CCSpellGestureLowerRight:
		obj.PlayerSpellPhoneme(spell.PhonDO, sound.SoundSpellPhonemeDownRight, resetTimer)
	case player.CCSpellGestureLowerLeft:
		obj.PlayerSpellPhoneme(spell.PhonRO, sound.SoundSpellPhonemeDownLeft, resetTimer)
	}
}
