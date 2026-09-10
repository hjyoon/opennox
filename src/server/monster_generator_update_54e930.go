package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

const (
	monsterGeneratorBlockedFlags54E930 = object.FlagDestroyed | object.FlagDead
	monsterGeneratorDisabled54E930     = uint32(0x800)
	monsterGeneratorNearRadius54EBA0   = float64(300)
	monsterGeneratorSpawnRadius54ED50  = float32(45)
	monsterGeneratorCollisionGap54EF00 = float32(15)
	monsterGeneratorRingStep54EF90     = float64(1.8849558)
)

type monsterGeneratorState54E930 struct {
	hardcoreStage        int32
	hardcoreRateIncrease uint32
	hardcoreSpawnCap     float32
	spawnRates           [5]uint32
	beholderType         uint16
}

// MonsterGeneratorRuntime54E930 supplies the few outer-server and legacy
// services used by the restored MonsterGenerator update. Object and update
// records never cross this boundary as fixed-width integer arguments.
type MonsterGeneratorRuntime54E930 struct {
	QuestStage           func() uint32
	QuestGroup           func() int
	MapTileAllowTeleport func(*types.Pointf) int32
	CreateAt             func(object, owner *Object, position types.Pointf)
	ScriptCallback       func(*ScriptCallback, *Object, *Object, ScriptEventType)
	SendSpawnFX          func([4]int32, int16)
	InventoryPut         func(owner, item *Object, report int32)
	EquipWeapon          func(owner, item *Object) int
	EquipArmor           func(owner, item *Object) int
	QuestHealthFactor    func() float64
	FloatToInt           func(float32) int32
}

type monsterGeneratorNativeDeps54E930 struct {
	questStage      func() uint32
	questGroup      func() int
	balanceFloat    func(string) float32
	floatToInt      func(float32) int32
	frame           func() uint32
	needSync        func(*Object)
	randomInt       func(int, int) int
	randomFloat     func(float32, float32) float64
	firstPlayerUnit func() *Object
	nextPlayerUnit  func(*Object) *Object
	playerByInd     func(ntype.PlayerInd) *Player
	trace           func(types.Pointf, types.Pointf, MapTraceFlags) bool
	occupied        func(types.Pointf) bool
	tileAllow       func(*types.Pointf) int32
	canPlace        func(types.Pointf) bool
	newObject       func(int) *Object
	applyAttrs      func(*Object, *ModifierInitData) bool
	inventoryPut    func(*Object, *Object, int32)
	equipWeapon     func(*Object, *Object) int
	equipArmor      func(*Object, *Object) int
	healthFactor    func() float64
	typeHealth      func(*Object) *HealthData
	typeByName      func(string) int
	register        func(*Object, *Object) bool
	freeObject      func(*Object)
	createAt        func(*Object, *Object, types.Pointf)
	scriptCallback  func(*ScriptCallback, *Object, *Object, ScriptEventType)
	spawnFX         func([4]int32, int16)
	audio           func(sound.ID, *Object)
}

func (s *Server) monsterGeneratorNativeDeps54E930(
	runtime MonsterGeneratorRuntime54E930,
) monsterGeneratorNativeDeps54E930 {
	return monsterGeneratorNativeDeps54E930{
		questStage: runtime.QuestStage,
		questGroup: runtime.QuestGroup,
		balanceFloat: func(name string) float32 {
			return float32(s.Balance.Float(name))
		},
		floatToInt: runtime.FloatToInt,
		frame:      s.Frame,
		needSync: func(obj *Object) {
			obj.NeedSync()
		},
		randomInt: s.Rand.Logic.IntClamp,
		randomFloat: func(minimum, maximum float32) float64 {
			return logicRandomFloat416030(s.Rand.Logic, minimum, maximum)
		},
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit:  s.Players.NextUnit,
		playerByInd:     s.Players.ByInd,
		trace:           s.MapTraceRay,
		occupied: func(point types.Pointf) bool {
			occupied := false
			rect := types.Rectf{
				Min: types.Ptf(point.X-monsterGeneratorCollisionGap54EF00, point.Y-monsterGeneratorCollisionGap54EF00),
				Max: types.Ptf(point.X+monsterGeneratorCollisionGap54EF00, point.Y+monsterGeneratorCollisionGap54EF00),
			}
			s.Map.EachObjInRect(rect, func(candidate *Object) bool {
				if !candidate.Class().Has(object.ClassMonsterGenerator) &&
					candidate.Field5&monsterGeneratorDisabled54E930 == 0 &&
					candidate.Sub547DB0(&point) {
					occupied = true
				}
				return true
			})
			return occupied
		},
		tileAllow: runtime.MapTileAllowTeleport,
		canPlace:  s.MonsterSpawnCanPlace50DE80,
		newObject: s.NewObjectByTypeInd,
		applyAttrs: func(obj *Object, attrs *ModifierInitData) bool {
			return s.ApplyModifierAttrs4E4990(obj, attrs)
		},
		inventoryPut: runtime.InventoryPut,
		equipWeapon:  runtime.EquipWeapon,
		equipArmor:   runtime.EquipArmor,
		healthFactor: runtime.QuestHealthFactor,
		typeHealth: func(obj *Object) *HealthData {
			if typ := obj.ObjectTypeC(); typ != nil {
				return typ.Health()
			}
			return nil
		},
		typeByName: s.Types.IndByID,
		register: func(generator, spawned *Object) bool {
			return s.MonsterSpawnRegister50E030(generator, spawned, MonsterSpawnRegisterRuntime50E030{
				CreateGlyph: func() *Object {
					return s.NewObjectByTypeInd(s.Types.GlyphID())
				},
				InventoryPut: runtime.InventoryPut,
			})
		},
		freeObject: func(obj *Object) {
			s.Objs.FreeObject(obj)
		},
		createAt:       runtime.CreateAt,
		scriptCallback: runtime.ScriptCallback,
		spawnFX:        runtime.SendSpawnFX,
		audio: func(id sound.ID, obj *Object) {
			s.Audio.EventObj(id, obj, 0, 0)
		},
	}
}

func monsterGeneratorLoadBalance54E930(
	state *monsterGeneratorState54E930,
	deps monsterGeneratorNativeDeps54E930,
) {
	if state.hardcoreStage != 0 {
		return
	}
	state.hardcoreStage = deps.floatToInt(deps.balanceFloat("QuestHardcoreStage"))
	state.hardcoreRateIncrease = uint32(deps.floatToInt(deps.balanceFloat("QuestHardcoreSpawnRateIncrease")))
	state.hardcoreSpawnCap = deps.balanceFloat("QuestHardcoreSpawnCap")
	keys := [...]string{
		"SpawnRateHighValue",
		"SpawnRateNormalValue",
		"SpawnRateLowValue",
		"SpawnRateVeryLowValue",
		"SpawnRateVeryVeryLowValue",
	}
	for i, key := range keys {
		state.spawnRates[i] = uint32(deps.floatToInt(deps.balanceFloat(key)))
	}
}

func monsterGeneratorSpawnDelay54E930(
	generator *Object,
	update *MonsterGenUpdateData,
	group int,
	questStage uint32,
	state *monsterGeneratorState54E930,
	deps monsterGeneratorNativeDeps54E930,
) uint32 {
	selector := update.SpawnRate[group]
	var delay uint32
	if int(selector) < len(state.spawnRates) {
		delay = state.spawnRates[selector]
	} else {
		// GAME.EXE falls through with EAX still holding the generator address.
		delay = uint32(uintptr(unsafe.Pointer(generator)))
	}
	if questStage >= uint32(state.hardcoreStage) {
		decrease := state.hardcoreRateIncrease * (questStage - uint32(state.hardcoreStage) + 1)
		adjusted := uint32(0)
		if decrease <= delay {
			adjusted = delay - decrease
		}
		ratio := float64(adjusted) / float64(delay)
		if ratio < float64(state.hardcoreSpawnCap) {
			adjusted = uint32(deps.floatToInt(float32(float64(delay) * float64(state.hardcoreSpawnCap))))
		}
		delay = adjusted
	}
	return delay
}

func monsterGeneratorPick54EBA0(
	generator, template *Object,
	update *MonsterGenUpdateData,
	destination *types.Pointf,
	deps monsterGeneratorNativeDeps54E930,
) bool {
	unit := deps.firstPlayerUnit()
	if unit == nil {
		return false
	}
	var candidates [32]ntype.PlayerInd
	count := 0
	for ; unit != nil; unit = deps.nextPlayerUnit(unit) {
		if unit.Flags().HasAny(monsterGeneratorBlockedFlags54E930) || unit.UpdateData == nil {
			continue
		}
		player := unit.UpdateDataPlayer().Player
		if player == nil || byte(player.Field3680)&1 != 0 {
			continue
		}
		distance := generator.PosVec.Sub(unit.PosVec).Len()
		if update.Field92&1 != 0 {
			if distance <= monsterGeneratorNearRadius54EBA0 {
				return monsterGeneratorSetCreaturePos54ED50(generator, template, nil, destination, update, deps)
			}
			continue
		}
		if update.Field92&2 == 0 {
			return false
		}
		viewportRadius := math.Sqrt(float64(player.Field12)*float64(player.Field12) +
			float64(player.Field10)*float64(player.Field10))
		if distance <= float64(float32(viewportRadius)) &&
			deps.trace(generator.PosVec, unit.PosVec, MapTraceFlags(69)) {
			if count < len(candidates) {
				candidates[count] = ntype.PlayerInd(player.PlayerInd)
				count++
			}
		}
	}
	if count == 0 {
		return false
	}
	player := deps.playerByInd(candidates[deps.randomInt(0, count-1)])
	if player == nil {
		return false
	}
	return monsterGeneratorSetCreaturePos54ED50(generator, template, player.PlayerUnit, destination, update, deps)
}

func monsterGeneratorSetCreaturePos54ED50(
	generator, template, target *Object,
	destination *types.Pointf,
	update *MonsterGenUpdateData,
	deps monsterGeneratorNativeDeps54E930,
) bool {
	traceFlags := MapTraceFlags(1)
	if template.Flags().Has(object.FlagAirborne) {
		traceFlags = MapTraceFlags(5)
	}
	if update.Field92&2 != 0 && target != nil {
		delta := target.PosVec.Sub(generator.PosVec)
		if delta.X == 0 {
			delta.X++
		}
		if delta.Y == 0 {
			delta.Y++
		}
		candidate := generator.PosVec.Add(delta.Normalize().Mul(monsterGeneratorSpawnRadius54ED50))
		if !deps.occupied(candidate) &&
			deps.trace(generator.PosVec, candidate, traceFlags) &&
			deps.tileAllow(&candidate) == 0 {
			*destination = candidate
			direction := DirFromVec(target.PosVec.Sub(candidate))
			template.Direction1 = direction
			template.Direction2 = direction
			return true
		}
	}
	if !monsterGeneratorRandomPosition54EF90(
		monsterGeneratorSpawnRadius54ED50,
		generator.PosVec,
		template,
		destination,
		deps,
	) {
		return false
	}
	if target != nil {
		direction := DirFromVec(target.PosVec.Sub(generator.PosVec))
		template.Direction1 = direction
		template.Direction2 = direction
	}
	return true
}

func monsterGeneratorRandomPosition54EF90(
	radius float32,
	center types.Pointf,
	template *Object,
	destination *types.Pointf,
	deps monsterGeneratorNativeDeps54E930,
) bool {
	angle := float32(deps.randomFloat(-3.1415927, 3.1415927))
	traceFlags := MapTraceFlags(1)
	if template.Flags().Has(object.FlagAirborne) {
		traceFlags = MapTraceFlags(5)
	}
	for attempt := 0; attempt < 32; attempt++ {
		sum := float64(angle) + monsterGeneratorRingStep54EF90
		angle = float32(sum)
		candidate := types.Ptf(
			float32(math.Cos(sum)*float64(radius)+float64(center.X)),
			float32(math.Sin(float64(angle))*float64(radius)+float64(center.Y)),
		)
		if deps.trace(center, candidate, traceFlags) &&
			!deps.occupied(candidate) &&
			deps.tileAllow(&candidate) == 0 {
			*destination = candidate
			return true
		}
	}
	return false
}

func monsterGeneratorCopyCreature54F2B0(
	template, spawned *Object,
	deps monsterGeneratorNativeDeps54E930,
) {
	*spawned.UpdateDataMonster() = *template.UpdateDataMonster()
	if uint8(template.ObjSubClass)&0x10 != 0 {
		for source := template.FirstItem(); source != nil; source = source.NextItem() {
			item := deps.newObject(int(source.TypeInd))
			if item == nil {
				continue
			}
			if uint32(item.ObjClass)&0x13001000 != 0 && source.InitData != nil {
				deps.applyAttrs(item, source.InitDataModifier())
			}
			deps.inventoryPut(spawned, item, 0)
			if source.Flags().Has(object.FlagEquipped) {
				if uint32(item.ObjClass)&0x1001000 != 0 {
					deps.equipWeapon(spawned, item)
				} else if uint32(item.ObjClass)&0x2000000 != 0 {
					deps.equipArmor(spawned, item)
				}
			}
		}
	}
	spawned.Direction1 = template.Direction1
	spawned.Direction2 = template.Direction1
}

func monsterGeneratorSpawn54F070(
	generator, template *Object,
	destination types.Pointf,
	state *monsterGeneratorState54E930,
	deps monsterGeneratorNativeDeps54E930,
) *Object {
	if !deps.canPlace(destination) {
		return nil
	}
	spawned := deps.newObject(int(template.TypeInd))
	if spawned == nil {
		return nil
	}
	monsterGeneratorCopyCreature54F2B0(template, spawned, deps)
	monster := spawned.UpdateDataMonster()
	if spawned.HealthData != nil {
		if monster.MonsterDef != nil {
			spawned.HealthData.Cur = uint16(deps.floatToInt(float32(
				deps.healthFactor() * float64(int32(monster.MonsterDef.HealthQuest72)),
			)))
			spawned.HealthData.Max = uint16(deps.floatToInt(float32(
				deps.healthFactor() * float64(int32(monster.MonsterDef.HealthQuest72)),
			)))
		} else if health := deps.typeHealth(spawned); health != nil {
			spawned.HealthData.Cur = uint16(deps.floatToInt(float32(
				deps.healthFactor() * float64(health.Cur),
			)))
			spawned.HealthData.Max = uint16(deps.floatToInt(float32(
				deps.healthFactor() * float64(health.Max),
			)))
		}
		if spawned.HealthData.Cur == 0 {
			spawned.HealthData.Cur = 1
		}
		if spawned.HealthData.Max == 0 {
			spawned.HealthData.Max = 1
		}
	}
	if state.beholderType == 0 {
		state.beholderType = uint16(deps.typeByName("Beholder"))
	}
	if spawned.TypeInd == state.beholderType {
		monster.Direction94 = 0
	}
	if !deps.register(generator, spawned) {
		deps.freeObject(spawned)
		return nil
	}
	deps.createAt(spawned, nil, destination)
	update := generator.UpdateDataMonsterGen()
	callback := (*ScriptCallback)(unsafe.Pointer(&update.Field64))
	deps.scriptCallback(callback, spawned, generator, NoxEventGeneratorSpawn)

	delta := destination.Sub(generator.PosVec).Normalize()
	fx := [4]int32{
		deps.floatToInt(generator.PosVec.X),
		deps.floatToInt(generator.PosVec.Y) - 50,
		deps.floatToInt(delta.X*30 + destination.X),
		deps.floatToInt(delta.Y*30 + destination.Y),
	}
	deps.spawnFX(fx, 10)
	deps.audio(sound.SoundMonsterGeneratorSpawn, spawned)
	return spawned
}

func monsterGeneratorUpdateNative54E930(
	generator *Object,
	state *monsterGeneratorState54E930,
	deps monsterGeneratorNativeDeps54E930,
) {
	questStage := deps.questStage()
	update := generator.UpdateDataMonsterGen()
	group := deps.questGroup()
	monsterGeneratorLoadBalance54E930(state, deps)
	if group < 0 || group >= len(update.SpawnRate) ||
		generator.Field5&monsterGeneratorDisabled54E930 != 0 ||
		!generator.Flags().Has(object.FlagEnabled) ||
		generator.Flags().HasAny(monsterGeneratorBlockedFlags54E930) {
		return
	}
	deps.needSync(generator)
	frame := deps.frame()
	delay := monsterGeneratorSpawnDelay54E930(generator, update, group, questStage, state, deps)
	if frame-update.Frame88 <= delay || update.ActiveCount >= update.MaxActive || byte(frame)&8 == 0 {
		return
	}
	base := group * 4
	count := 0
	for i := 0; i < 4; i++ {
		if update.Field0[base+i] != nil {
			count++
		}
	}
	if count == 0 {
		return
	}
	template := update.Field0[base+deps.randomInt(0, count-1)]
	if template == nil {
		return
	}
	var destination types.Pointf
	if monsterGeneratorPick54EBA0(generator, template, update, &destination, deps) {
		monsterGeneratorSpawn54F070(generator, template, destination, state, deps)
		update.Frame88 = deps.frame()
	}
}

// MonsterGeneratorUpdate54E930 restores GAME.EXE 0054E930 through 0054F2B0
// with native-width Object, MonsterGenUpdateData, MonsterUpdateData, inventory,
// and spawn-list pointers.
func (s *Server) MonsterGeneratorUpdate54E930(
	generator *Object,
	runtime MonsterGeneratorRuntime54E930,
) {
	monsterGeneratorUpdateNative54E930(
		generator,
		&s.monsterGenerator54E930,
		s.monsterGeneratorNativeDeps54E930(runtime),
	)
}
