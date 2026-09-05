package server

const (
	spellGestureCancelPlayerClass4FE680 = uint8(0x04)
	spellGestureCancelResult4FE680      = int32(15)
	spellGestureCancelFizzle4FE680      = int32(231)
	spellGestureCancelState4FE680       = int32(13)
	spellGestureCancelEpsilon4FE680     = float32(0.1)
)

// spellGestureCancelHooks4FE680 exposes every observable read, write, and
// callback in GAME.EXE 004FE680. Pointer-bearing values are generic tokens so
// the semantic core cannot inherit PE32 pointer truncation. Scalar fields keep
// the exact width consumed by the original instruction stream.
type spellGestureCancelHooks4FE680[
	Entity, Object, Team, Update, Player comparable,
	Allocator any,
] struct {
	loadHead      func() Entity
	loadSourceArg func() Object
	loadObject    func(Entity) Object
	loadClass     func(Object) uint32
	loadTeam      func(Object) Team
	compareTeams  func(Team, Team) int32
	loadPosX      func(Object) float32
	loadPosY      func(Object) float32
	loadRadiusArg func() float32
	mapCheck      func(Object, Object) int32

	loadUpdate          func(Object) Update
	storeSpellCastStart func(Update, uint32)
	storeCasting        func(Update, uint8)
	loadPlayer          func(Update) Player
	loadPlayerIndex     func(Player) uint8
	informResult        func(uint8, uint8, int32)
	audioEvent          func(int32, Object, int32, uint32)
	setPlayerState      func(Object, int32)

	loadNext      func(Entity) Entity
	loadPrev      func(Entity) Entity
	storePrev     func(Entity, Entity)
	storeNext     func(Entity, Entity)
	storeHead     func(Entity)
	loadAllocator func() Allocator
	free          func(Allocator, Entity)
}

// spellGestureCancelWithinRadius4FE680 models the unspilled x87 expression.
// Nox runs this code with 53-bit precision control, which is represented by
// the address-specific binary64 helpers shared with Warcry's proximity scan.
// FCOMP is followed by a C0-only test, so unordered values pass just like
// ordered values strictly below the supplied radius.
func spellGestureCancelWithinRadius4FE680(
	targetX, sourceX, targetY, sourceY, radius float32,
) bool {
	distance := warcryProximityDistance4FC4C0(targetX, sourceX, targetY, sourceY)
	withEpsilon := warcryProximityAdd64_4FC4C0(
		distance,
		float64(spellGestureCancelEpsilon4FE680),
	)
	return !(withEpsilon >= float64(radius))
}

// spellGestureCancel4FE680 preserves GAME.EXE 004FE680's exact gates,
// callback order, live object reloads, and intrusive-list unlink sequence.
// In particular, team equality is recognized only for the canonical return
// value one, the map check receives the target cached for distance reads, and
// Next52/Prev56 are deliberately reloaded at every original instruction.
// No nil guards are added at dereference or callback boundaries.
func spellGestureCancel4FE680[
	Entity, Object, Team, Update, Player comparable,
	Allocator any,
](h spellGestureCancelHooks4FE680[
	Entity, Object, Team, Update, Player, Allocator,
]) {
	entity := h.loadHead()
	var nilEntity Entity
	if entity == nilEntity {
		return
	}
	source := h.loadSourceArg()

	for {
		target := h.loadObject(entity)
		if uint8(h.loadClass(target))&spellGestureCancelPlayerClass4FE680 != 0 {
			targetTeam := h.loadTeam(target)
			sourceTeam := h.loadTeam(source)
			if h.compareTeams(sourceTeam, targetTeam) == 1 {
				entity = h.loadNext(entity)
				if entity == nilEntity {
					return
				}
				continue
			}
		}

		target = h.loadObject(entity)
		targetX := h.loadPosX(target)
		sourceX := h.loadPosX(source)
		targetY := h.loadPosY(target)
		sourceY := h.loadPosY(source)
		radius := h.loadRadiusArg()
		if !spellGestureCancelWithinRadius4FE680(targetX, sourceX, targetY, sourceY, radius) ||
			h.mapCheck(source, target) == 0 {
			entity = h.loadNext(entity)
			if entity == nilEntity {
				return
			}
			continue
		}

		target = h.loadObject(entity)
		if uint8(h.loadClass(target))&spellGestureCancelPlayerClass4FE680 != 0 {
			update := h.loadUpdate(target)
			h.storeSpellCastStart(update, 0)
			h.storeCasting(update, 0)
			player := h.loadPlayer(update)
			playerIndex := h.loadPlayerIndex(player)
			h.informResult(playerIndex, 0, spellGestureCancelResult4FE680)

			target = h.loadObject(entity)
			h.audioEvent(spellGestureCancelFizzle4FE680, target, 0, 0)
			target = h.loadObject(entity)
			h.setPlayerState(target, spellGestureCancelState4FE680)
		}

		next := h.loadNext(entity)
		if next != nilEntity {
			prev := h.loadPrev(entity)
			h.storePrev(next, prev)
		}
		prev := h.loadPrev(entity)
		if prev != nilEntity {
			next = h.loadNext(entity)
			h.storeNext(prev, next)
		} else {
			next = h.loadNext(entity)
			h.storeHead(next)
		}
		allocator := h.loadAllocator()
		next = h.loadNext(entity)
		h.free(allocator, entity)
		entity = next
		if entity == nilEntity {
			return
		}
	}
}
