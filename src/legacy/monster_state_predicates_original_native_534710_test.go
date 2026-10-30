package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type statePredicateInput534710 struct {
	Class         uint32 `json:"cls"`
	Status        uint32 `json:"status"`
	Head          int8   `json:"head"`
	Action        uint32 `json:"action"`
	Prefix        uint32 `json:"prefix528"`
	Type          uint16 `json:"type_index"`
	Poison        uint8  `json:"poison"`
	HealthPresent bool   `json:"health_present"`
	Cur           uint16 `json:"health_cur"`
	Max           uint16 `json:"health_max"`
	Frame         uint32 `json:"frame"`
	FPS           uint32 `json:"fps"`
	MoveFrame     uint32 `json:"move_frame"`
	Cache         uint32 `json:"cache"`
	Lookup        uint16 `json:"lookup_index"`
	Repeat        int    `json:"repeat"`
}

func TestMonsterStatePredicatesOriginalActualC534710(t *testing.T) {
	path := os.Getenv("NOX_STATE_PREDICATE_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_state_predicates_classes_original_534710.json.gz"
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(blob)
	if hex.EncodeToString(sum[:]) != "4d493f816951c0eec91f37c41e1312611ae72d7dbedc128db211cb9aee37dbc8" {
		t.Fatal("unsealed original state reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var reference struct {
		Schema      string `json:"schema"`
		OriginalSHA string `json:"original_sha256"`
		Control     uint16 `json:"control_word"`
		Rows        []struct {
			Name       string                    `json:"name"`
			Entry      uint32                    `json:"entry"`
			Input      statePredicateInput534710 `json:"input"`
			Returns    []uint32                  `json:"returns"`
			CacheAfter uint32                    `json:"cache_after"`
			Lookups    []struct {
				Service     string  `json:"service"`
				Name        string  `json:"name"`
				Found       bool    `json:"found"`
				Type        *uint16 `json:"type_index"`
				CacheBefore uint32  `json:"cache_before"`
				Call        int     `json:"call"`
			} `json:"lookups"`
			Writes []struct {
				Field string `json:"field"`
				Value uint32 `json:"value"`
				Call  int    `json:"call"`
			} `json:"cache_writes"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&reference); err != nil {
		t.Fatal(err)
	}
	if reference.Schema != "original-monster-state-predicates-classes-534710-v1" || reference.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || reference.Control != 0xe3f || len(reference.Rows) != 16440 {
		t.Fatal("original state provenance")
	}
	srv, _, _, _ := monsterMainBlockLegacyFixture547210(t)
	owner := GetServer
	serverReads := 0
	GetServer = func() Server { serverReads++; return owner() }
	t.Cleanup(func() { GetServer = owner })
	oldCache := statePredicateOriginalSwapMimicCache534840(0)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldCache) })
	cachePtr := statePredicateOriginalMimicCachePointer534840()
	type unitGuard struct {
		Before [64]byte
		Value  server.Object
		After  [64]byte
	}
	type updateGuard struct {
		Before [64]byte
		Value  server.MonsterUpdateData
		After  [64]byte
	}
	type healthGuard struct {
		Before [64]byte
		Value  server.HealthData
		After  [64]byte
	}
	ug, freeU := alloc.New(unitGuard{})
	t.Cleanup(freeU)
	dg, freeD := alloc.New(updateGuard{})
	t.Cleanup(freeD)
	hg, freeH := alloc.New(healthGuard{})
	t.Cleanup(freeH)
	regions := []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(ug), unsafe.Sizeof(*ug)}, {unsafe.Pointer(dg), unsafe.Sizeof(*dg)}, {unsafe.Pointer(hg), unsafe.Sizeof(*hg)}}
	snapshot := func(ptr unsafe.Pointer, size uintptr) []byte {
		return append([]byte(nil), unsafe.Slice((*byte)(ptr), int(size))...)
	}
	for _, region := range regions {
		for i := range unsafe.Slice((*byte)(region.ptr), int(region.size)) {
			unsafe.Slice((*byte)(region.ptr), int(region.size))[i] = byte(i*31 + 19)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(region.ptr) <= math.MaxUint32 {
			t.Fatal("native C state records must exceed4GiB")
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(cachePtr) <= math.MaxUint32 {
		t.Fatal("native Mimic cache must exceed4GiB")
	}
	registry := -1
	setRegistry := func(index uint16) {
		if registry == int(index) {
			return
		}
		srv.FreeObjectTypes()
		for i := 1; i <= int(index); i++ {
			name := fmt.Sprintf("StatePredicateFixture%d", i)
			if i == int(index) {
				name = "Mimic"
			}
			if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
				t.Fatal(err)
			}
		}
		if got := srv.Types.IndByID("Mimic"); got != int(index) {
			t.Fatalf("real native type registry got=%d want=%d", got, index)
		}
		registry = int(index)
	}
	t.Cleanup(srv.FreeObjectTypes)
	controlOnly := os.Getenv("NOX_STATE_PREDICATE_REFERENCE_CONTROL_ONLY") == "true"
	entries := map[uint32]int{}
	totalCalls, lookupControls := 0, 0
	for _, row := range reference.Rows {
		t.Run(row.Name, func(t *testing.T) {
			c := row.Input
			if c.Repeat < 1 || c.Repeat > 3 || len(row.Returns) != c.Repeat {
				t.Fatal("original call count")
			}
			ug.Value = server.Object{ObjClass: object.Class(c.Class), TypeInd: c.Type, Poison540: c.Poison, UpdateData: unsafe.Pointer(&dg.Value)}
			dg.Value = server.MonsterUpdateData{StatusFlags: object.MonsterStatus(c.Status), AIStackInd: c.Head, Field132: c.Prefix, Field127: c.MoveFrame}
			if c.Head >= 0 {
				dg.Value.AIStack[c.Head].Action = c.Action
			}
			hg.Value = server.HealthData{Cur: c.Cur, Max: c.Max}
			if c.HealthPresent {
				ug.Value.HealthData = &hg.Value
			}
			srv.SetFrame(c.Frame)
			srv.SetTickRate(c.FPS)
			if row.Entry == 0x534840 {
				setRegistry(c.Lookup)
			}
			statePredicateOriginalSwapMimicCache534840(c.Cache)
			before := make([][]byte, len(regions))
			for i, r := range regions {
				before[i] = snapshot(r.ptr, r.size)
			}
			cacheBefore := snapshot(cachePtr, 4)
			cacheRegion := unsafe.Add(cachePtr, -16)
			cacheGuardBefore := snapshot(cacheRegion, 36)
			if row.Entry == 0x534840 {
				if got := statePredicateOriginalNameIDControl534840(); got != uint32(c.Lookup) {
					t.Fatalf("ordinary actual C name-ID control got=%d want=%d", got, c.Lookup)
				}
				lookupControls++
				if !bytes.Equal(cacheBefore, snapshot(cachePtr, 4)) {
					t.Fatal("ordinary name-ID control changed predicate cache")
				}
			}
			target := statePredicateOriginalPublicC534710(row.Entry)
			if target == nil {
				t.Fatal("missing actual C entry", row.Entry)
			}
			entries[row.Entry]++
			serverReads = 0
			if !controlOnly {
				for call, want := range row.Returns {
					got := uint32(ccall.CallIntPtr(target, ug.Value.CObj()))
					if got != want {
						t.Errorf("actual C entry=%#x call=%d got=%d original=%d", row.Entry, call, got, want)
					}
					totalCalls++
				}
				if got := *(*uint32)(cachePtr); got != row.CacheAfter {
					t.Errorf("original cache got=%08x want=%08x", got, row.CacheAfter)
				}
				if row.Entry == 0x534840 && serverReads != len(row.Lookups) {
					t.Errorf("actual name lookup count got=%d original=%d", serverReads, len(row.Lookups))
				}
				if row.Entry == 0x534810 && serverReads != c.Repeat {
					t.Errorf("recent-move server binding calls got=%d want=%d", serverReads, c.Repeat)
				}
			}
			for i, r := range regions {
				if !bytes.Equal(before[i], snapshot(r.ptr, r.size)) {
					t.Errorf("actual C changed native input/guard region%d", i)
				}
			}
			cacheGuardAfter := snapshot(cacheRegion, 36)
			if !bytes.Equal(cacheGuardBefore[:16], cacheGuardAfter[:16]) || !bytes.Equal(cacheGuardBefore[20:], cacheGuardAfter[20:]) {
				t.Error("state predicate changed neighboring global cache bytes")
			}
			if !controlOnly && row.Entry != 0x534840 && row.Entry != 0x534810 && serverReads != 0 {
				t.Error("pure state predicate accessed the server")
			}
			if srv.Frame() != c.Frame || srv.TickRate() != c.FPS {
				t.Error("state predicate changed server frame/FPS")
			}
		})
	}
	expected := map[uint32]int{0x534710: 8300, 0x5347a0: 544, 0x5347c0: 404, 0x5347f0: 1024, 0x534810: 3968, 0x534840: 2200}
	if len(entries) != len(expected) {
		t.Fatal("original root coverage")
	}
	for entry, want := range expected {
		if entries[entry] != want {
			t.Fatalf("entry%#x contexts=%d want=%d", entry, entries[entry], want)
		}
	}
	if lookupControls != 2200 || (!controlOnly && totalCalls != 16740) {
		t.Fatal("actual C or lookup-control count", totalCalls, lookupControls)
	}
	t.Logf("original actual C state contexts=%d calls=%d entries=%d real-name-controls=%d control-only=%t", len(reference.Rows), totalCalls, len(entries), lookupControls, controlOnly)
}
