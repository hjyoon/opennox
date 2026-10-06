package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

type burnCastDeps52C3E0 struct {
	loadGlyph  func() uint32
	storeGlyph func(uint32)
	loadFlame  func() uint32
	storeFlame func(uint32)
	lookupType func(string) uint32
	traceRay   func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject  func(uint32) *Object
	createAt   func(*Object, *Object, types.Pointf)
	balance    func(string) float64
	setDecay   func(*Object, uint32)
	spark      func(types.Pointf, byte)
	castSound  func(int32) sound.ID
	audio      func(sound.ID, types.Pointf, int, uint32)
	inform     func(uint8, byte, int32)
}

// burnCastDuration52C3E0 retains the gameplay x87 round-toward-zero mode
// set at 0043E2C1. The binary32 spill at 0052C4D4 and FISTPL at 00419A77
// both use that mode; NaN and out-of-range conversions produce 80000000.
func burnCastDuration52C3E0(value float64) uint32 {
	value = monsterMoveToRunSpill544440(value)
	if math.IsNaN(value) || value >= 2147483648 || value < -2147483648 {
		return 0x80000000
	}
	return uint32(int32(value))
}

// burnCast52C3E0 restores GAME.EXE 0052C3E0..0052C521. Object, argument,
// update-data and Player pointers stay native-width; cache, duration and
// notification values retain DWORD/BYTE widths. The second/third objects,
// argument Obj and spell level are deliberately unused by the original.
func burnCast52C3E0(id int32, caster *Object, arg *SpellAcceptArg, h burnCastDeps52C3E0) int32 {
	glyph := h.loadGlyph()
	if glyph == 0 {
		glyph = h.lookupType("Glyph")
		h.storeGlyph(glyph)
	}
	// Cache initialization precedes both guards; the argument guard is first.
	if arg == nil || caster == nil {
		return 0
	}
	var position *types.Pointf
	if uint32(caster.TypeInd) == glyph {
		position = &caster.PosVec
	} else {
		from, to := caster.PosVec, arg.Pos
		if !h.traceRay(from, to, MapTraceFlags(9)) {
			// The class and Player chain are read after the trace callback.
			if caster.ObjClass.Has(object.ClassPlayer) {
				player := (*PlayerUpdateData)(caster.UpdateData).Player
				h.inform(player.PlayerInd, 0, 2)
			}
			return 0
		}
		position = &arg.Pos
	}
	flameKind := h.loadFlame()
	if flameKind == 0 {
		flameKind = h.lookupType("MediumFlame")
		h.storeFlame(flameKind)
	}
	// Both cache values are local DWORDs, not reloaded after their lookups.
	// The selected coordinate pointer is dereferenced only after allocation.
	if flame := h.newObject(flameKind); flame != nil {
		h.createAt(flame, caster, *position)
		duration := burnCastDuration52C3E0(h.balance("BurnDuration"))
		h.setDecay(flame, duration)
		h.spark(flame.PosVec, 64)
	}
	// Allocation failure still plays the cast sound and returns success.
	// Sound position is always the live argument position, including Glyph.
	castSound := h.castSound(id)
	h.audio(castSound, arg.Pos, 0, 0)
	return 1
}

type BurnCastRuntime52C3E0 struct {
	GlyphTypeCache *uint32
	FlameTypeCache *uint32
	CreateAt       func(*Object, *Object, types.Pointf)
}

func (s *Server) CastBurn52C3E0(id int32, _ *Object, _ *Object, caster *Object, arg *SpellAcceptArg, _ int32, runtime BurnCastRuntime52C3E0) int32 {
	return burnCast52C3E0(id, caster, arg, burnCastDeps52C3E0{
		loadGlyph:  func() uint32 { return *runtime.GlyphTypeCache },
		storeGlyph: func(kind uint32) { *runtime.GlyphTypeCache = kind },
		loadFlame:  func() uint32 { return *runtime.FlameTypeCache },
		storeFlame: func(kind uint32) { *runtime.FlameTypeCache = kind },
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		traceRay:   s.MapTraceRay,
		newObject:  func(kind uint32) *Object { return s.NewObjectByTypeInd(int(int32(kind))) },
		createAt:   runtime.CreateAt,
		balance:    s.Balance.Float,
		setDecay:   func(obj *Object, delay uint32) { s.DecaySetTime511660(obj, delay) },
		spark:      s.Nox_xxx_netSparkExplosionFx_5231B0,
		castSound:  func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:      s.Audio.EventPos,
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
