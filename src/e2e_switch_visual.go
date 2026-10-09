package opennox

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"sort"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Only ordinary map-script Enable calls supply inputs. The stock switch update,
// network draw-frame packet, client SlaveDraw and raster supply all results.
func (sc *e2eScenario) CheckSpikeSwitchVisual(name string) {
	var host, spike *server.Object
	var original types.Pointf
	var data server.DamageCollideData
	var health uint16
	var code uint16
	var originallyEnabled bool
	var hashes [2][32]byte
	var seen [2]bool
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil
	}, func() {
		host = noxServer.Players.HostUnit()
		original, health = host.PosVec, host.HealthData.Cur
		var candidates []*server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.ObjectTypeC().ID() == "Spike" && obj.CollideData != nil && obj.Extent != 0 && obj.Shape.Kind == server.ShapeKindCircle {
				candidates = append(candidates, obj)
			}
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].PosVec.Sub(original).Len() < candidates[j].PosVec.Sub(original).Len()
		})
		var observer types.Pointf
		for _, candidate := range candidates {
			for dir := 0; dir < 256; dir += 16 {
				x, y := server.SinCosDir(byte(dir))
				pos := candidate.PosVec.Add(types.Ptf(80*x, 80*y))
				if !noxServer.MapTraceRay(candidate.PosVec, pos, server.MapTraceFlag1) {
					continue
				}
				clear := true
				for around := 0; around < 256; around += 16 {
					x, y := server.SinCosDir(byte(around))
					if !e2eWarriorLaneClear(host, pos, pos.Add(types.Ptf((host.Shape.Circle.R+4)*x, (host.Shape.Circle.R+4)*y))) {
						clear = false
						break
					}
				}
				if clear {
					spike, observer = candidate, pos
					break
				}
			}
			if spike != nil {
				break
			}
		}
		if spike == nil {
			e2eError(fmt.Errorf("map has no original Spike with a safe observer: candidates=%d", len(candidates)))
			return
		}
		originallyEnabled = spike.IsEnabled()
		asObjectS(host).SetPos(observer)
		asObjectS(spike).Enable(false)
		code, data = uint16(noxServer.GetUnitNetCode(spike)), *(*server.DamageCollideData)(spike.CollideData)
		e2eLog.Printf("SWITCH SPIKE PREPARED: original-map-object=%p extent=%d wire=%#x world=%v observer=%v HP=%d stock-data=%+v", spike, spike.Extent, code, spike.PosVec, observer, health, data)
	})
	for step, enabled := range []bool{false, true, true, false, true, false} {
		sc.add(8, fmt.Sprintf("%s step%d input", name, step), func() { asObjectS(spike).Enable(enabled) })
		sc.addWhen(8, fmt.Sprintf("%s step%d live frame", name, step), 600, func() bool {
			dr := noxClient.Objs.ByNetCode(code)
			want := uint32(1)
			if enabled {
				want = 0
			}
			if dr == nil || dr.AnimFrameSlave != want || dr.Flags().Has(object.FlagEnabled) != enabled {
				if noxServer.Frame()%120 == 0 {
					e2eLog.Printf("SWITCH SPIKE WAIT: step=%d wire=%#x server-flags=%#x drawable=%p", step, code, spike.Flags(), dr)
					if dr != nil {
						e2eLog.Printf("SWITCH SPIKE CLIENT WAIT: frame=%d flags=%#x want=%d/%t", dr.AnimFrameSlave, dr.Flags(), want, enabled)
					}
				}
			}
			return dr != nil && dr.AnimFrameSlave == want && spike.IsEnabled() == enabled &&
				spike.Flags().Has(object.FlagNoCollide) != enabled && dr.Flags().Has(object.FlagEnabled) == enabled
		}, func() {
			dr := noxClient.Objs.ByNetCode(code)
			if host.HealthData.Cur != health || *(*server.DamageCollideData)(spike.CollideData) != data ||
				spike.Collide != spike.ObjectTypeC().Collide || spike.Update != spike.ObjectTypeC().Update ||
				dr.DrawFuncPtr == nil || dr.DrawData == nil {
				e2eError(fmt.Errorf("switch visual changed stock damage/update/draw data or injured observer"))
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(spike), spike.CollideData, unsafe.Pointer(dr), dr.DrawData} {
					if uintptr(ptr) <= math.MaxUint32 {
						e2eError(fmt.Errorf("switch visual lost native pointer %p", ptr))
						return
					}
				}
			}
			pix, err := e2eSwitchSpikeRaster(dr)
			if err != nil {
				e2eError(err)
				return
			}
			packed := make([]byte, len(pix.Pix)*2)
			pixels, bounds := 0, image.Rectangle{}
			for y := pix.Rect.Min.Y; y < pix.Rect.Max.Y; y++ {
				for x := pix.Rect.Min.X; x < pix.Rect.Max.X; x++ {
					index := pix.PixOffset(x, y)
					binary.LittleEndian.PutUint16(packed[2*index:], pix.Pix[index])
					if pix.Pix[index] != 0 {
						pixels++
						bounds = bounds.Union(image.Rect(x, y, x+1, y+1))
					}
				}
			}
			hash, frame := sha256.Sum256(packed), int(dr.AnimFrameSlave)
			if pixels == 0 || seen[frame] && hashes[frame] != hash || seen[1-frame] && hashes[1-frame] == hash {
				e2eError(fmt.Errorf("stock Spike pixels unchanged/unstable: step=%d frame=%d pixels=%d SHA256=%x previous=%x", step, frame, pixels, hash, hashes))
				return
			}
			if !seen[frame] {
				path, err := e2eWriteMagicFrame("", pix.SubImage(bounds.Inset(-8).Intersect(pix.Rect)))
				if err != nil {
					e2eError(err)
					return
				}
				e2eLog.Printf("SWITCH SPIKE CAPTURE: enabled=%t frame=%d path=%s", enabled, frame, path)
			}
			hashes[frame], seen[frame] = hash, true
			e2eLog.Printf("SWITCH SPIKE VERIFIED: step=%d enabled=%t server-collidable=%t client-frame=%d pixels=%d SHA256=%x HP=%d stock-data=unchanged", step, enabled, !spike.Flags().Has(object.FlagNoCollide), frame, pixels, hash, health)
		})
	}
	sc.add(1, name+" cleanup", func() {
		asObjectS(spike).Enable(originallyEnabled)
		asObjectS(host).SetPos(original)
	})
}

func e2eSwitchSpikeRaster(dr *client.Drawable) (*noximage.Image16, error) {
	live, data := noxClient.r.PixBuffer(), noxClient.r.Data()
	before := *data
	pix := noximage.NewImage16(live.Rect)
	noxClient.r.SetPixBuffer(pix)
	defer func() { noxClient.r.SetPixBuffer(live); *data = before }()
	if dr.CallDraw(noxClient.Viewport()) == 0 {
		return nil, fmt.Errorf("live stock Spike SlaveDraw returned no draw")
	}
	return pix, nil
}

// GAME.EXE 00472600/004730D0: at zoom 100 a diagonal wall is 23 pixels.
// The interior avoids the shared endpoints of adjacent stock map walls.
func e2eMinimapWallSamples(grid image.Point, direction byte, center, size image.Point) ([]image.Point, error) {
	if direction > 1 || size.X <= 0 || size.Y <= 0 || grid.X < 0 || grid.Y < 0 ||
		grid.X >= server.WallGridSize || grid.Y >= server.WallGridSize || (grid.X+grid.Y)&1 != 0 {
		return nil, fmt.Errorf("invalid minimap wall projection: grid=%v dir=%d size=%v", grid, direction, size)
	}
	width := size.X / 6
	top := (size.Y - width) / 2
	rect := image.Rect(0, top, width, top+width)
	origin := center.Sub(image.Pt(width/2, width/2))
	anchor := image.Pt(grid.X*23-origin.X, top+grid.Y*23-origin.Y)
	points := make([]image.Point, 0, 16)
	for offset := 4; offset < 20; offset++ {
		p := anchor.Add(image.Pt(offset, offset))
		if direction == 0 {
			p.Y = anchor.Y + 23 - offset
		}
		if !p.In(rect.Inset(3)) {
			return nil, fmt.Errorf("minimap wall projection outside interior: grid=%v point=%v rect=%v center=%v", grid, p, rect, center)
		}
		points = append(points, p)
	}
	return points, nil
}

func e2eMinimapWallRaster(grid image.Point, direction byte, visible bool, label string) error {
	if sub_473670() == 0 {
		return fmt.Errorf("minimap was not opened through normal Tab input")
	}
	legacy.SetMinimapZoom(100)
	live, data := noxClient.r.PixBuffer(), noxClient.r.Data()
	before := *data
	points, err := e2eMinimapWallSamples(grid, direction, noxClient.Viewport().World.Max, live.Rect.Size())
	if err != nil {
		return err
	}
	pix := noximage.NewImage16(live.Rect)
	noxClient.r.SetPixBuffer(pix)
	defer func() { noxClient.r.SetPixBuffer(live); *data = before }()
	if !e2eClientDrawMinimap(noxClient.ClientPlayerUnit()) {
		return fmt.Errorf("actual C minimap pass unavailable")
	}
	color := uint16(memmap.Uint32(0x85B3FC, 956))
	count := 0
	for _, p := range points {
		if pix.Pix[pix.PixOffset(p.X, p.Y)] == color {
			count++
		}
	}
	want := 0
	if visible {
		want = len(points)
	}
	if count != want || data.Clip() != before.Clip() || data.ClipRect() != before.ClipRect() || data.ClipRect2() != before.ClipRect2() {
		return fmt.Errorf("minimap wall pixels: %s grid=%v dir=%d color=%#x count=%d want=%d", label, grid, direction, color, count, want)
	}
	path, err := e2eWriteMagicFrame("", pix.SubImage(image.Rect(0, (live.Rect.Dy()-live.Rect.Dx()/6)/2, live.Rect.Dx()/6, (live.Rect.Dy()+live.Rect.Dx()/6)/2)))
	if err != nil {
		return err
	}
	e2eLog.Printf("SWITCH MINIMAP VERIFIED: phase=%s grid=%v dir=%d visible=%t wall-pixels=%d/%d native-C=true path=%s", label, grid, direction, visible, count, len(points), path)
	return nil
}

// Select original stock map secret walls. Enable/Toggle are the same methods
// used by map scripts; natural wall updates must complete both animations.
func (sc *e2eScenario) CheckWallSwitchMinimap(direction int, name string) {
	f := &e2eSecretWallTouchFixture{}
	sc.addWhen(0, name+" select stock wall", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil
	}, func() {
		f.unit = noxServer.Players.HostUnit()
		f.original = f.unit.PosVec
		walls := noxServer.Walls.All()
		sort.SliceStable(walls, func(i, j int) bool {
			return walls[i].Pos().Sub(f.original).Len() < walls[j].Pos().Sub(f.original).Len()
		})
		for _, w := range walls {
			s := w.Secret()
			if int(w.Dir0) != direction || s == nil || s.State != 1 || s.OpenDelay != 0 || s.Flags&4 != 0 || s.Wall != w {
				continue
			}
			center, start, end, normal, err := e2eSecretWallTouchLane(w.GridPos(), w.Dir0, f.unit.Shape.Circle.R,
				func(from, to types.Pointf) bool { return noxServer.MapTraceRay(from, to, server.MapTraceFlag1) },
				func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.unit, from, to) })
			if err != nil {
				continue
			}
			f.wall, f.secret, f.before = w, s, *s
			f.grid, f.direction, f.tile = w.GridPos(), w.Dir0, w.Tile1
			f.center, f.start, f.end, f.normal = center, start, end, normal
			break
		}
		if f.wall == nil {
			e2eError(fmt.Errorf("no original closed non-timed secret wall with clear lane: direction=%d", direction))
			return
		}
		asObjectS(f.unit).SetPos(f.start)
		e2eLog.Printf("SWITCH MINIMAP PREPARED: wall=%p secret=%p grid=%v dir=%d state=%d flags=%#x", f.wall, f.secret, f.grid, f.direction, f.secret.State, f.secret.Flags)
	})
	sc.addWhen(20, name+" settled", 600, f.baseline, nil)
	check := func(state byte, visible bool, phase string) {
		if !f.binding() || f.secret.State != state {
			e2eError(fmt.Errorf("secret wall native binding/state changed: phase=%s state=%d want=%d", phase, f.secret.State, state))
			return
		}
		if err := e2eMinimapWallRaster(f.grid, f.direction, visible, phase); err != nil {
			e2eError(err)
		}
	}
	sc.add(0, name+" closed baseline", func() { check(1, true, "closed") })
	for cycle := 0; cycle < 2; cycle++ {
		sc.add(1, fmt.Sprintf("%s cycle%d open", name, cycle), func() {
			asWallS(f.wall).Enable(false)
			check(4, true, fmt.Sprintf("cycle%d opening", cycle))
		})
		sc.addWhen(1, fmt.Sprintf("%s cycle%d opened", name, cycle), 180, func() bool {
			return f.secret.State == 3 && f.secret.OpenDelay == 23
		}, func() {
			if !noxServer.MapTraceRay(f.start, f.end, server.MapTraceFlag1) {
				e2eError(fmt.Errorf("open secret wall still blocks actual map trace"))
				return
			}
			check(3, false, fmt.Sprintf("cycle%d open", cycle))
		})
		if cycle == 0 {
			sc.Key(keybind.KeyTab, name+" close minimap")
			sc.addWhen(2, name+" map hidden", 120, func() bool { return sub_473670() == 0 }, nil)
			sc.Key(keybind.KeyTab, name+" reopen minimap")
			sc.addWhen(2, name+" map visible", 120, func() bool { return sub_473670() != 0 }, func() { check(3, false, "reopened minimap") })
		}
		sc.add(1, fmt.Sprintf("%s cycle%d close", name, cycle), func() {
			asWallS(f.wall).Enable(true)
			check(2, false, fmt.Sprintf("cycle%d closing", cycle))
		})
		sc.addWhen(1, fmt.Sprintf("%s cycle%d closed", name, cycle), 180, func() bool {
			return f.secret.State == 1 && f.secret.OpenDelay == 0
		}, func() {
			if noxServer.MapTraceRay(f.start, f.end, server.MapTraceFlag1) {
				e2eError(fmt.Errorf("closed secret wall no longer blocks actual map trace"))
				return
			}
			check(1, true, fmt.Sprintf("cycle%d closed", cycle))
		})
	}
	sc.add(0, name+" ordinary wall lifecycle", func() {
		if err := e2eMinimapDynamicWallLifecycle(f); err != nil {
			e2eError(err)
		}
	})
	sc.add(0, name+" restore observer", func() { asObjectS(f.unit).SetPos(f.original) })
}

func e2eMinimapDynamicWallLifecycle(f *e2eSecretWallTouchFixture) error {
	var created []*server.Wall
	defer func() {
		for _, w := range created {
			if noxServer.Walls.GetWallAtGridRaw(w.GridPos()) == w {
				noxServer.Walls.DeleteAtGrid(w.GridPos())
			}
		}
	}()
	beforeCount := len(noxServer.Walls.All())
	// Empty grid cells are inputs, never deleted stock map walls. Three new
	// walls ensure the first deletion is not the global list head.
	center := noxClient.Viewport().World.Max
	for y := center.Y/23 - 3; y <= center.Y/23+3 && len(created) < 3; y++ {
		for x := center.X/23 - 3; x <= center.X/23+3 && len(created) < 3; x++ {
			grid := image.Pt(x, y)
			if (x+y)&1 != 0 || noxServer.Walls.GetWallAtGridRaw(grid) != nil {
				continue
			}
			if _, err := e2eMinimapWallSamples(grid, 1, center, noxClient.r.PixBuffer().Rect.Size()); err != nil {
				continue
			}
			if err := e2eMinimapWallRaster(grid, 1, false, "empty grid"); err != nil {
				continue
			}
			w := noxServer.Walls.CreateAtGrid(grid)
			if w == nil {
				return fmt.Errorf("normal wall pool could not create at %v", grid)
			}
			w.Tile1, w.Dir0 = f.tile, 1
			created = append(created, w)
			if err := e2eMinimapWallRaster(grid, 1, true, "created wall"); err != nil {
				return err
			}
		}
	}
	if len(created) != 3 || len(noxServer.Walls.All()) != beforeCount+3 {
		return fmt.Errorf("dynamic wall fixture lacks three separate visible empty cells: count=%d", len(created))
	}
	for i, w := range created {
		grid := w.GridPos()
		noxServer.Walls.DeleteAtGrid(grid)
		if noxServer.Walls.GetWallAtGridRaw(grid) != nil || len(noxServer.Walls.All()) != beforeCount+2-i {
			return fmt.Errorf("dynamic wall delete left stale grid/global index at %v", grid)
		}
		if err := e2eMinimapWallRaster(grid, 1, false, "deleted wall"); err != nil {
			return err
		}
		reused := noxServer.Walls.CreateAtGrid(grid)
		if reused != w {
			return fmt.Errorf("normal wall free list did not reuse the deleted native wall")
		}
		reused.Tile1, reused.Dir0 = f.tile, 1
		if err := e2eMinimapWallRaster(grid, 1, true, "recreated wall"); err != nil {
			return err
		}
		noxServer.Walls.DeleteAtGrid(grid)
		if err := e2eMinimapWallRaster(grid, 1, false, "removed again"); err != nil {
			return err
		}
	}
	e2eLog.Printf("SWITCH MINIMAP LIFECYCLE VERIFIED: original=%d created=3 deleted=3 reused=3 grid/global/row-indices=consistent", beforeCount)
	return nil
}
