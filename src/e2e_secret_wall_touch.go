package opennox

import (
	"fmt"
	"image"
	"math"
	"sort"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2eSecretWallTouchOutcome struct {
	beforeState, beforeDelay      byte
	contact, opening, moved       bool
	full, passable, stable, bound bool
	crossed, clientCrossed        [2]bool
	progress                      uint32
	sounds                        int
}

func (o e2eSecretWallTouchOutcome) validate() error {
	// Every natural delay from 1 through 23 must be observed. Delay zero may
	// precede the E2E predicate in the frame that collision starts opening.
	const progress = uint32(0xFFFFFE)
	if o.beforeState != 1 || o.beforeDelay != 0 || !o.contact || !o.opening || !o.moved ||
		!o.full || !o.passable || !o.stable || !o.bound || o.sounds != 1 ||
		o.progress&progress != progress || o.progress&^uint32(0xFFFFFF) != 0 ||
		o.crossed != [2]bool{true, true} || o.clientCrossed != [2]bool{true, true} {
		return fmt.Errorf("secret-wall touch: before=%d/%d contact=%t opening=%t moved=%t full=%t passable=%t stable=%t bound=%t progress=%#x sounds=%d crossed=%v client=%v",
			o.beforeState, o.beforeDelay, o.contact, o.opening, o.moved, o.full, o.passable, o.stable,
			o.bound, o.progress, o.sounds, o.crossed, o.clientCrossed)
	}
	return nil
}

func e2eSecretWallTouchGeometry(grid image.Point, direction byte, radius float32) (center, start, end, normal types.Pointf, err error) {
	if grid.X < 0 || grid.Y < 0 || grid.X >= server.WallGridSize || grid.Y >= server.WallGridSize ||
		(grid.X+grid.Y)%2 != 0 || direction > 1 || radius <= 0 || radius > 32 || math.IsNaN(float64(radius)) {
		return center, start, end, normal, fmt.Errorf("invalid secret-wall lane: grid=%v dir=%d radius=%g", grid, direction, radius)
	}
	center = types.Ptf(float32(23*grid.X)+11.5, float32(23*grid.Y)+11.5)
	normal = types.Ptf(1, 1).Normalize()
	if direction == 1 {
		normal.Y = -normal.Y
	}
	start, end = center.Sub(normal.Mul(radius+36)), center.Add(normal.Mul(radius+36))
	return center, start, end, normal, nil
}

// Retain the wide approach when it fits; a nearby parallel stock wall may
// require the shorter approach. Both candidates keep the same clear-circle
// checks and reach the existing radius+16 full-crossing threshold.
func e2eSecretWallTouchLane(grid image.Point, direction byte, radius float32,
	traceClear, laneClear func(types.Pointf, types.Pointf) bool,
) (center, start, end, normal types.Pointf, err error) {
	center, start, end, normal, err = e2eSecretWallTouchGeometry(grid, direction, radius)
	if err != nil {
		return
	}
	if traceClear == nil || laneClear == nil {
		err = fmt.Errorf("secret-wall lane requires both read-only clearance checks")
		return
	}
	for _, distance := range []float32{radius + 36, radius + 16} {
		start, end = center.Sub(normal.Mul(distance)), center.Add(normal.Mul(distance))
		if traceClear(start, end) ||
			!laneClear(start, center.Sub(normal.Mul(radius+8))) ||
			!laneClear(end, center.Add(normal.Mul(radius+8))) {
			continue
		}
		clear := true
		for around := 0; around < 256; around += 16 {
			cx, cy := server.SinCosDir(byte(around))
			offset := types.Ptf((radius+4)*cx, (radius+4)*cy)
			if !laneClear(start, start.Add(offset)) || !laneClear(end, end.Add(offset)) {
				clear = false
				break
			}
		}
		if clear {
			return center, start, end, normal, nil
		}
	}
	err = fmt.Errorf("secret wall has no clear closed approach: grid=%v dir=%d radius=%g", grid, direction, radius)
	return
}

type e2eSecretWallTouchFixture struct {
	unit                         *server.Object
	wall                         *server.Wall
	secret                       *server.SecretWall
	before                       server.SecretWall
	grid                         image.Point
	direction, tile              byte
	original, center, start, end types.Pointf
	normal                       types.Pointf
	frame, lastLog               uint32
	active                       bool
	outcome                      e2eSecretWallTouchOutcome
}

// Select a closed, non-timed, touch-enabled wall loaded from the stock map.
// Neither the map nor any wall state/flags are supplied by this fixture.
func (f *e2eSecretWallTouchFixture) prepare() {
	f.unit = noxServer.Players.HostUnit()
	if f.unit == nil || f.unit.UpdateData == nil || f.unit.ControllingPlayer() == nil ||
		f.unit.Shape.Kind != server.ShapeKindCircle || e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) != "g_crypts" {
		e2eError(fmt.Errorf("secret-wall touch requires a real connected G_Crypts player"))
		return
	}
	f.original = f.unit.PosVec
	walls := noxServer.Walls.All()
	sort.SliceStable(walls, func(i, j int) bool {
		return walls[i].Pos().Sub(f.original).Len() < walls[j].Pos().Sub(f.original).Len()
	})
	radius := f.unit.Shape.Circle.R
	for _, wall := range walls {
		secret := wall.Secret()
		if secret == nil || secret.State != 1 || secret.OpenDelay != 0 || secret.Flags&2 == 0 || secret.Flags&4 != 0 ||
			secret.Wall != wall || wall.Dir0 > 1 {
			continue
		}
		center, start, end, normal, err := e2eSecretWallTouchLane(wall.GridPos(), wall.Dir0, radius,
			func(from, to types.Pointf) bool { return noxServer.MapTraceRay(from, to, server.MapTraceFlag1) },
			func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.unit, from, to) })
		if err != nil {
			continue
		}
		f.wall, f.secret, f.before = wall, secret, *secret
		f.grid, f.direction, f.tile = wall.GridPos(), wall.Dir0, wall.Tile1
		f.center, f.start, f.end, f.normal = center, start, end, normal
		break
	}
	if f.wall == nil {
		for _, wall := range walls {
			if secret := wall.Secret(); secret != nil {
				e2eLog.Printf("SECRET WALL CANDIDATE: grid=%v dir=%d flags=%#x state=%d delay=%d wait=%d", wall.GridPos(),
					wall.Dir0, secret.Flags, secret.State, secret.OpenDelay, secret.OpenWait)
			}
		}
		e2eError(fmt.Errorf("G_Crypts has no closed touch-enabled stock wall with a clear lane on both sides"))
		return
	}
	definition := noxServer.Walls.DefByInd(int(f.wall.Tile1))
	if definition == nil || sound.ByName(definition.OpenSound()) == 0 {
		e2eError(fmt.Errorf("stock secret wall lacks its tile open sound"))
		return
	}
	wantSound := sound.ByName(definition.OpenSound())
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
		if f.active && id == wantSound && kind == 0 && owner == nil && pos == f.center {
			f.outcome.sounds++
		}
	})
	// Only place the player in an unobstructed starting circle. All subsequent
	// contact and crossing must come from queued ordinary movement input.
	asObjectS(f.unit).SetPos(f.start)
	e2eLog.Printf("SECRET WALL PREPARED: wall=%p secret=%p unit=%p grid=%v dir=%d flags=%#x start=%v end=%v sound=%s",
		f.wall, f.secret, f.unit, f.wall.GridPos(), f.wall.Dir0, f.secret.Flags, f.start, f.end, wantSound)
}

func (f *e2eSecretWallTouchFixture) binding() bool {
	if f.wall == nil || f.unit == nil || f.unit.UpdateData == nil || noxClient.ClientPlayerUnit() == nil ||
		f.wall.Secret() != f.secret || noxServer.Walls.GetWallAtGrid(f.wall.GridPos()) != f.wall ||
		f.wall.GridPos() != f.grid || f.wall.Dir0 != f.direction || f.wall.Tile1 != f.tile ||
		int(f.secret.X) != f.grid.X || int(f.secret.Y) != f.grid.Y ||
		f.secret.Next != f.before.Next || f.secret.Wall != f.before.Wall || f.secret.X != f.before.X || f.secret.Y != f.before.Y ||
		f.secret.Flags != f.before.Flags || f.secret.OpenWait != f.before.OpenWait {
		return false
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.wall), unsafe.Pointer(f.secret), unsafe.Pointer(f.unit), f.unit.UpdateData} {
			if uintptr(ptr) <= math.MaxUint32 {
				return false
			}
		}
	}
	return true
}

func (f *e2eSecretWallTouchFixture) baseline() bool {
	if !f.binding() || f.unit.PosVec.Sub(f.start).Len() > 2 ||
		types.Ptf(float32(noxClient.ClientPlayerUnit().PosVec.X), float32(noxClient.ClientPlayerUnit().PosVec.Y)).Sub(f.start).Len() > 3 {
		return false
	}
	f.outcome.beforeState, f.outcome.beforeDelay = f.secret.State, f.secret.OpenDelay
	if f.secret.State != 1 || f.secret.OpenDelay != 0 || noxServer.MapTraceRay(f.start, f.end, server.MapTraceFlag1) {
		e2eError(fmt.Errorf("secret wall changed or became passable before actual movement"))
		return true
	}
	e2eLog.Printf("SECRET WALL BASELINE PASS: state=1 delay=0 blocked=true frame=%d", noxServer.Frame())
	return true
}

func (f *e2eSecretWallTouchFixture) input(sign float32) {
	if !f.binding() {
		e2eError(fmt.Errorf("secret wall lost its native binding before input"))
		return
	}
	f.active = true
	f.frame, f.lastLog = noxServer.Frame(), 0
	aim := f.center.Add(f.normal.Mul(112 * sign))
	mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(aim.X), int(aim.Y)))
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
	e2eLog.Printf("SECRET WALL INPUT: sign=%g frame=%d mouse=%v player=%v", sign, f.frame, mouse, f.unit.PosVec)
}

func (f *e2eSecretWallTouchFixture) observeOpening() bool {
	if !f.binding() || f.secret.OpenDelay > 23 || f.outcome.sounds > 1 {
		e2eError(fmt.Errorf("secret-wall binding, delay, or duplicate sound failed during contact"))
		return true
	}
	f.outcome.moved = f.outcome.moved || f.unit.PosVec.Sub(f.start).Len() > 5
	f.outcome.contact = f.outcome.contact || f.unit.UpdateDataPlayer().CollisionWall == f.wall
	if f.secret.State == 4 || f.secret.State == 3 {
		f.outcome.progress |= uint32(1) << f.secret.OpenDelay
	}
	f.outcome.opening = f.outcome.opening || f.secret.State == 4
	f.outcome.passable = f.outcome.passable || f.secret.OpenDelay > 11 &&
		noxServer.Sub_57B500(f.wall.GridPos(), 64) == -1 && noxServer.MapTraceRay(f.start, f.end, server.MapTraceFlag1)
	if f.secret.State == 3 && f.secret.OpenDelay == 23 {
		f.outcome.full = true
		if !f.outcome.contact || !f.outcome.opening || !f.outcome.moved || !f.outcome.passable ||
			f.outcome.progress&0xFFFFFE != 0xFFFFFE || f.outcome.sounds != 1 {
			e2eError(fmt.Errorf("secret wall opened without the complete real contact/progress/sound trace: %+v", f.outcome))
			return true
		}
		e2eLog.Printf("SECRET WALL OPEN PASS: state=3 delay=23 contact=true progress=%#x sounds=%d elapsed=%d", f.outcome.progress,
			f.outcome.sounds, noxServer.Frame()-f.frame)
		return true
	}
	if f.lastLog == 0 || noxServer.Frame()-f.lastLog >= 10 {
		f.lastLog = noxServer.Frame()
		e2eLog.Printf("SECRET WALL WAIT: state=%d delay=%d contact=%t progress=%#x sounds=%d collision=%p player=%v elapsed=%d",
			f.secret.State, f.secret.OpenDelay, f.outcome.contact, f.outcome.progress, f.outcome.sounds,
			f.unit.UpdateDataPlayer().CollisionWall, f.unit.PosVec, noxServer.Frame()-f.frame)
	}
	return false
}

func (f *e2eSecretWallTouchFixture) observeCrossing(index int, sign float32) bool {
	if !f.binding() || f.secret.State != 3 || f.secret.OpenDelay != 23 || f.outcome.sounds != 1 {
		e2eError(fmt.Errorf("secret wall closed, duplicated sound, or lost binding during crossing"))
		return true
	}
	projection := func(pos types.Pointf) float32 {
		delta := pos.Sub(f.center)
		return sign * (delta.X*f.normal.X + delta.Y*f.normal.Y)
	}
	drawable := noxClient.ClientPlayerUnit()
	clientPos := types.Ptf(float32(drawable.PosVec.X), float32(drawable.PosVec.Y))
	threshold := f.unit.Shape.Circle.R + 16
	f.outcome.crossed[index] = projection(f.unit.PosVec) >= threshold
	f.outcome.clientCrossed[index] = projection(clientPos) >= threshold
	if !f.outcome.crossed[index] || !f.outcome.clientCrossed[index] {
		return false
	}
	e2eLog.Printf("SECRET WALL CROSS PASS: sign=%g server=%v client=%v state=3 delay=23 sounds=%d", sign, f.unit.PosVec, clientPos, f.outcome.sounds)
	return true
}

func (f *e2eSecretWallTouchFixture) observeStable() bool {
	f.outcome.bound = f.binding()
	f.outcome.stable = f.outcome.bound && f.secret.State == 3 && f.secret.OpenDelay == 23 && f.outcome.sounds == 1 &&
		noxServer.Sub_57B500(f.wall.GridPos(), 64) == -1 && noxServer.MapTraceRay(f.start, f.end, server.MapTraceFlag1)
	if err := f.outcome.validate(); err != nil {
		e2eError(err)
		return true
	}
	f.active = false
	e2eLog.Printf("SECRET WALL STABLE PASS: real contact, natural opening, server/client round-trip and one sound; wall=%p secret=%p", f.wall, f.secret)
	return true
}

func (sc *e2eScenario) CheckSecretWallTouch(name string) {
	f := &e2eSecretWallTouchFixture{}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(20, name+" normal position replication")
	sc.addWhen(0, name+" closed baseline", 120, f.baseline, func() {})
	sc.add(0, name+" actual contact input", func() { f.input(1) })
	sc.addWhen(1, name+" natural opening", 180, f.observeOpening, func() {})
	sc.addWhen(0, name+" outward crossing", 120, func() bool { return f.observeCrossing(0, 1) }, func() {})
	sc.Input(0, name+" release outward input", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	sc.Wait(8, name+" outward settle")
	sc.add(0, name+" actual return input", func() { f.input(-1) })
	sc.addWhen(1, name+" return crossing", 180, func() bool { return f.observeCrossing(1, -1) }, func() {})
	sc.Input(0, name+" release return input", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	sc.Wait(8, name+" return settle")
	sc.addWhen(0, name+" stable open wall", 120, f.observeStable, func() {})
	sc.add(0, name+" restore fixture position", func() { asObjectS(f.unit).SetPos(f.original) })
}
