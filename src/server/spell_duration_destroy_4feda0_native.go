package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
)

// SpellDurationDestroyRuntime4FEDA0 supplies the operations still owned by
// the outer game runtime. Callback, duration-record, and object identities
// remain native-width; player state retains its original byte domain.
type SpellDurationDestroyRuntime4FEDA0 struct {
	CallDestroy    func(unsafe.Pointer, *DurSpell)
	SetPlayerState func(*Object, PlayerState)
}

type spellDurationDestroyNativeDeps4FEDA0 struct {
	spellAudio     func(int32, int32) int32
	audioEvent     func(int32, *Object, int32, int32)
	callDestroy    func(unsafe.Pointer, *DurSpell)
	abilityActive  func(*Object, int32) int32
	setPlayerState func(*Object, int32)
	monsterCancel  func(*Object, int32)
	unlink         func(*DurSpell)
	freeRecursive  func(*DurSpell)
}

func spellDurationDestroyNative4FEDA0(
	record *DurSpell,
	deps spellDurationDestroyNativeDeps4FEDA0,
) {
	SpellDurationDestroy4FEDA0(record, SpellDurationDestroyHooks4FEDA0[
		*DurSpell,
		*Object,
		*PlayerUpdateData,
		*Player,
		unsafe.Pointer,
	]{
		LoadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		LoadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		SpellAudio: deps.spellAudio,
		AudioEvent: deps.audioEvent,
		LoadDestroy: func(record *DurSpell) unsafe.Pointer {
			return record.Destroy
		},
		CallDestroy: deps.callDestroy,
		LoadObjectClassLow: func(object *Object) byte {
			return byte(object.ObjClass)
		},
		LoadPlayerUpdate: func(object *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(object.UpdateData)
		},
		LoadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		LoadPlayerClass: func(player *Player) byte {
			// PlayerClass is deliberately nil-safe elsewhere in the port. The
			// executable directly loads Player+2251 and therefore faults.
			if player == nil {
				panic("004FEDA0: nil Player")
			}
			return byte(player.Info().PlayerClass())
		},
		AbilityActive:  deps.abilityActive,
		SetPlayerState: deps.setPlayerState,
		MonsterCancel:  deps.monsterCancel,
		Unlink:         deps.unlink,
		FreeRecursive:  deps.freeRecursive,
	})
}

func spellDurationDestroyServerDeps4FEDA0(
	sp *SpellsDuration,
	runtime SpellDurationDestroyRuntime4FEDA0,
) spellDurationDestroyNativeDeps4FEDA0 {
	return spellDurationDestroyNativeDeps4FEDA0{
		spellAudio: func(spellID, selector int32) int32 {
			return int32(sp.s.Spells.DefByInd(spell.ID(spellID)).GetAudio(int(selector)))
		},
		audioEvent: func(id int32, object *Object, kind, code int32) {
			sp.s.Audio.EventObj(sound.ID(id), object, int(kind), uint32(code))
		},
		callDestroy: runtime.CallDestroy,
		abilityActive: func(object *Object, ability int32) int32 {
			if sp.s.Abils.IsActive(object, Ability(ability)) {
				return 1
			}
			return 0
		},
		setPlayerState: func(object *Object, state int32) {
			runtime.SetPlayerState(object, PlayerState(state))
		},
		monsterCancel: func(object *Object, spellID int32) {
			object.MonsterCancelDurSpell(spell.ID(spellID))
		},
		unlink:        sp.SpellDurationUnlink4FE900,
		freeRecursive: sp.SpellDurationFreeRecursive4FE980,
	}
}

// SpellDurationDestroy4FEDA0 binds GAME.EXE 004FEDA0 to native-width
// duration records, objects, player data, and callback pointers. The adapter
// deliberately preserves all original nil-dereference boundaries and live
// field reloads instead of adding Go-level guards. Fixed-width spell, audio,
// ability, callback-result, and state values retain their dword/byte domains.
// There is no retained independent C ABI because both known callers are now
// owned by the Go duration-spell runtime.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationDestroy4FEDA0(
	record *DurSpell,
	runtime SpellDurationDestroyRuntime4FEDA0,
) {
	spellDurationDestroyNative4FEDA0(
		record,
		spellDurationDestroyServerDeps4FEDA0(sp, runtime),
	)
}
