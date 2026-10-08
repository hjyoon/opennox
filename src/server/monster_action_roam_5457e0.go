package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterActionRoamHooks5457E0 struct {
	frame            func() uint32
	tickRate         func() uint32
	random           func(int, int) int
	noticeThreat     func() int
	interestingSound func() int
	nearestWaypoint  func(types.Pointf, byte) *Waypoint
	fallbackWaypoint func(*Object, *types.Pointf) *Waypoint
	setDetailedPath  func(*Object, *types.Pointf)
	actuallyMove     func(*Object) bool
	moveAudio        func(*Object)
	push             func(ai.ActionType, ...any) *AIStackItem
	pop              func() int
}

func monsterWaypointValid547EE0(waypoint *Waypoint, mask byte) bool {
	return waypoint != nil && waypoint.IsEnabled() && waypoint.HasFlag2Mask(mask)
}

func monsterRoamRecordWaypoint545B00(update *MonsterUpdateData, waypoint *Waypoint) {
	update.Field91++
	if update.Field91 >= uint32(len(update.Waypoints)) {
		update.Field91 = 0
	}
	index := int(update.Field91)
	update.Waypoints[index] = waypoint
	for i := range update.Waypoints {
		if i != index && update.Waypoints[i] == waypoint {
			update.Waypoints[i] = nil
		}
	}
}

func monsterRoamPreviousWaypoint545B60(update *MonsterUpdateData, mask byte) *Waypoint {
	for distance := 1; distance < len(update.Waypoints); distance++ {
		index := int(update.Field91) - distance
		if index < 0 {
			index += len(update.Waypoints)
		}
		waypoint := update.Waypoints[index]
		if monsterWaypointValid547EE0(waypoint, mask) {
			return waypoint
		}
	}
	return nil
}

func monsterRoamNextWaypoint545C60(update *MonsterUpdateData, current *Waypoint, mask byte, random func(int, int) int) *Waypoint {
	if current == nil {
		return nil
	}
	count := int(current.PointsCnt)
	if count > len(current.Points) {
		count = len(current.Points)
	}
	choices := make([]*Waypoint, 0, count)
	for i := 0; i < count; i++ {
		candidate := current.Points[i].Waypoint
		if !monsterWaypointValid547EE0(candidate, mask) {
			continue
		}
		seen := false
		for _, waypoint := range update.Waypoints {
			if candidate == waypoint {
				seen = true
				break
			}
		}
		if !seen {
			choices = append(choices, candidate)
		}
	}
	if len(choices) != 0 {
		return choices[random(0, len(choices)-1)]
	}

	for distance := 0; distance < len(update.Waypoints); distance++ {
		index := (int(update.Field91) + distance + 1) % len(update.Waypoints)
		candidate := update.Waypoints[index]
		if !monsterWaypointValid547EE0(candidate, mask) {
			continue
		}
		for i := 0; i < count; i++ {
			if current.Points[i].Waypoint == candidate {
				return candidate
			}
		}
	}
	return nil
}

func monsterRoamDeadEnd545BB0(unit *Object, current *Waypoint, mask byte, random func(int, int) int, pop func() int) bool {
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil {
		return false
	}
	update.Field2 = 0
	if current != nil && current.PointsCnt != 0 {
		next := monsterRoamNextWaypoint545C60(update, current, mask, random)
		head.Args[0] = uintptr(unsafe.Pointer(next))
		if next != nil {
			monsterRoamRecordWaypoint545B00(update, next)
			return true
		}
	}
	pop()
	return false
}

// MonsterActionRoamStart545790 restores the state changes made by GAME.EXE
// 00545790 without reading the native-width update-data pointer at PE32 offset
// 748. The original routine clears only the first history slot and its cursor.
func (s *Server) MonsterActionRoamStart545790(unit *Object) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	update.Waypoints[0] = nil
	update.Field91 = 0
}

// MonsterActionRoamCancel5457C0 clears the first argument of the active ROAM
// action using the native-width AI stack representation.
func (s *Server) MonsterActionRoamCancel5457C0(unit *Object) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	if head := unit.UpdateDataMonster().AIStackHead(); head != nil {
		head.Args[0] = 0
	}
}

func monsterActionRoam5457E0(unit *Object, hooks monsterActionRoamHooks5457E0) {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return
	}
	update := unit.UpdateDataMonster()
	head := update.AIStackHead()
	if head == nil || head.Type() != ai.ACTION_ROAM {
		return
	}

	if unit.Sub_5343C0() || unit.Nox_xxx_monsterCanAttackAtWill_534390() {
		if hooks.noticeThreat() != 0 {
			return
		}
	}
	if update.StatusFlags.Has(object.MonStatusBot) && update.CurrentEnemy == nil &&
		unit.HasEnchant(ENCHANT_INVISIBLE) && byte(hooks.frame())&0x1f == 0 && hooks.random(0, 100) < 10 {
		hooks.push(ai.DEPENDENCY_ENEMY_FARTHER_THAN, float32(150))
		hooks.push(ai.DEPENDENCY_IS_ENCHANTED, uint32(0))
		hooks.push(ai.ACTION_WAIT, hooks.frame()+hooks.tickRate()*uint32(hooks.random(3, 10)))
		return
	}
	if unit.Nox_xxx_monsterCanAttackAtWill_534390() {
		if enemy := update.CurrentEnemy; enemy != nil {
			hooks.push(ai.ACTION_FIGHT, enemy.PosVec, hooks.frame())
			return
		}
		if hooks.interestingSound() != 0 {
			return
		}
	}

	mask := byte(head.ArgU32(2))
	waypoint := (*Waypoint)(unsafe.Pointer(head.Args[0]))
	if !monsterWaypointValid547EE0(waypoint, mask) {
		waypoint = nil
		head.Args[0] = 0
	}
	if waypoint == nil {
		waypoint = hooks.nearestWaypoint(unit.PosVec, mask)
		if waypoint == nil && int8(mask) == -128 {
			waypoint = hooks.fallbackWaypoint(unit, &unit.PosVec)
		}
		head.Args[0] = uintptr(unsafe.Pointer(waypoint))
		if waypoint == nil {
			hooks.pop()
			hooks.push(ai.ACTION_WAIT, hooks.frame()+uint32(hooks.random(int(hooks.tickRate()), 2*int(hooks.tickRate()))))
			return
		}
		monsterRoamRecordWaypoint545B00(update, waypoint)
	}

	delta := waypoint.PosVec.Sub(unit.PosVec)
	if float64(delta.X*delta.X+delta.Y*delta.Y) <= 64.0 {
		if !monsterRoamDeadEnd545BB0(unit, waypoint, mask, hooks.random, hooks.pop) {
			return
		}
		update.Field70 = 0
		waypoint = (*Waypoint)(unsafe.Pointer(head.Args[0]))
	}
	if update.Field2 == 0 {
		hooks.setDetailedPath(unit, &waypoint.PosVec)
	}
	if byte(update.Field71) == 2 {
		previous := monsterRoamPreviousWaypoint545B60(update, mask)
		if previous == nil {
			hooks.pop()
			return
		}
		if !monsterRoamDeadEnd545BB0(unit, previous, mask, hooks.random, hooks.pop) {
			return
		}
	}
	if hooks.actuallyMove(unit) {
		update.Field2 = 0
	}
	hooks.moveAudio(unit)
}

func monsterCreatureActuallyMove50D3B0(unit *Object, trace func(types.Pointf, types.Pointf, MapTraceFlags) bool) bool {
	update := unit.UpdateDataMonster()
	if update.Field2 == 0 {
		return false
	}
	// 0050D3CC caches the trace origin, but 0050D42A reads the live unit
	// position after each callback. Retain the entry update-data pointer too.
	origin := unit.PosVec
	pathCount := int(int32(update.Field2))
	if pathCount > len(update.Path) {
		pathCount = len(update.Path)
	}
	start := int(int32(update.Field67))
	if start >= pathCount || start < 0 {
		update.Field2 = 0
		return false
	}

	selected := -1
	closePoint := -1
	bestDistance := float32(10000000.0)
	// The original signed loop reloads Field2 after every ray trace, including
	// rejected rays. Preserve the native array bound for malformed counts.
	for i := start; i < len(update.Path) && i < int(int32(update.Field2)); i++ {
		point := update.Path[i]
		if !trace(origin, point, MapTraceFlags(132)) {
			continue
		}
		// 0050D42A..0050D480 retains precision-53 / ToZero differences,
		// Y-square then X-square and the sum; only the best distance spills.
		dx := monsterMoveToRunAddChop53_544434(float64(point.X), -float64(unit.PosVec.X))
		dy := monsterMoveToRunAddChop53_544434(float64(point.Y), -float64(unit.PosVec.Y))
		ySquare := monsterMoveToRunSquareChop53_544434(dy)
		xSquare := monsterMoveToRunSquareChop53_544434(dx)
		distance := monsterMoveToRunAddChop53_544434(ySquare, xSquare)
		if distance > 64.0 {
			if selected < 0 || float64(bestDistance) > distance {
				bestDistance = float32(monsterMoveToRunSpill544440(distance))
				selected = i
			}
			continue
		}
		if i == int(int32(update.Field2))-1 {
			update.Field2 = 0
			return true
		}
		closePoint = i
	}
	if selected < 0 {
		if closePoint < 0 {
			update.Field2 = 0
			return false
		}
		selected = closePoint
	}

	// 0050D4DA publishes the scalar debug cursor before the cached path cursor.
	*memmap.PtrUint32(0x5D4594, 2386204) = uint32(selected)
	update.Field67 = uint32(selected)
	// 0050D4EF spills X, but 0050D4FA retains Y after also spilling it.
	// The length uses retained-Y * spilled-Y, not two spilled squares.
	dx := monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(float64(update.Path[selected].X), -float64(unit.PosVec.X)))
	dyRetained := monsterMoveToRunAddChop53_544434(float64(update.Path[selected].Y), -float64(unit.PosVec.Y))
	dy := monsterMoveToRunSpill544440(dyRetained)
	ySquare := monsterMoveForceMulChop53_50D4FE(dyRetained, dy)
	xSquare := monsterMoveToRunSquareChop53_544434(dx)
	square := monsterMoveToRunAddChop53_544434(ySquare, xSquare)
	root := monsterMoveForceSqrtChop53_50D50C(square)
	distance := monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(root, monsterMoveDistanceBias50D3B0))
	// Slot zero deliberately uses slot one even for a one-point path.
	from, to := update.Path[0], update.Path[1]
	if selected > 0 {
		from, to = update.Path[selected-1], update.Path[selected]
	}
	segment := types.Ptf(
		float32(monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(float64(to.X), -float64(from.X)))),
		float32(monsterMoveToRunSpill544440(monsterMoveToRunAddChop53_544434(float64(to.Y), -float64(from.Y)))),
	)
	// 0050D549 calls 00509ED0 in the gameplay precision-53 / ToZero
	// environment. The shared DirFromVec models default nearest-even;
	// using it here rounds the final 00419A70 FISTP up for some segments.
	angle := math.Atan2(float64(segment.Y), float64(segment.X))
	angle = monsterMoveToRunAddChop53_544434(angle, float64(math.Float32frombits(boomDirectionTauBits509ED0)))
	angle = monsterMoveForceMulChop53_50D4FE(angle, float64(math.Float32frombits(boomDirectionScaleBits509ED0)))
	angle = monsterMoveToRunAddChop53_544434(angle, float64(math.Float32frombits(boomDirectionHalfBits509ED0)))
	value := monsterMoveToRunSpill544440(angle)
	direction := Dir16(0)
	if value >= -2147483648.0 && value < 2147483648.0 {
		// The original modulo-256 suffix also maps integer-indefinite to 0.
		direction = Dir16(uint8(int32(value)))
	}
	unit.Direction1 = direction
	unit.Direction2 = direction
	speed := float64(unit.SpeedCur)
	if update.StatusFlags.Has(object.MonStatusRunning) && update.MonsterDef != nil {
		speed = monsterMoveForceMulChop53_50D4FE(float64(update.MonsterDef.RunMultiplier96), speed)
	}
	unit.ForceVec.X = float32(monsterMoveToRunSpill544440(monsterMoveForceDivChop53_50D581(monsterMoveForceMulChop53_50D4FE(speed, dx), distance)))
	unit.ForceVec.Y = float32(monsterMoveToRunSpill544440(monsterMoveForceDivChop53_50D581(monsterMoveForceMulChop53_50D4FE(speed, dy), distance)))
	return false
}

func (s *Server) monsterMoveAudio534030(unit *Object) {
	update := unit.UpdateDataMonster()
	def := update.MonsterDef
	if def == nil || update.SoundSet122 == nil {
		return
	}
	play := false
	if uint32(unit.ObjSubClass)&0x30 != 0 {
		frames, delay := s.PlayerAnimFrames(4)
		if frames > 0 {
			cur := (int(unit.NetCode) + int(s.Frame())) / (delay + 1) % frames
			prev := (int(unit.NetCode) + int(s.Frame()) - 1) / (delay + 1) % frames
			play = cur != prev && (uint32(cur) == def.MoveSndFrameA100 || uint32(cur) == def.MoveSndFrameB104)
		}
	} else {
		play = (uint32(update.Field120_1) == def.MoveSndFrameA100 || uint32(update.Field120_1) == def.MoveSndFrameB104) && update.Field120_2 == 0
	}
	if play {
		id := sound.ID(*(*uint32)(unsafe.Add(update.SoundSet122, 72)))
		s.Audio.EventObj(id, unit, 0, 0)
	}
}

// MonsterActionRoam5457E0 binds the native-width restoration of GAME.EXE
// 005457E0 to the live pathfinder and AI services.
func (s *Server) MonsterActionRoam5457E0(unit *Object, interestingSound func(*Object) int, fallbackWaypoint func(*Object, *types.Pointf) *Waypoint, setDetailedPath func(*Object, *types.Pointf)) {
	monsterActionRoam5457E0(unit, monsterActionRoamHooks5457E0{
		frame:            s.Frame,
		tickRate:         s.TickRate,
		random:           s.Rand.Logic.IntClamp,
		noticeThreat:     unit.Sub_545E60,
		interestingSound: func() int { return interestingSound(unit) },
		nearestWaypoint: func(pos types.Pointf, mask byte) *Waypoint {
			return s.Sub_518460(pos, mask, true)
		},
		fallbackWaypoint: fallbackWaypoint,
		setDetailedPath:  setDetailedPath,
		actuallyMove: func(unit *Object) bool {
			return monsterCreatureActuallyMove50D3B0(unit, s.MapTraceRay)
		},
		moveAudio: s.monsterMoveAudio534030,
		push:      unit.MonsterPushAction,
		pop:       unit.MonsterPopAction,
	})
}
