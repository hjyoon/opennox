package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_spawn_raw_links_original_v1.json.gz
var monsterSpawnRawLinksOriginal []byte

type monsterSpawnNativeServer struct {
	Server
	native  *server.Server
	deleted []*server.Object
}

func (s *monsterSpawnNativeServer) S() *server.Server { return s.native }
func (s *monsterSpawnNativeServer) DelayedDelete(obj *server.Object) {
	s.deleted = append(s.deleted, obj)
	s.native.MonsterSpawnCleanup50E140(obj)
}

type monsterSpawnNativeOwner struct {
	t         *testing.T
	bytes     []byte
	place     int
	outer     *monsterSpawnNativeServer
	generator *server.Object
	genData   *server.MonsterGenUpdateData
	objects   [4]*server.Object
	updates   [4]*server.MonsterUpdateData
	links     [5]*server.MonsterSpawnLink
}

func newMonsterSpawnNativeOwner(t *testing.T, place int) *monsterSpawnNativeOwner {
	t.Helper()
	buffer, free := alloc.MakeHigh([]byte{}, 65536+64)
	s := server.NewHigh(nil, nil, strman.New())
	outer := &monsterSpawnNativeServer{native: s}
	old := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { s.Close(); GetServer = old; free() })
	o := &monsterSpawnNativeOwner{t: t, bytes: buffer, place: place, outer: outer}
	pointer := func(offset int, size uintptr) unsafe.Pointer {
		if 32+offset+place+int(size) > len(buffer)-32 {
			t.Fatal("native fixture extent")
		}
		return unsafe.Pointer(&buffer[32+offset+place])
	}
	o.generator = (*server.Object)(pointer(0x1000, unsafe.Sizeof(server.Object{})))
	o.genData = (*server.MonsterGenUpdateData)(pointer(0x1800, unsafe.Sizeof(server.MonsterGenUpdateData{})))
	for i := range o.objects {
		o.objects[i] = (*server.Object)(pointer(0x2000+i*0x800, unsafe.Sizeof(server.Object{})))
		o.updates[i] = (*server.MonsterUpdateData)(pointer(0x6000+i*0xe00, unsafe.Sizeof(server.MonsterUpdateData{})))
		o.links[i] = (*server.MonsterSpawnLink)(pointer(0xe000+i*64, unsafe.Sizeof(server.MonsterSpawnLink{})))
	}
	o.links[4] = (*server.MonsterSpawnLink)(pointer(0xe200, unsafe.Sizeof(server.MonsterSpawnLink{})))
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o.generator)) <= math.MaxUint32 {
		t.Fatal("actual C owner must be above4GiB")
	}
	rt := s.TypePredicateRuntime534A10()
	rt.Write(1, 11)
	rt.Write(2, 12)
	return o
}

func (o *monsterSpawnNativeOwner) seed(seed byte, base uint8) {
	for i := range o.bytes {
		o.bytes[i] = 0xd7
	}
	for i := 0; i < 65536; i++ {
		o.bytes[32+i] = byte(int(seed) + 17*i + 3*(i>>8))
	}
	*o.genData = server.MonsterGenUpdateData{ActiveCount: base}
	*o.generator = server.Object{ObjClass: object.ClassMonsterGenerator, UpdateData: unsafe.Pointer(o.genData)}
	for i, obj := range o.objects {
		*o.updates[i] = server.MonsterUpdateData{}
		*obj = server.Object{ObjClass: object.ClassMonster, TypeInd: 99, UpdateData: unsafe.Pointer(o.updates[i])}
		*o.links[i] = server.MonsterSpawnLink{Object: obj}
		if monsterSpawnRegistryNativeObjectData(unsafe.Pointer(obj)) != uintptr(unsafe.Pointer(o.updates[i])) {
			o.t.Fatal("C Object update pointer differs")
		}
	}
	*o.links[4] = server.MonsterSpawnLink{}
	o.outer.deleted = nil
}

func (o *monsterSpawnNativeOwner) expectedBytes(expected []byte, field unsafe.Pointer, value []byte) {
	offset := uintptr(field) - uintptr(unsafe.Pointer(&o.bytes[0]))
	if offset+uintptr(len(value)) > uintptr(len(expected)) {
		o.t.Fatal("expected owner extent")
	}
	copy(expected[int(offset):], value)
}
func (o *monsterSpawnNativeOwner) expectedPointer(expected []byte, field unsafe.Pointer, value uintptr) {
	var word [8]byte
	if unsafe.Sizeof(value) == 4 {
		binary.LittleEndian.PutUint32(word[:], uint32(value))
	} else {
		binary.LittleEndian.PutUint64(word[:], uint64(value))
	}
	o.expectedBytes(expected, field, word[:unsafe.Sizeof(value)])
}
func (o *monsterSpawnNativeOwner) checkOwner(expected []byte) {
	o.t.Helper()
	if !bytes.Equal(o.bytes, expected) {
		for i, value := range expected {
			if o.bytes[i] != value {
				o.t.Fatalf("whole native owner differs at%#x:got%#x want%#x", i, o.bytes[i], value)
			}
		}
	}
}

func TestMonsterSpawnRegistryNativeLayout(t *testing.T) {
	var obj server.Object
	var link server.MonsterSpawnLink
	var screen server.MonsterSpawnScreenRecord50D960
	prefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	if unsafe.Sizeof(obj) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("Go-only Object handles changed")
	}
	want := [12]uintptr{unsafe.Sizeof(uintptr(0)), prefix, unsafe.Offsetof(obj.UpdateData), unsafe.Sizeof(link), unsafe.Offsetof(link.Older), unsafe.Offsetof(link.Newer), unsafe.Sizeof(screen), unsafe.Offsetof(screen.Average), unsafe.Offsetof(screen.Mask), unsafe.Offsetof(screen.Distances), unsafe.Offsetof(screen.Older), unsafe.Offsetof(screen.Newer)}
	if got := monsterSpawnRegistryNativeLayout(); got != want {
		t.Fatal("all16 typed public entries and native records", got, want)
	}
}

func TestMonsterSpawnRegistryNativeOriginalRawLinks(t *testing.T) {
	var cases []struct {
		Name     string
		Seed     byte
		Order    []int `json:"initial_order"`
		Entry    string
		Argument int `json:"argument_role"`
		Return   int `json:"return_role"`
		Head     int `json:"head_role"`
		Links    [4]struct {
			Object int `json:"object_role"`
			Older  int `json:"older_role"`
			Newer  int `json:"newer_role"`
		}
	}
	npcVoiceOriginalDecode(t, monsterSpawnRawLinksOriginal, "8c7799474ba598cbb6a574034b2e81a4a5b947d7f151bbcecb2c4d5254b7544e", "bbbca8e6171e9805e9ea7ad6997ffc5fc59d2bf955d1bb1f56d5279d37a0430f", 129, &cases)
	for route := 0; route < 2; route++ {
		for _, place := range []int{0, 512, 4096} {
			t.Run(fmt.Sprintf("route%d/place%d", route, place), func(t *testing.T) {
				o := newMonsterSpawnNativeOwner(t, place)
				linkAt := func(role int) *server.MonsterSpawnLink {
					if role == 0 {
						return nil
					}
					return o.links[role-1]
				}
				for _, c := range cases {
					t.Run(c.Name, func(t *testing.T) {
						o.t = t
						o.seed(c.Seed, 0)
						if monsterSpawnRegistryNativeCall(route, 0, nil, nil) != 1 {
							t.Fatal("native init")
						}
						expected := bytes.Clone(o.bytes)
						owned := [5]bool{}
						for i := len(c.Order) - 1; i >= 0; i-- {
							role := c.Order[i]
							owned[role] = true
							if monsterSpawnRegistryNativeCall(route, 11, unsafe.Pointer(o.links[role]), nil) != uintptr(unsafe.Pointer(o.links[role])) {
								t.Fatal("seed prepend truncated")
							}
						}
						entry := 13
						if c.Entry == "prepend" {
							entry = 11
							owned[c.Argument-1] = true
						}
						result := monsterSpawnRegistryNativeCall(route, entry, unsafe.Pointer(linkAt(c.Argument)), nil)
						if result != uintptr(unsafe.Pointer(linkAt(c.Return))) {
							t.Fatal("original full pointer return", result, c.Return)
						}
						// Derive expected complete owner bytes from the sealed original
						// topology, never from the implementation's observed node fields.
						for i, want := range c.Links {
							o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Object), uintptr(unsafe.Pointer(o.objects[want.Object-1])))
							o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Older), uintptr(unsafe.Pointer(linkAt(want.Older))))
							o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Newer), uintptr(unsafe.Pointer(linkAt(want.Newer))))
							if monsterSpawnRegistryNativeRead(unsafe.Pointer(o.links[i]), 0) != uintptr(unsafe.Pointer(o.objects[want.Object-1])) || monsterSpawnRegistryNativeRead(unsafe.Pointer(o.links[i]), unsafe.Offsetof(o.links[i].Older)) != uintptr(unsafe.Pointer(linkAt(want.Older))) || monsterSpawnRegistryNativeRead(unsafe.Pointer(o.links[i]), unsafe.Offsetof(o.links[i].Newer)) != uintptr(unsafe.Pointer(linkAt(want.Newer))) {
								t.Fatal("actual C topology differs from original", i)
							}
						}
						o.checkOwner(expected)
						// Prepending a separate probe exposes the otherwise private head.
						probe := o.links[4]
						owned[4] = true
						if monsterSpawnRegistryNativeCall(route, 11, unsafe.Pointer(probe), nil) != uintptr(unsafe.Pointer(probe)) || monsterSpawnRegistryNativeRead(unsafe.Pointer(probe), unsafe.Offsetof(probe.Older)) != uintptr(unsafe.Pointer(linkAt(c.Head))) {
							t.Fatal("native head differs from original")
						}
						monsterSpawnRegistryNativeCall(route, 1, nil, nil)
						for i, yes := range owned {
							if yes {
								o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Object), 0)
								o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Older), 0)
								o.expectedPointer(expected, unsafe.Pointer(&o.links[i].Newer), 0)
							}
						}
						o.checkOwner(expected)
					})
				}
			})
		}
	}
}

func TestMonsterSpawnRegistryNativeEntriesAndGC(t *testing.T) {
	for route := 0; route < 2; route++ {
		for _, place := range []int{0, 512, 4096} {
			t.Run(fmt.Sprintf("route%d/place%d", route, place), func(t *testing.T) {
				o := newMonsterSpawnNativeOwner(t, place)
				s := o.outer.native
				for mode, cleanup := range []int{12, 14, 15} {
					base := []uint8{0, 253, 255}[mode]
					o.seed(byte(11+mode*23), base)
					if monsterSpawnRegistryNativeCall(route, 0, nil, nil) != 1 {
						t.Fatal("C init failed")
					}
					expected := bytes.Clone(o.bytes)
					var links [4]*server.MonsterSpawnLink
					for i, obj := range o.objects {
						if monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(obj)) != 1 {
							t.Fatal("C register failed")
						}
						links[i] = o.updates[i].Field549
						if links[i] == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(links[i])) <= math.MaxUint32 {
							t.Fatal("native link address truncated")
						}
						o.expectedPointer(expected, unsafe.Pointer(&o.updates[i].Field548), uintptr(unsafe.Pointer(o.generator)))
						o.expectedPointer(expected, unsafe.Pointer(&o.updates[i].Field549), uintptr(unsafe.Pointer(links[i])))
						o.expectedBytes(expected, unsafe.Pointer(&o.genData.ActiveCount), []byte{base + uint8(i) + 1})
						o.checkOwner(expected)
						if monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(obj)) != 1 {
							t.Fatal("duplicate C register")
						}
						o.checkOwner(expected)
					}
					runtime.GC()
					for i, link := range links {
						var older, newer *server.MonsterSpawnLink
						if i > 0 {
							older = links[i-1]
						}
						if i+1 < len(links) {
							newer = links[i+1]
						}
						if monsterSpawnRegistryNativeRead(unsafe.Pointer(link), 0) != uintptr(unsafe.Pointer(o.objects[i])) || monsterSpawnRegistryNativeRead(unsafe.Pointer(link), unsafe.Offsetof(link.Older)) != uintptr(unsafe.Pointer(older)) || monsterSpawnRegistryNativeRead(unsafe.Pointer(link), unsafe.Offsetof(link.Newer)) != uintptr(unsafe.Pointer(newer)) {
							t.Fatal("actual C link reads after GC", i)
						}
					}
					for n, i := range []int{1, 3, 0, 2} {
						monsterSpawnRegistryNativeCall(route, cleanup, unsafe.Pointer(o.objects[i]), nil)
						o.expectedPointer(expected, unsafe.Pointer(&o.updates[i].Field548), 0)
						o.expectedPointer(expected, unsafe.Pointer(&o.updates[i].Field549), 0)
						o.expectedBytes(expected, unsafe.Pointer(&o.genData.ActiveCount), []byte{base + 4 - uint8(n) - 1})
						o.checkOwner(expected)
						if links[i].Object != nil || links[i].Older != nil || links[i].Newer != nil {
							t.Fatal("released node edges remain")
						}
					}
					// Exercise all scalar, no-player and rejected-class public entries.
					if monsterSpawnRegistryNativeCall(route, 5, nil, nil) != 0 || monsterSpawnRegistryNativeCall(route, 6, nil, nil) != 0 || monsterSpawnRegistryNativeCall(route, 7, unsafe.Pointer(o.objects[0]), nil) != 0 {
						t.Fatal("empty scratch entries")
					}
					point := (*types.Pointf)(unsafe.Pointer(&o.bytes[32+0xd800+place]))
					*point = types.Ptf(1, 2)
					expected = bytes.Clone(o.bytes)
					if monsterSpawnRegistryNativeCall(route, 8, unsafe.Pointer(o.generator), unsafe.Pointer(point)) != 1 {
						t.Fatal("no-player placement")
					}
					o.objects[0].ObjClass = 0
					expected = bytes.Clone(o.bytes)
					monsterSpawnRegistryNativeCall(route, 9, unsafe.Pointer(o.objects[0]), unsafe.Pointer(o.objects[1]))
					o.checkOwner(expected)
					s.SetTickRate(30)
					s.SetFrame(31)
					if monsterSpawnRegistryNativeCall(route, 3, nil, nil) != 2 {
						t.Fatal("original cadence scalar return")
					}
					monsterSpawnRegistryNativeCall(route, 2, nil, nil)
					o.checkOwner(expected)
				}
				// Actual C distance cleanup with synchronous owner removal must use
				// the cached next link, including the tail after the first GC.
				o.seed(73, 255)
				monsterSpawnRegistryNativeCall(route, 0, nil, nil)
				for _, obj := range o.objects {
					if monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(obj)) != 1 {
						t.Fatal("distance seed")
					}
				}
				runtime.GC()
				monsterSpawnRegistryNativeCall(route, 4, nil, nil)
				if len(o.outer.deleted) != 4 || o.genData.ActiveCount != 255 {
					t.Fatal("actual C cached-next cleanup", len(o.outer.deleted), o.genData.ActiveCount)
				}
				for i, obj := range o.outer.deleted {
					if obj != o.objects[3-i] {
						t.Fatal("actual C deletion order")
					}
				}
				monsterSpawnRegistryNativeCall(route, 2, nil, nil)
			})
		}
	}
}

func TestMonsterSpawnRegistryNativeLiveScratchQuery(t *testing.T) {
	for route := 0; route < 2; route++ {
		for _, place := range []int{0, 512, 4096} {
			t.Run(fmt.Sprintf("route%d/place%d", route, place), func(t *testing.T) {
				o := newMonsterSpawnNativeOwner(t, place)
				s := o.outer.native
				o.seed(91, 0)
				monsterSpawnRegistryNativeCall(route, 0, nil, nil)
				if monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(o.objects[0])) != 1 {
					t.Fatal("scratch seed")
				}
				player := (*server.Object)(unsafe.Pointer(&o.bytes[32+0x4000+place]))
				update := (*server.PlayerUpdateData)(unsafe.Pointer(&o.bytes[32+0x5000+place]))
				info := (*server.Player)(unsafe.Pointer(&o.bytes[32+0xa000+place]))
				if uintptr(unsafe.Pointer(info))+unsafe.Sizeof(*info) > uintptr(unsafe.Pointer(&o.bytes[len(o.bytes)-32])) {
					t.Fatal("player owner extent")
				}
				*info = server.Player{Field4792: 1, Field10: 300, Field12: 200}
				*update = server.PlayerUpdateData{Player: info}
				*player = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
				queries := 0
				rt := server.MonsterSpawnDeleteRuntime50E210{
					FirstPlayerUnit: func() *server.Object { return player }, NextPlayerUnit: func(*server.Object) *server.Object { return nil }, MaxOnscreen: func() float64 { return 1 },
					PointInRect: func(types.Pointf, types.Rectf) bool { return true }, Trace: func(types.Pointf, types.Pointf, server.MapTraceFlags) int { return 1 },
					Distance: func(obj, unit *server.Object) float64 {
						record := s.MonsterSpawnScratchFind50DE60(obj)
						if record == nil || unit != player {
							t.Fatal("live native scratch owner")
						}
						if value := monsterSpawnRegistryNativeCall(route, 7, unsafe.Pointer(obj), nil); value != uintptr(unsafe.Pointer(record)) {
							t.Fatal("public C scratch pointer truncated", value)
						}
						runtime.GC()
						if monsterSpawnRegistryNativeRead(unsafe.Pointer(record), 0) != uintptr(unsafe.Pointer(obj)) {
							t.Fatal("C scratch record lost across GC")
						}
						queries++
						return 10
					},
				}
				if s.MonsterSpawnCullVisible50D960(rt) != 0 || queries != 1 || monsterSpawnRegistryNativeCall(route, 7, unsafe.Pointer(o.objects[0]), nil) != 0 {
					t.Fatal("scratch query lifetime")
				}
				monsterSpawnRegistryNativeCall(route, 2, nil, nil)
			})
		}
	}
}

func TestMonsterSpawnRegistryNativeDetachedResetAndZombie(t *testing.T) {
	for route := 0; route < 2; route++ {
		for _, place := range []int{0, 512, 4096} {
			t.Run(fmt.Sprintf("route%d/place%d", route, place), func(t *testing.T) {
				o := newMonsterSpawnNativeOwner(t, place)
				for _, reset := range []int{1, 2, 0} {
					o.seed(53, 253)
					monsterSpawnRegistryNativeCall(route, 0, nil, nil)
					expected := bytes.Clone(o.bytes)
					var links [4]*server.MonsterSpawnLink
					for i, obj := range o.objects {
						if monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(obj)) != 1 {
							t.Fatal("detached seed")
						}
						links[i] = o.updates[i].Field549
					}
					if result := monsterSpawnRegistryNativeCall(route, 13, unsafe.Pointer(links[1]), nil); result != uintptr(unsafe.Pointer(links[0])) {
						t.Fatal("raw unlink pointer result")
					}
					if links[1].Older != links[0] || links[1].Newer != links[2] {
						t.Fatal("raw unlink lost original detached edges")
					}
					runtime.GC()
					result := monsterSpawnRegistryNativeCall(route, reset, nil, nil)
					if reset == 0 && result != 1 {
						t.Fatal("native reinit")
					}
					o.checkOwner(expected)
					for _, link := range links {
						if monsterSpawnRegistryNativeRead(unsafe.Pointer(link), 0) != 0 || monsterSpawnRegistryNativeRead(unsafe.Pointer(link), unsafe.Offsetof(link.Older)) != 0 || monsterSpawnRegistryNativeRead(unsafe.Pointer(link), unsafe.Offsetof(link.Newer)) != 0 {
							t.Fatal("detached pinned owner survived reset/free/reinit")
						}
					}
				}
				for _, zombieType := range []uint16{11, 12} {
					o.seed(69, 255)
					monsterSpawnRegistryNativeCall(route, 0, nil, nil)
					o.objects[0].TypeInd = zombieType
					expected := bytes.Clone(o.bytes)
					monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(o.objects[0]))
					registered := bytes.Clone(o.bytes)
					monsterSpawnRegistryNativeCall(route, 14, unsafe.Pointer(o.objects[0]), nil)
					o.checkOwner(registered)
					monsterSpawnRegistryNativeCall(route, 12, unsafe.Pointer(o.objects[0]), nil)
					monsterSpawnRegistryNativeCall(route, 12, nil, nil)
					monsterSpawnRegistryNativeCall(route, 14, nil, nil)
					o.checkOwner(expected)
				}
				o.seed(83, 1)
				monsterSpawnRegistryNativeCall(route, 0, nil, nil)
				expected := bytes.Clone(o.bytes)
				for _, obj := range o.objects {
					monsterSpawnRegistryNativeCall(route, 10, unsafe.Pointer(o.generator), unsafe.Pointer(obj))
				}
				o.outer.native.SetTickRate(30)
				o.outer.native.SetFrame(150)
				if monsterSpawnRegistryNativeCall(route, 3, nil, nil) != 0 || len(o.outer.deleted) != 4 {
					t.Fatal("actual C combined cadence")
				}
				o.checkOwner(expected)
			})
		}
	}
}
