package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

type toxicCloudCastDeps52DB60 struct {
	loadCache  func() uint32
	storeCache func(uint32)
	lookupType func(string) uint32
	traceRay   func(types.Pointf, types.Pointf, MapTraceFlags) bool
	newObject  func(uint32) *Object
	createAt   func(*Object, *Object, types.Pointf)
	lifetime   func(string) float32
	fps        func() uint32
	castSound  func(int32) sound.ID
	audio      func(sound.ID, *Object, int, uint32)
	inform     func(uint8, byte, int32)
}

// toxicCloudLifetime52DB60 models 0052DC3F FIMUL, the single FSTPS spill,
// and 00419A70 FISTP. 00419D40 loads a binary32 balance value; FPS is a
// signed DWORD. Gameplay uses precision 53 and round-toward-zero. Recover
// only the multiplication residual with FMA, without changing thread state.
func toxicCloudLifetime52DB60(lifetime float32, fps uint32) int32 {
	a, b := float64(lifetime), float64(int32(fps))
	product := float64(a * b)
	residual := math.FMA(a, b, -product)
	if product > 0 && residual < 0 || product < 0 && residual > 0 {
		product = math.Nextafter(product, 0)
	}
	value := float32(product)
	if math.Abs(float64(value)) > math.Abs(product) {
		value = math.Nextafter32(value, 0)
	}
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32 // masked x87 invalid conversion's integer indefinite
	}
	return int32(value)
}

// toxicCloudCast52DB60 restores GAME.EXE 0052DB60..0052DC78. Owner,
// caster, SpellAcceptArg, update data and Player pointers are native-width;
// the type cache, lifetime and notification fields keep their DWORD/BYTE
// widths. The second object, argument Obj and level are unused originally.
func toxicCloudCast52DB60(id int32, owner, caster *Object, arg *SpellAcceptArg, h toxicCloudCastDeps52DB60) int32 {
	if h.loadCache() == 0 {
		h.storeCache(h.lookupType("ToxicCloud"))
	}
	from, to := caster.PosVec, arg.Pos
	if !h.traceRay(from, to, MapTraceFlags(9)) {
		if caster.ObjClass.Has(object.ClassPlayer) {
			player := (*PlayerUpdateData)(caster.UpdateData).Player
			h.inform(player.PlayerInd, 0, 2)
		}
		return 0
	}
	// Allocation reloads the live cache after trace. Keep the update pointer
	// cached before placement, but read the live argument position afterward.
	cloud := h.newObject(h.loadCache())
	if cloud != nil {
		data := (*ToxicCloudUpdateData)(cloud.UpdateData)
		h.createAt(cloud, owner, arg.Pos)
		lifetime := h.lifetime("ToxicCloudLifetime")
		fps := h.fps()
		data.Duration = toxicCloudLifetime52DB60(lifetime, fps)
	}
	// Even a failed allocation performs sound lookup/audio and returns 1.
	castSound := h.castSound(id)
	h.audio(castSound, cloud, 0, 0)
	return 1
}

type ToxicCloudCastRuntime52DB60 struct {
	TypeCache *uint32
	CreateAt  func(*Object, *Object, types.Pointf)
}

func (s *Server) CastToxicCloud52DB60(id int32, _ *Object, owner, caster *Object, arg *SpellAcceptArg, _ int32, runtime ToxicCloudCastRuntime52DB60) int32 {
	return toxicCloudCast52DB60(id, owner, caster, arg, toxicCloudCastDeps52DB60{
		loadCache:  func() uint32 { return *runtime.TypeCache },
		storeCache: func(kind uint32) { *runtime.TypeCache = kind },
		lookupType: func(name string) uint32 { return uint32(s.Types.IndByID(name)) },
		traceRay:   s.MapTraceRay,
		newObject:  func(kind uint32) *Object { return s.NewObjectByTypeInd(int(int32(kind))) },
		createAt:   runtime.CreateAt,
		lifetime:   func(key string) float32 { return float32(s.Balance.Float(key)) },
		fps:        s.TickRate,
		castSound:  func(id int32) sound.ID { return s.Spells.DefByInd(spell.ID(id)).GetCastSound() },
		audio:      func(id sound.ID, cloud *Object, kind int, code uint32) { s.Audio.EventObj(id, cloud, kind, code) },
		inform: func(index uint8, code byte, value int32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(value))
		},
	})
}
