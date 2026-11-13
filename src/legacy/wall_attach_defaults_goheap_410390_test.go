//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"image"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func waExtraCall410390(t *testing.T, backend string, drawable unsafe.Pointer, x, y int32) unsafe.Pointer {
	t.Helper()
	if backend == "Go" {
		return sub410390Native(drawable, x, y)
	}
	if backend == "C" {
		return wallAttachFixtureCall410390(drawable, x, y)
	}
	t.Fatal("wall-attach extra backend")
	return nil
}

func waExtraSnapshot410390(pointer unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(pointer), int(size))...)
}

func waExtraWitness410390(t *testing.T, pointer unsafe.Pointer) {
	t.Helper()
	w := twoSettingsPointerFixturePointerWitness40A740(pointer)
	if !w.Above4GiB || !w.Low32LosesBits || w.Native != uintptr(pointer) {
		t.Fatal("actual C full-pointer witness; numeric low32 loss only")
	}
}

func waExtraLayout410390(t *testing.T) {
	t.Helper()
	offsets, layout := wallAttachNativeOffsets410390(), wallAttachFixtureCLayout410390()
	widths := [5]uintptr{unsafe.Sizeof(int(0)), unsafe.Sizeof(int(0)), 1, unsafe.Sizeof(unsafe.Pointer(nil)), 1}
	if wallAttachFixturePointerWidth410390() != unsafe.Sizeof(unsafe.Pointer(nil)) {
		t.Fatal("actual C/Go pointer widths")
	}
	for i := range offsets {
		if layout[i].Offset != offsets[i] || layout[i].Width != widths[i] {
			t.Fatalf("named C/Go offset/width field%d", i)
		}
	}
}

type waExtraServer410390 struct {
	Server
	native *server.Server
}

func (s *waExtraServer410390) S() *server.Server { return s.native }

// The actual linked polygon functions both have bounded empty-data paths when
// their existing C count is1. The original helper instructions are not credited.
func waExtraWorld410390(t *testing.T) *server.Server {
	t.Helper()
	if wallAttachNativeObserverCurrent410390() != nil {
		t.Fatal("default fixture starts with nil observer")
	}
	previous := GetServer
	srv := server.NewHigh(nil, nil, strman.New())
	if srv.Walls.InitHigh() == 0 {
		srv.Close()
		t.Fatal("actual wall pool initialization")
	}
	GetServer = func() Server { return &waExtraServer410390{native: srv} }
	var count unsafe.Pointer
	var oldCount uint32
	changedCount := false
	t.Cleanup(func() {
		if changedCount {
			*(*uint32)(count) = oldCount
		}
		srv.Walls.Free()
		srv.Close()
		GetServer = previous
	})
	count = wallAttachFixturePolygonCountCell410390()
	waExtraWitness410390(t, count)
	oldCount = *(*uint32)(count)
	*(*uint32)(count) = 1
	changedCount = true
	return srv
}

func waExtraDrawable410390(t *testing.T) *client.Drawable {
	t.Helper()
	drawable, free := alloc.NewHigh(client.Drawable{})
	*drawable = client.Drawable{}
	drawable.PosVec = image.Pt(31, -44)
	drawable.Field_0, drawable.Field_1 = 0x12345678, 0xfedcba98
	t.Cleanup(free)
	waExtraWitness410390(t, unsafe.Pointer(drawable))
	return drawable
}

func waExtraUntouchedWall410390(t *testing.T, before []byte, wall *server.Wall) {
	t.Helper()
	after := waExtraSnapshot410390(unsafe.Pointer(wall), unsafe.Sizeof(*wall))
	offsets := wallAttachNativeOffsets410390()
	for i := range before {
		x := uintptr(i)
		selected := x == offsets[2] || x == offsets[4] || offsets[3] <= x && x < offsets[3]+unsafe.Sizeof(unsafe.Pointer(nil))
		if !selected && before[i] != after[i] {
			t.Fatalf("untouched full wall allocation byte%d", i)
		}
	}
}

func TestWallAttachDefaultLinked410390(t *testing.T) {
	waExtraLayout410390(t)
	srv := waExtraWorld410390(t)
	ordinal := 0
	for _, backend := range []string{"Go", "C"} {
		for _, control := range []string{"lookup-present", "create-missing"} {
			t.Run(backend+"/"+control, func(t *testing.T) {
				ordinal++
				grid := image.Pt(11+ordinal, 21+ordinal)
				drawable := waExtraDrawable410390(t)
				beforeDrawable := waExtraSnapshot410390(unsafe.Pointer(drawable), unsafe.Sizeof(*drawable))
				var present *server.Wall
				var beforeWall []byte
				if control == "lookup-present" {
					present = srv.Walls.CreateAtGrid(grid)
					if present == nil {
						t.Fatal("actual pre-existing wall")
					}
					present.Flags4, present.Field8 = 0x80, 0xa6ff
					beforeWall = waExtraSnapshot410390(unsafe.Pointer(present), unsafe.Sizeof(*present))
				} else if srv.Walls.GetWallAtGridRaw(grid) != nil {
					t.Fatal("actual create fixture starts empty")
				}
				t.Cleanup(func() {
					if wall := srv.Walls.GetWallAtGridRaw(grid); wall != nil {
						wall.ClientData = nil
					}
				})
				if wallAttachNativeObserverCurrent410390() != nil {
					t.Fatal("actual nil-observer default entry")
				}
				got := waExtraCall410390(t, backend, unsafe.Pointer(drawable), int32(grid.X), int32(grid.Y))
				wall := srv.Walls.GetWallAtGridRaw(grid)
				if got != nil || wall == nil || present != nil && wall != present || wall.ClientData != unsafe.Pointer(drawable) || byte(wall.Flags4)&0x10 == 0 || byte(wall.Field8) != 1 {
					t.Fatal("actual linked lookup/create and both empty polygon defaults")
				}
				waExtraWitness410390(t, unsafe.Pointer(wall))
				if present != nil {
					if wall.Field8 != 0xa601 || byte(wall.Flags4) != 0x90 {
						t.Fatal("actual preserved wall scalar canaries")
					}
					waExtraUntouchedWall410390(t, beforeWall, wall)
				}
				if !bytes.Equal(beforeDrawable, waExtraSnapshot410390(unsafe.Pointer(drawable), unsafe.Sizeof(*drawable))) {
					t.Fatal("whole default drawable remains unchanged")
				}
				t.Logf("WALL_ATTACH_DEFAULT_LINKED_OK backend=%s control=%s observer=nil full-native-stored-pointer=true actual-C-empty-query-fallback=true source-bound-linked-services=true helper-original-body-credit=0", backend, control)
			})
		}
	}
}

type waExtraForeign410390 struct {
	Drawable *client.Drawable
	Wall     *server.Wall
	Polygon  unsafe.Pointer
}

func waExtraForeignObjects410390(t *testing.T) waExtraForeign410390 {
	t.Helper()
	drawable := waExtraDrawable410390(t)
	wall, freeWall := alloc.NewHigh(server.Wall{})
	*wall = server.Wall{}
	wall.Flags4, wall.Field8 = 0x80, 0xabcd
	polygon, freePolygon := alloc.MallocHigh(160)
	for i := range unsafe.Slice((*byte)(polygon), 160) {
		*(*byte)(unsafe.Add(polygon, i)) = byte(i*17 + 9)
	}
	*(*byte)(unsafe.Add(polygon, 130)) = 0xd7
	t.Cleanup(func() { wall.ClientData = nil; freePolygon(); freeWall() })
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(drawable), unsafe.Pointer(wall), polygon} {
		waExtraWitness410390(t, pointer)
	}
	return waExtraForeign410390{drawable, wall, polygon}
}

func waExtraForeignReply410390(t *testing.T, objects waExtraForeign410390, site uint32, args []wallAttachNativeArg410390) wallAttachNativeReply410390 {
	t.Helper()
	switch site {
	case 0x4103a0:
		return wallAttachNativeReply410390{Pointer: unsafe.Pointer(objects.Wall)}
	case 0x4103e3:
		if len(args) != 2 || !args[0].IsPointer || args[0].Pointer == nil || args[1].Word != 0 || *(*[2]int32)(args[0].Pointer) != [2]int32{31, -44} {
			t.Fatal("actual current local8 in reentrant query")
		}
		if objects.Wall.ClientData != unsafe.Pointer(objects.Drawable) || byte(objects.Wall.Flags4) != 0x90 {
			t.Fatal("pointer and flags already stored before query")
		}
		return wallAttachNativeReply410390{Pointer: objects.Polygon}
	default:
		t.Fatalf("unexpected foreign-control service %#x", site)
		return wallAttachNativeReply410390{}
	}
}

func TestWallAttachObserverControls410390(t *testing.T) {
	waExtraLayout410390(t)
	srv := waExtraWorld410390(t)
	ordinal := 0
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend+"/partial-decline", func(t *testing.T) {
			ordinal++
			grid := image.Pt(51+ordinal, 61+ordinal)
			drawable := waExtraDrawable410390(t)
			poison, freePoison := alloc.MallocHigh(160)
			defer freePoison()
			counts := map[uint32]int{}
			var local unsafe.Pointer
			observer := &wallAttachNativeObserver410390{}
			observer.Service = func(site uint32, args []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool) {
				counts[site]++
				if site == 0x4103a0 {
					// The handled lookup still uses the actual production helper.
					return wallAttachNativeServiceDefault410390(site, args), true
				}
				if site == 0x4103e3 || site == 0x4103fa {
					wall := srv.Walls.GetWallAtGridRaw(grid)
					if wall == nil || wall.ClientData != unsafe.Pointer(drawable) || byte(wall.Flags4)&0x10 == 0 {
						t.Fatal("early pointer/flags visible before declined actual query")
					}
					if len(args) < 2 || !args[0].IsPointer || args[0].Pointer == nil || *(*[2]int32)(args[0].Pointer) != [2]int32{31, -44} {
						t.Fatal("actual local8 passed to declined service")
					}
					if local != nil && local != args[0].Pointer {
						t.Fatal("query/fallback reuse the same actual local")
					}
					local = args[0].Pointer
				}
				// A nonnil ignored reply proves handled=false selects defaults.
				return wallAttachNativeReply410390{Pointer: poison, Word: 0x80000001}, false
			}
			pop := pushWallAttachNativeObserver410390(observer)
			defer pop()
			t.Cleanup(func() {
				if wall := srv.Walls.GetWallAtGridRaw(grid); wall != nil {
					wall.ClientData = nil
				}
			})
			got := waExtraCall410390(t, backend, unsafe.Pointer(drawable), int32(grid.X), int32(grid.Y))
			if got != nil || local == nil || len(counts) != 4 {
				t.Fatal("declined services reached real defaults")
			}
			for _, site := range []uint32{0x4103a0, 0x4103b0, 0x4103e3, 0x4103fa} {
				if counts[site] != 1 {
					t.Fatal("one actual visit per default service")
				}
			}
			t.Logf("WALL_ATTACH_PARTIAL_DECLINE_OK backend=%s handled-actual-lookup=1 declined-actual-services=3 same-current-local8=true helper-original-body-credit=0", backend)
		})
		t.Run(backend+"/reentry", func(t *testing.T) {
			outerObjects, innerObjects := waExtraForeignObjects410390(t), waExtraForeignObjects410390(t)
			outer, inner := &wallAttachNativeObserver410390{}, &wallAttachNativeObserver410390{}
			inner.Service = func(site uint32, args []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool) {
				return waExtraForeignReply410390(t, innerObjects, site, args), true
			}
			nested := 0
			outer.Service = func(site uint32, args []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool) {
				if site == 0x4103a0 {
					nested++
					popInner := pushWallAttachNativeObserver410390(inner)
					got := func() unsafe.Pointer {
						defer popInner()
						return waExtraCall410390(t, backend, unsafe.Pointer(innerObjects.Drawable), 7, 9)
					}()
					if got != innerObjects.Polygon || wallAttachNativeObserverCurrent410390() != outer {
						t.Fatal("nested actual entry returns full identity and restores outer observer")
					}
				}
				return waExtraForeignReply410390(t, outerObjects, site, args), true
			}
			popOuter := pushWallAttachNativeObserver410390(outer)
			defer popOuter()
			got := waExtraCall410390(t, backend, unsafe.Pointer(outerObjects.Drawable), -1, 0x7fffffff)
			if got != outerObjects.Polygon || nested != 1 || wallAttachNativeObserverCurrent410390() != outer {
				t.Fatal("outer reentrant entry and exact observer identity")
			}
			t.Logf("WALL_ATTACH_REENTRY_OK backend=%s actual-entry-calls=2 full-native-return-identities=true serial-LIFO=true helper-original-body-credit=0", backend)
		})
	}
	observer := &wallAttachNativeObserver410390{}
	popOuter := pushWallAttachNativeObserver410390(observer)
	popInner := pushWallAttachNativeObserver410390(observer)
	rejected := false
	func() {
		defer func() { rejected = recover() != nil }()
		popOuter()
	}()
	if !rejected || wallAttachNativeObserverCurrent410390() != observer {
		t.Fatal("distinct push frames reject out-of-order same-observer pop")
	}
	popInner()
	popOuter()
	if wallAttachNativeObserverCurrent410390() != nil {
		t.Fatal("observer LIFO restores nil")
	}
	t.Log("WALL_ATTACH_OBSERVER_LIFO_OK same-observer-distinct-push-frames=true out-of-order-rejected=true")
}

func TestWallAttachPinnedGo410390(t *testing.T) {
	waExtraLayout410390(t)
	for _, backend := range []string{"Go", "C"} {
		for _, owner := range []string{"Go-wall", "foreign-wall"} {
			t.Run(backend+"/"+owner, func(t *testing.T) {
				drawable := new(client.Drawable)
				drawable.PosVec = image.Pt(int(int64(0x1234567880000001)), int(int64(0x123456787fffffff)))
				drawable.Field_0, drawable.Field_1 = 0x10203040, 0xfedcba98
				polygon := new([160]byte)
				for i := range polygon {
					polygon[i] = byte(i*13 + 7)
				}
				polygon[130] = 0xe3
				wall := new(server.Wall)
				freeWall := func() {}
				if owner == "foreign-wall" {
					wall, freeWall = alloc.NewHigh(server.Wall{})
					*wall = server.Wall{}
				}
				wall.Flags4, wall.Field8 = 0x80, 0xabcd
				var pin runtime.Pinner
				pin.Pin(drawable)
				pin.Pin(polygon)
				pinned := 2
				if owner == "Go-wall" {
					pin.Pin(wall)
					pinned++
				}
				t.Cleanup(func() {
					// Clear the actual C-retained Go reference before unpinning.
					wall.ClientData = nil
					pin.Unpin()
					freeWall()
					runtime.KeepAlive(drawable)
					runtime.KeepAlive(polygon)
					runtime.KeepAlive(wall)
				})
				for _, p := range []unsafe.Pointer{unsafe.Pointer(drawable), unsafe.Pointer(wall), unsafe.Pointer(polygon)} {
					waExtraWitness410390(t, p)
				}
				beforeDrawable := waExtraSnapshot410390(unsafe.Pointer(drawable), unsafe.Sizeof(*drawable))
				beforePolygon := waExtraSnapshot410390(unsafe.Pointer(polygon), unsafe.Sizeof(*polygon))
				beforeWall := waExtraSnapshot410390(unsafe.Pointer(wall), unsafe.Sizeof(*wall))
				gcInCall, gcPost, xDone, query := 0, 0, 0, 0
				gc := func() {
					for i := 0; i < 3; i++ {
						runtime.GC()
						gcInCall++
					}
				}
				observer := &wallAttachNativeObserver410390{}
				observer.Operand = func(e wallAttachNativeOperand410390) {
					if e.Site == 0x4103c4 && e.Complete {
						xDone++
						if e.Width != 8 || e.ValueBits != 0x1234567880000001 {
							t.Fatal("actual native intptr coordinate retains high32 before narrowing")
						}
						gc()
					}
				}
				observer.Service = func(site uint32, args []wallAttachNativeArg410390) (wallAttachNativeReply410390, bool) {
					switch site {
					case 0x4103a0:
						return wallAttachNativeReply410390{Pointer: unsafe.Pointer(wall)}, true
					case 0x4103e3:
						query++
						gc()
						if len(args) != 2 || !args[0].IsPointer || args[0].Pointer == nil {
							t.Fatal("actual pinned local8 query argument")
						}
						bits := unsafe.Slice((*byte)(args[0].Pointer), 8)
						if binary.LittleEndian.Uint32(bits[:4]) != 0x80000001 || binary.LittleEndian.Uint32(bits[4:]) != 0x7fffffff || wall.ClientData != unsafe.Pointer(drawable) || byte(wall.Flags4) != 0x90 {
							t.Fatal("nativeintptr narrowing and early full-pointer store survive GC")
						}
						return wallAttachNativeReply410390{Pointer: unsafe.Pointer(polygon)}, true
					default:
						t.Fatalf("unexpected pinned-control service %#x", site)
						return wallAttachNativeReply410390{}, true
					}
				}
				pop := pushWallAttachNativeObserver410390(observer)
				defer pop()
				got := waExtraCall410390(t, backend, unsafe.Pointer(drawable), -1, 0x7fffffff)
				for i := 0; i < 3; i++ {
					runtime.GC()
					gcPost++
				}
				if got != unsafe.Pointer(polygon) || wall.ClientData != unsafe.Pointer(drawable) || wall.Field8 != 0xabe3 || xDone != 1 || query != 1 || gcInCall != 6 || gcPost != 3 {
					t.Fatal("actual pinned full stored/returned identities through in-call and retained-pointer GC")
				}
				waExtraUntouchedWall410390(t, beforeWall, wall)
				if !bytes.Equal(beforeDrawable, waExtraSnapshot410390(unsafe.Pointer(drawable), unsafe.Sizeof(*drawable))) || !bytes.Equal(beforePolygon, waExtraSnapshot410390(unsafe.Pointer(polygon), unsafe.Sizeof(*polygon))) {
					t.Fatal("whole pinned drawable/polygon canaries unchanged")
				}
				t.Logf("WALL_ATTACH_PINNED_GO_OK backend=%s owner=%s individually-pinned-Go-owners=%d actual-C-pointer-witnesses=3 native-intptr-read-width=8 narrowed-local8=true full-stored-and-returned-pointers=true forced-in-call-GC=6 forced-post-call-GC=3 whole-owner-canaries=true clear-C-retained-before-Unpin=true helper-original-body-credit=0 physical-stack-FPU-credit=0", backend, owner, pinned)
			})
		}
	}
}
