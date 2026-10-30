package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type deathServicesNativeServer544CE0 struct {
	Server
	srv     *server.Server
	create  func(*server.Object, *server.Object, types.Pointf)
	deleted []*server.Object
}

func (s *deathServicesNativeServer544CE0) S() *server.Server { return s.srv }
func (s *deathServicesNativeServer544CE0) CreateObjectAt(obj, owner server.Obj, pos types.Pointf) {
	var o *server.Object
	if owner != nil {
		o = owner.SObj()
	}
	s.create(obj.SObj(), o, pos)
}
func (s *deathServicesNativeServer544CE0) DelayedDelete(obj *server.Object) {
	s.deleted = append(s.deleted, obj)
}

type deathServicesCRow544CE0 struct {
	Input struct {
		Name         string    `json:"name"`
		Entry        string    `json:"entry"`
		Owner        int       `json:"owner"`
		OwnerClass   uint32    `json:"owner_class"`
		PlayerIndex  byte      `json:"player_index"`
		Subclass     uint32    `json:"subclass"`
		Cache        uint32    `json:"cache"`
		Position     [2]uint32 `json:"position"`
		Direction    uint16    `json:"direction"`
		LookupValues []uint32  `json:"lookup_values"`
		Allocations  []int     `json:"allocations"`
		Repeats      int       `json:"repeats"`
		Mutation     struct {
			Stage  string `json:"stage"`
			Action string `json:"action"`
			Value  int    `json:"value"`
		} `json:"mutation"`
	} `json:"input"`
	Result struct {
		Events []struct {
			Kind      string  `json:"kind"`
			Args      []int64 `json:"args"`
			Iteration int     `json:"iteration"`
			State     struct {
				Cache      uint32       `json:"cache"`
				Position   [2]uint32    `json:"position"`
				Directions [3][2]uint16 `json:"directions"`
			} `json:"state"`
		} `json:"events"`
		Snapshots []struct {
			Cache      uint32       `json:"cache"`
			Position   [2]uint32    `json:"position"`
			Directions [3][2]uint16 `json:"directions"`
		} `json:"snapshots"`
	} `json:"result"`
}

// Real type registration/index lookup/allocation, shield packet construction,
// minimap updates and the non-coop reward gate run. World creation/deletion,
// packet delivery and scorch are declared observed boundaries.
func TestMonsterDeathServicesNativeComposition544CE0(t *testing.T) {
	path := "../server/testdata/monster_death_services_native_original_544ce0.json"
	if p := os.Getenv("NOX_MONSTER_DEATH_SERVICES_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []deathServicesCRow544CE0 `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	selected := make(map[string]deathServicesCRow544CE0)
	for _, row := range fixture.Rows {
		c, m := row.Input, row.Input.Mutation
		if c.Entry != "soul" || c.LookupValues[0] != 17 {
			continue
		}
		if strings.HasPrefix(c.Name, "soul-position-bits-") {
			selected[c.Name] = row
		} else if strings.HasPrefix(c.Name, "soul-cache-") && (c.Direction == 0 || c.Direction == 0xffff) && ((c.Cache == 0 || c.Cache == 17) && c.Allocations[0] == 0 || (c.Cache == 0x80000000 || c.Cache == 0xffffffff) && c.Allocations[0] == -1) {
			selected[c.Name] = row
			if c.Cache == 0 && c.Direction == 0 {
				selected["replace-name-after-first-create"] = row
			}
		} else if m.Stage == "create" && c.Allocations[0] == 0 && (m.Action == "unit-direction" || m.Action == "position" || m.Action == "soul-direction" || (m.Action == "cache" || m.Action == "all") && m.Value == 1) {
			selected[c.Name] = row
		}
	}
	if len(selected) != 25 {
		t.Fatal("frozen 8cache/8position/8create-mutation/1name-replacement contexts", len(selected))
	}
	modes := []string{"burn-no-owner", "burn-monster-owner", "burn-player", "burn-player-reload", "burn-owner-update-replaced", "burn-owner-replaced"}
	for _, row := range fixture.Rows {
		if _, ok := selected[row.Input.Name]; ok {
			modes = append(modes, row.Input.Name)
		}
	}
	modes = append(modes, "replace-name-after-first-create")
	const child = "NOX_MONSTER_DEATH_SERVICES_COMPOSITION_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDeathServicesNativeComposition544CE0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual public C death services: %v", err)
				}
			})
		}
		return
	}
	mode := os.Getenv(child)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(16) {
		t.Fatal("native allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	outer := &deathServicesNativeServer544CE0{srv: s}
	oldServer := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldServer })
	unit := s.Objs.NewObject(&server.ObjectType{})
	unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, 0x12340080, 0x12000004
	unit.PosVec = types.Ptf(math.Float32frombits(0x43a08000), math.Float32frombits(0x44238000))
	unit.Direction1, unit.Direction2 = 13, 0xbabe
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unit.CObj()) <= 0xffffffff {
		t.Fatal("actual native unit must exceed4GiB")
	}
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), int(n)) }
	if strings.HasPrefix(mode, "burn-") {
		type guarded struct {
			Before  [32]byte
			Owners  [2]server.Object
			Updates [2]server.PlayerUpdateData
			Marks   [2]server.MinimapItem
			After   [32]byte
		}
		p, free := alloc.New(guarded{})
		t.Cleanup(free)
		*p = guarded{}
		for i := range p.Before {
			p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
		}
		players := [2]*server.Player{s.Players.ByIndRaw(4), s.Players.ByIndRaw(9)}
		for i := range p.Owners {
			p.Owners[i].ObjClass = object.ClassPlayer
			p.Owners[i].UpdateData = unsafe.Pointer(&p.Updates[i])
			p.Updates[i].Player = players[i]
			players[i].PlayerInd = byte(4 + 5*i)
			p.Marks[i] = server.MinimapItem{Field0: 3, Field4: unit, Field8: &p.Marks[i], Field12: &p.Marks[i]}
			players[i].Field4580 = &p.Marks[i]
		}
		unit.ObjOwner = &p.Owners[0]
		if mode == "burn-no-owner" {
			unit.ObjOwner = nil
		}
		if mode == "burn-monster-owner" {
			p.Owners[0].ObjClass = object.ClassMonster
		}
		initialUnit, initialExtra := *unit, *p
		initialPlayers := [2]server.Player{*players[0], *players[1]}
		var packets [][]byte
		var recipients []int
		var positions []types.Pointf
		s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, seq int) int {
			if related != nil || remove != 1 || seq != 1 {
				t.Fatal("shield packet boundary metadata")
			}
			packets = append(packets, append([]byte(nil), packet...))
			recipients = append(recipients, recipient)
			if mode == "burn-player-reload" {
				p.Updates[0].Player = players[1]
			}
			if mode == "burn-owner-update-replaced" {
				p.Owners[0].UpdateData = unsafe.Pointer(&p.Updates[1])
			}
			if mode == "burn-owner-replaced" {
				unit.ObjOwner = &p.Owners[1]
			}
			return 1
		}
		oldScorch := makeScorchCall537AF0
		makeScorchCall537AF0 = func(pos types.Pointf, size int) {
			if size != 1 {
				t.Fatal("scorch size")
			}
			positions = append(positions, pos)
		}
		t.Cleanup(func() { makeScorchCall537AF0 = oldScorch })
		zombieBurnDeleteNative544CE0(unit)
		wantUnit, wantExtra := *unit, *p
		wantPlayers := [2]server.Player{*players[0], *players[1]}
		playerPath := mode != "burn-no-owner" && mode != "burn-monster-owner"
		if len(outer.deleted) != 1 || outer.deleted[0] != unit || len(positions) != 1 || positions[0] != unit.PosVec {
			t.Fatal("ordinary reward/scorch/delete boundary")
		}
		if playerPath {
			if len(packets) != 1 || len(packets[0]) != 3 || binary.LittleEndian.Uint16(packets[0][1:]) != uint16(unit.NetCode) || recipients[0] != 4 || uint32(unit.ObjSubClass) != 0x12340000 {
				t.Fatal("ordinary shield/subclass")
			}
			n := 0
			if mode == "burn-player-reload" {
				n = 1
			}
			if p.Marks[n].Field0 != 2 || p.Marks[1-n].Field0 != 3 {
				t.Fatal("cached owner update and reloaded Player pointer minimap")
			}
		} else if len(packets) != 0 || uint32(unit.ObjSubClass) != 0x12340080 {
			t.Fatal("ordinary non-player path")
		}
		wantPackets, wantRecipients, wantPositions := packets, recipients, positions
		*unit, *p = initialUnit, initialExtra
		*players[0], *players[1] = initialPlayers[0], initialPlayers[1]
		outer.deleted = nil
		packets = nil
		recipients = nil
		positions = nil
		t.Logf("DEATH SERVICES CONTROL COMPLETED: mode=%s native-unit=%p native-owner=%p", mode, unit, unit.ObjOwner)
		ccall.CallVoidPtr(monsterBurnDeletePublicC544CE0(), unit.CObj())
		if len(outer.deleted) != 1 || outer.deleted[0] != unit || !bytes.Equal(raw(unit.CObj(), unsafe.Sizeof(*unit)), raw(unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit))) || !bytes.Equal(raw(unsafe.Pointer(p), unsafe.Sizeof(*p)), raw(unsafe.Pointer(&wantExtra), unsafe.Sizeof(wantExtra))) {
			t.Fatal("native burn identity/full records/canaries")
		}
		for i := range players {
			if !bytes.Equal(raw(unsafe.Pointer(players[i]), unsafe.Sizeof(*players[i])), raw(unsafe.Pointer(&wantPlayers[i]), unsafe.Sizeof(wantPlayers[i]))) {
				t.Fatal("native complete Player record")
			}
		}
		a, _ := json.Marshal([]any{packets, recipients, positions})
		b, _ := json.Marshal([]any{wantPackets, wantRecipients, wantPositions})
		if !bytes.Equal(a, b) {
			t.Fatal("native packet/scorch boundary order/arguments")
		}
		t.Logf("DEATH SERVICES NATIVE C PASS: mode=%s full-records=true", mode)
		return
	}
	row, ok := selected[mode]
	if !ok {
		t.Fatal("unknown frozen original C context", mode)
	}
	c := row.Input
	registerTypes := func() {
		s.FreeObjectTypes()
		for i := 1; i <= 16; i++ {
			if err := s.Types.ReadObjectType(&things.Thing{Name: strings.Repeat("x", i)}); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.Types.ReadObjectType(&things.Thing{Name: "ReleasedSoul", Class: []things.Class{"SIMPLE"}, Flags: []things.Flag{"NO_COLLIDE"}}); err != nil {
			t.Fatal(err)
		}
		if err := s.Types.ReadObjectType(&things.Thing{Name: "AlternateSoul", Class: []things.Class{"SIMPLE"}, Flags: []things.Flag{"NO_COLLIDE"}}); err != nil {
			t.Fatal(err)
		}
		if s.Types.IndByID("ReleasedSoul") != 17 || s.Types.IndByID("AlternateSoul") != 18 {
			t.Fatal("original type-index service mapping")
		}
	}
	cache := memmap.PtrUint32(0x5D4594, 2489456)
	oldCache := *cache
	t.Cleanup(func() { *cache = oldCache })
	unit.PosVec = types.Ptf(math.Float32frombits(c.Position[0]), math.Float32frombits(c.Position[1]))
	unit.Direction1 = server.Dir16(c.Direction)
	initialUnit := *unit
	for _, actualC := range []bool{false, true} {
		registerTypes()
		*unit = initialUnit
		*cache = c.Cache
		created := 0
		currentIteration := 0
		expectedType := uint32(0)
		var souls []*server.Object
		var soulExpected []server.Object
		outer.create = func(obj, owner *server.Object, pos types.Pointf) {
			var eventFound bool
			for _, event := range row.Result.Events {
				if event.Iteration != currentIteration {
					continue
				}
				if event.Kind == "new" {
					expectedType = uint32(event.Args[0])
				}
				if event.Kind != "create" {
					continue
				}
				eventFound = true
				if obj == nil || owner != nil || uint32(obj.TypeInd) != expectedType || [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)} != [2]uint32{uint32(event.Args[2]), uint32(event.Args[3])} || *cache != event.State.Cache {
					t.Errorf("original real allocation/create boundary actualC=%t iteration=%d type=%d want=%d cache=%x want=%x", actualC, currentIteration, obj.TypeInd, expectedType, *cache, event.State.Cache)
				}
			}
			if !eventFound {
				t.Error("unexpected actual create")
				return
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.CObj()) <= 0xffffffff {
				t.Fatal("real allocated soul must exceed4GiB")
			}
			before := *obj
			m := c.Mutation
			if m.Stage == "create" {
				if m.Action == "cache" || m.Action == "all" {
					*cache = 18
				}
				if m.Action == "unit-direction" || m.Action == "all" {
					unit.Direction1 = 0x1234
					if m.Value != 0 {
						unit.Direction1 = 0xffff
					}
				}
				if m.Action == "position" || m.Action == "all" {
					unit.PosVec = types.Ptf(-37.5, 19.25)
				}
				if m.Action == "soul-direction" || m.Action == "all" {
					obj.Direction1, obj.Direction2 = 0xa55a, 0x5aa5
				}
			}
			want := row.Result.Snapshots[currentIteration].Directions[1]
			before.Direction1, before.Direction2 = server.Dir16(want[0]), server.Dir16(want[1])
			souls = append(souls, obj)
			soulExpected = append(soulExpected, before)
			created++
			if mode == "replace-name-after-first-create" && currentIteration == 0 {
				if err := s.Types.ReadObjectType(&things.Thing{Name: "ReleasedSoul", Class: []things.Class{"SIMPLE"}, Flags: []things.Flag{"NO_COLLIDE"}}); err != nil {
					t.Fatal(err)
				}
			}
		}
		for currentIteration = 0; currentIteration < c.Repeats; currentIteration++ {
			if actualC {
				ccall.CallVoidPtr(monsterReleasedSoulPublicC544E60(), unit.CObj())
			} else {
				monsterActionDeadRuntime544D80().CreateReleasedSoul(unit)
			}
			want := row.Result.Snapshots[currentIteration]
			if *cache != want.Cache || [2]uint32{math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y)} != want.Position || [2]uint16{uint16(unit.Direction1), uint16(unit.Direction2)} != want.Directions[0] {
				t.Errorf("original native unit/cache snapshot actualC=%t iteration=%d cache=%x want=%x", actualC, currentIteration, *cache, want.Cache)
			}
		}
		wantCreated := 0
		for _, event := range row.Result.Events {
			if event.Kind == "create" {
				wantCreated++
			}
		}
		if created != wantCreated {
			t.Errorf("original allocation outcomes actualC=%t created=%d want=%d", actualC, created, wantCreated)
		}
		for i, obj := range souls {
			if !bytes.Equal(raw(obj.CObj(), unsafe.Sizeof(*obj)), raw(unsafe.Pointer(&soulExpected[i]), unsafe.Sizeof(soulExpected[i]))) {
				t.Error("whole real allocated soul differs beyond original directions")
			}
		}
		wantUnit := initialUnit
		final := row.Result.Snapshots[len(row.Result.Snapshots)-1]
		wantUnit.PosVec = types.Ptf(math.Float32frombits(final.Position[0]), math.Float32frombits(final.Position[1]))
		wantUnit.Direction1, wantUnit.Direction2 = server.Dir16(final.Directions[0][0]), server.Dir16(final.Directions[0][1])
		if !bytes.Equal(raw(unit.CObj(), unsafe.Sizeof(*unit)), raw(unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit))) {
			t.Error("whole native unit differs beyond original callback mutations")
		}
		if actualC {
			t.Logf("DEATH SERVICES NATIVE C PASS: mode=%s created=%d full-records=true", mode, created)
		} else {
			t.Logf("DEATH SERVICES CONTROL COMPLETED: mode=%s native-unit=%p created=%d", mode, unit, created)
		}
	}
}
