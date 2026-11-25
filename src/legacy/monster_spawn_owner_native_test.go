package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_spawn_owner_lifetime_v1.json.gz
var monsterSpawnOwnerNativeFixture []byte

type monsterSpawnOwnerNativeNode struct {
	Generator int
	Link      int
}

type monsterSpawnOwnerNativeLink struct {
	Object int
	Older  int
	Newer  int
}

type monsterSpawnOwnerNativeEvent struct {
	Operation []json.RawMessage
	Result    int
	Counts    [2]byte
	Nodes     [4]monsterSpawnOwnerNativeNode
	Links     []monsterSpawnOwnerNativeLink
	Head      int
}

type monsterSpawnOwnerNativeCase struct {
	Input struct {
		Name          string
		Program       string
		Seed          byte
		InitialCounts [2]byte `json:"initial_counts"`
	}
	Expected []monsterSpawnOwnerNativeEvent `json:"target_expected"`
}

func TestMonsterSpawnOwnerNativeLayout(t *testing.T) {
	want := [4]uintptr{unsafe.Sizeof(server.MonsterSpawnLink{}), unsafe.Alignof(server.MonsterSpawnLink{}), unsafe.Offsetof(server.MonsterSpawnLink{}.Older), unsafe.Offsetof(server.MonsterSpawnLink{}.Newer)}
	if got := cfixture.MonsterSpawnOwnerNativeCLayout(); got != want || want[0] != 3*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("native C/Go link layout", got, want)
	}
}

func monsterSpawnOwnerNativeCases(t *testing.T) []monsterSpawnOwnerNativeCase {
	t.Helper()
	sum := sha256.Sum256(monsterSpawnOwnerNativeFixture)
	if hex.EncodeToString(sum[:]) != "7b03f41957a40c62151e6cbea82ba876528a58d574a997ee7c064466402911b6" {
		t.Fatal("preport owner contract changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(monsterSpawnOwnerNativeFixture))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var cases []monsterSpawnOwnerNativeCase
	if err := json.Unmarshal(raw, &cases); err != nil || len(cases) != 24 {
		t.Fatal("sealed model", len(cases), err)
	}
	return cases
}

func monsterSpawnOwnerNativeRun(t *testing.T, c monsterSpawnOwnerNativeCase, place int) {
	t.Helper()
	owner, free := alloc.MakeHigh([]byte{}, 65536)
	defer free()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&owner[0])) <= math.MaxUint32 {
		t.Fatal("foreign owner must exceed4GiB")
	}
	if unsafe.Sizeof(server.Object{}) > 0x800 || unsafe.Sizeof(server.MonsterGenUpdateData{}) > 0x800 || unsafe.Sizeof(server.MonsterUpdateData{}) > 0x1800 {
		t.Fatal("declared foreign owner ranges overlap")
	}
	for i := range owner {
		owner[i] = byte(int(c.Input.Seed) + i*17 + 3*(i>>8))
	}
	ptr := func(offset int) unsafe.Pointer { return unsafe.Pointer(&owner[place+offset]) }
	zero := func(pointer unsafe.Pointer, size uintptr) { clear(unsafe.Slice((*byte)(pointer), size)) }
	var nodes [4]*server.Object
	var data [4]*server.MonsterUpdateData
	for i := range nodes {
		nodes[i] = (*server.Object)(ptr(0x1000 + 0x800*i))
		data[i] = (*server.MonsterUpdateData)(ptr(0x6000 + 0x2000*i))
		zero(unsafe.Pointer(nodes[i]), unsafe.Sizeof(*nodes[i]))
		zero(unsafe.Pointer(data[i]), unsafe.Sizeof(*data[i]))
		nodes[i].ObjClass = object.ClassMonster
		nodes[i].UpdateData = unsafe.Pointer(data[i])
	}
	var generators [2]*server.Object
	var generatorData [2]*server.MonsterGenUpdateData
	for i := range generators {
		generators[i] = (*server.Object)(ptr(0x3000 + 0x800*i))
		generatorData[i] = (*server.MonsterGenUpdateData)(ptr(0x4000 + 0x800*i))
		zero(unsafe.Pointer(generators[i]), unsafe.Sizeof(*generators[i]))
		zero(unsafe.Pointer(generatorData[i]), unsafe.Sizeof(*generatorData[i]))
		generators[i].ObjClass = object.ClassMonsterGenerator
		generators[i].UpdateData = unsafe.Pointer(generatorData[i])
		generatorData[i].ActiveCount = c.Input.InitialCounts[i]
	}
	before := append([]byte(nil), owner...)
	s := server.NewHigh(nil, nil, strman.New())
	defer s.Close()
	links := map[int]*server.MonsterSpawnLink{}
	word := unsafe.Sizeof(uintptr(0))
	putWord := func(image []byte, offset int, value uintptr) {
		for i := uintptr(0); i < word; i++ {
			image[offset+int(i)] = byte(value >> (8 * i))
		}
	}
	linkPointer := func(role int) *server.MonsterSpawnLink {
		if role == 0 {
			return nil
		}
		v := links[role]
		if v == nil {
			t.Fatal("model refers to an unknown allocated link", c.Input.Name, role)
		}
		return v
	}
	for index, expected := range c.Expected {
		var operation string
		if err := json.Unmarshal(expected.Operation[0], &operation); err != nil {
			t.Fatal(err)
		}
		argument := func(i int) int {
			var value int
			if err := json.Unmarshal(expected.Operation[i], &value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		result := -1
		switch operation {
		case "initial":
		case "init":
			result = 0
			if s.MonsterSpawnInit50D780() {
				result = 1
			}
		case "register":
			var node, generator *server.Object
			if n := argument(1); n >= 0 {
				node = nodes[n]
			}
			if g := argument(2); g >= 0 {
				generator = generators[g]
			}
			result = 0
			if s.MonsterSpawnRegister50E030(generator, node, server.MonsterSpawnRegisterRuntime50E030{}) {
				result = 1
			}
			runtime.GC()
		case "cleanup":
			var node *server.Object
			if n := argument(1); n >= 0 {
				node = nodes[n]
			}
			s.MonsterSpawnCleanup50E140(node)
		case "reset":
			s.MonsterSpawnReset50D7E0()
		case "free":
			s.MonsterSpawnFree50D820()
		case "close":
			s.Close()
		default:
			t.Fatal("undeclared operation", operation)
		}
		if result != expected.Result {
			t.Fatal("owner return", c.Input.Name, place, index, operation, result, expected.Result)
		}
		for i, state := range expected.Nodes {
			if state.Link == 0 || links[state.Link] != nil {
				continue
			}
			actual := data[i].Field549
			if actual == nil || state.Link != len(links)+1 {
				t.Fatal("new managed link", c.Input.Name, place, index, i, state.Link, actual)
			}
			for _, previous := range links {
				if previous == actual {
					t.Fatal("new link aliases a retained allocation", c.Input.Name, index, state.Link)
				}
			}
			links[state.Link] = actual
		}
		image := append([]byte(nil), before...)
		for i, count := range expected.Counts {
			image[place+0x4000+0x800*i+int(unsafe.Offsetof(server.MonsterGenUpdateData{}.ActiveCount))] = count
			if generatorData[i].ActiveCount != count {
				t.Fatal("native wrapping counter", c.Input.Name, place, index, i, generatorData[i].ActiveCount, count)
			}
		}
		for i, state := range expected.Nodes {
			var generator *server.Object
			if state.Generator != 0 {
				generator = generators[state.Generator-1]
			}
			link := linkPointer(state.Link)
			if data[i].Field548 != generator || data[i].Field549 != link {
				t.Fatal("native backreferences", c.Input.Name, place, index, i)
			}
			for _, field := range []struct {
				offset uintptr
				value  uintptr
			}{
				{unsafe.Offsetof(server.MonsterUpdateData{}.Field548), uintptr(unsafe.Pointer(generator))},
				{unsafe.Offsetof(server.MonsterUpdateData{}.Field549), uintptr(unsafe.Pointer(link))},
			} {
				if got := cfixture.MonsterSpawnOwnerNativeRead(unsafe.Pointer(data[i]), field.offset); got != field.value {
					t.Fatal("C native backreference read", c.Input.Name, place, index, i, got, field.value)
				}
				putWord(image, place+0x6000+0x2000*i+int(field.offset), field.value)
			}
		}
		if len(links) != len(expected.Links) {
			t.Fatal("allocation history", c.Input.Name, index, len(links), len(expected.Links))
		}
		for i, state := range expected.Links {
			link := linkPointer(i + 1)
			var objectPointer *server.Object
			if state.Object != 0 {
				objectPointer = nodes[state.Object-1]
			}
			older, newer := linkPointer(state.Older), linkPointer(state.Newer)
			if link.Object != objectPointer || link.Older != older || link.Newer != newer {
				t.Fatal("owned link graph", c.Input.Name, place, index, i)
			}
			want := [3]uintptr{uintptr(unsafe.Pointer(objectPointer)), uintptr(unsafe.Pointer(older)), uintptr(unsafe.Pointer(newer))}
			var actual [3]uintptr
			for field := range actual {
				actual[field] = cfixture.MonsterSpawnOwnerNativeRead(unsafe.Pointer(link), uintptr(field)*word)
			}
			if actual != want {
				t.Fatal("C pinned link graph", c.Input.Name, place, index, i, actual, want)
			}
			if state.Object != 0 && state.Newer == 0 && i+1 != expected.Head {
				t.Fatal("ordered native head", c.Input.Name, place, index, i)
			}
		}
		if !bytes.Equal(owner, image) {
			t.Fatal("complete64KiB foreign owner changed outside declared effects", c.Input.Name, place, index, operation)
		}
		runtime.KeepAlive(links)
	}
}

func TestMonsterSpawnOwnerNativeContract(t *testing.T) {
	cases := monsterSpawnOwnerNativeCases(t)
	seen, events := 0, 0
	for _, program := range []string{"ordered-cleanup", "duplicate-and-nil", "reset-and-reuse", "reinit-live", "free-and-reinit", "close-live"} {
		t.Run(program, func(t *testing.T) {
			for _, place := range []int{0, 256, 1024} {
				t.Run(fmt.Sprintf("place%d", place), func(t *testing.T) {
					count := 0
					for _, c := range cases {
						if c.Input.Program != program {
							continue
						}
						monsterSpawnOwnerNativeRun(t, c, place)
						count++
						seen++
						events += len(c.Expected)
					}
					if count != 4 {
						t.Fatal("sealed input coverage", count)
					}
				})
			}
		})
	}
	if seen != 72 || events != 804 {
		t.Fatal("complete owner coverage", seen, events)
	}
}
