package server

const (
	spellDurationDestroyMonsterClass4FEDA0  = byte(2)
	spellDurationDestroyPlayerClass4FEDA0   = byte(4)
	spellDurationDestroyWarrior4FEDA0       = byte(0)
	spellDurationDestroyBerserk4FEDA0       = int32(1)
	spellDurationDestroyPlayerState4FEDA0   = int32(13)
	spellDurationDestroyAudioSelector4FEDA0 = int32(2)
)

// SpellDurationDestroyHooks4FEDA0 exposes every observable record, object,
// player, and callback access in GAME.EXE 004FEDA0. Comparable tokens model
// native null and pointer identity without inheriting the executable's PE32
// pointer width.
type SpellDurationDestroyHooks4FEDA0[Record, Object, PlayerUpdate, Player, Callback comparable] struct {
	LoadCaster func(Record) Object
	LoadSpell  func(Record) uint32
	SpellAudio func(int32, int32) int32
	AudioEvent func(int32, Object, int32, int32)

	LoadDestroy func(Record) Callback
	CallDestroy func(Callback, Record)

	LoadObjectClassLow func(Object) byte
	LoadPlayerUpdate   func(Object) PlayerUpdate
	LoadPlayer         func(PlayerUpdate) Player
	LoadPlayerClass    func(Player) byte
	AbilityActive      func(Object, int32) int32
	SetPlayerState     func(Object, int32)
	MonsterCancel      func(Object, int32)

	Unlink        func(Record)
	FreeRecursive func(Record)
}

// SpellDurationDestroy4FEDA0 preserves GAME.EXE 004FEDA0's exact branch,
// access, and callback order. The sound target is the first caster snapshot.
// The destroy callback is snapshotted after sound delivery, then the caster is
// reloaded for class handling. Player takes precedence over Monster when both
// low class bits are set. Non-Warriors and Warriors without active Berserk are
// reset through one final live caster reload. Monster cancellation receives a
// live spell dword. Every path unlinks before recursively freeing the record;
// no record, object, player, callback, or runtime guard is added.
func SpellDurationDestroy4FEDA0[Record, Object, PlayerUpdate, Player, Callback comparable](
	record Record,
	h SpellDurationDestroyHooks4FEDA0[Record, Object, PlayerUpdate, Player, Callback],
) {
	var nilObject Object
	caster := h.LoadCaster(record)
	if caster != nilObject {
		spellID := int32(h.LoadSpell(record))
		audio := h.SpellAudio(spellID, spellDurationDestroyAudioSelector4FEDA0)
		h.AudioEvent(audio, caster, 0, 0)
	}

	destroy := h.LoadDestroy(record)
	var nilCallback Callback
	if destroy != nilCallback {
		h.CallDestroy(destroy, record)
	}

	caster = h.LoadCaster(record)
	if caster != nilObject {
		class := h.LoadObjectClassLow(caster)
		if class&spellDurationDestroyPlayerClass4FEDA0 != 0 {
			update := h.LoadPlayerUpdate(caster)
			player := h.LoadPlayer(update)
			playerClass := h.LoadPlayerClass(player)
			if playerClass != spellDurationDestroyWarrior4FEDA0 ||
				h.AbilityActive(caster, spellDurationDestroyBerserk4FEDA0) == 0 {
				liveCaster := h.LoadCaster(record)
				h.SetPlayerState(liveCaster, spellDurationDestroyPlayerState4FEDA0)
				h.Unlink(record)
				h.FreeRecursive(record)
				return
			}
		} else if class&spellDurationDestroyMonsterClass4FEDA0 != 0 {
			spellID := int32(h.LoadSpell(record))
			h.MonsterCancel(caster, spellID)
		}
	}

	h.Unlink(record)
	h.FreeRecursive(record)
}
