package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type obeliskInputOriginal struct {
	Kind           string    `json:"kind"`
	Name           string    `json:"name"`
	Count          int       `json:"count"`
	Source         [2]uint32 `json:"source"`
	Position       [2]uint32 `json:"position"`
	Position2      [2]uint32 `json:"position2"`
	Mana           int32     `json:"mana"`
	AlternateMana  int32     `json:"alternate_mana"`
	Cur            uint16    `json:"cur"`
	Maximum        uint16    `json:"maximum"`
	Progress       int32     `json:"progress"`
	Charge         uint8     `json:"charge"`
	MaxCharge      uint8     `json:"max_charge"`
	Amount         int32     `json:"amount"`
	Rate           int32     `json:"rate"`
	Slot           int       `json:"slot"`
	ItemClass      uint32    `json:"item_class"`
	ItemSubclass   uint32    `json:"item_subclass"`
	Balance        uint32    `json:"balance"`
	Multipliers    [3]uint32 `json:"multipliers"`
	PlayerClass    uint8     `json:"player_class"`
	Quest          bool      `json:"quest"`
	Online         bool      `json:"online"`
	Frame          uint32    `json:"frame"`
	TickRate       uint32    `json:"tick_rate"`
	LastFrame      uint32    `json:"last_frame"`
	Dead           bool      `json:"dead"`
	Team           uint8     `json:"team"`
	PlayerTeam     uint8     `json:"player_team"`
	Vision         bool      `json:"vision"`
	NilItem        bool      `json:"nil_item"`
	NilUse         bool      `json:"nil_use"`
	MutationEvent  string    `json:"mutation_event"`
	MutationAction string    `json:"mutation_action"`
	ZeroObjects    bool      `json:"zero_objects"`
}

type obeliskStateOriginal struct {
	ObjectWords   [][3]uint32 `json:"object_words"`
	ObjectUpdates []int       `json:"object_updates"`
	ObjectUses    []int       `json:"object_uses"`
	Mana          []uint32    `json:"mana"`
	Updates       [][4]int    `json:"updates"`
	Wands         [][3]uint32 `json:"wands"`
	Players       [][3]uint32 `json:"players"`
	Flags         uint32      `json:"flags"`
	Frame         uint32      `json:"frame"`
	NextPlayer    int         `json:"next_player"`
}

type obeliskEventOriginal struct {
	Service string `json:"service"`
	Object  int    `json:"object"`
	Value   uint32 `json:"value"`
	Maximum uint32 `json:"maximum"`
	Index   uint32 `json:"index"`
	Token   uint32 `json:"token"`
	obeliskStateOriginal
}

type obeliskRowOriginal struct {
	Mode        string                 `json:"mode"`
	Input       obeliskInputOriginal   `json:"input"`
	Events      []obeliskEventOriginal `json:"events"`
	Mutated     bool                   `json:"mutated"`
	ReturnRaw   uint32                 `json:"return_raw"`
	ReturnOwner bool                   `json:"return_owner_address"`
	Preserved   bool                   `json:"nonoutput_bytes_preserved"`
	obeliskStateOriginal
}

func obeliskFrameOriginal[T any](t *testing.T) (*T, []byte) {
	t.Helper()
	var zero T
	frame, free := alloc.Make([]byte(nil), int(unsafe.Sizeof(zero))+64)
	t.Cleanup(free)
	for n := range frame {
		frame[n] = 0xa5
	}
	p := (*T)(unsafe.Pointer(&frame[32]))
	*p = zero
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("native frame below 4 GiB")
	}
	return p, frame
}

func TestObeliskStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/obelisk_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string               `json:"original_sha256"`
		Unchanged  bool                 `json:"original_unchanged"`
		CW         uint16               `json:"actual_startup_control_word"`
		Contexts   int                  `json:"contexts"`
		Executions int                  `json:"root_executions"`
		Faults     []json.RawMessage    `json:"memory_faults"`
		Rows       []obeliskRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !f.Unchanged || f.CW != 0xe3f || f.Contexts != 772 || f.Executions != 1544 || len(f.Rows) != 1544 || len(f.Faults) != 0 {
		t.Fatal("incomplete original Obelisk fixture")
	}
	count := 0
	for _, row := range f.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) { testObeliskOriginal(t, row) })
	}
	if count != 772 {
		t.Fatalf("startup conditions=%d", count)
	}
}

func testObeliskOriginal(t *testing.T, row obeliskRowOriginal) {
	t.Helper()
	i := row.Input
	var objs [5]*Object
	var objectFrames [5][]byte
	var mana [2]*ObeliskUpdateData
	var manaFrames [2][]byte
	var uds [3]*PlayerUpdateData
	var udFrames [3][]byte
	var players [3]*Player
	var playerFrames [3][]byte
	var wands [3]*WandUseData
	var wandFrames [3][]byte
	var attrs [2]*ModifierInitData
	var attrFrames [2][]byte
	var effects [2]*ModifierEff
	var effectFrames [2][]byte
	for n := range objs {
		objs[n], objectFrames[n] = obeliskFrameOriginal[Object](t)
	}
	for n := range mana {
		mana[n], manaFrames[n] = obeliskFrameOriginal[ObeliskUpdateData](t)
	}
	for n := range uds {
		uds[n], udFrames[n] = obeliskFrameOriginal[PlayerUpdateData](t)
		players[n], playerFrames[n] = obeliskFrameOriginal[Player](t)
		wands[n], wandFrames[n] = obeliskFrameOriginal[WandUseData](t)
	}
	for n := range attrs {
		attrs[n], attrFrames[n] = obeliskFrameOriginal[ModifierInitData](t)
		effects[n], effectFrames[n] = obeliskFrameOriginal[ModifierEff](t)
	}
	rep, repFrame := obeliskFrameOriginal[uint32](t)
	*rep = 0x12345678
	obelisk := objs[0]
	positions := [5][2]uint32{i.Source, i.Position, i.Position2, {0, 0}, {0, 0}}
	for n, p := range objs {
		p.PosVec = rigidPointOriginal(positions[n])
		p.ObjClass = object.Class([5]uint32{0, 4, 4, i.ItemClass, 0x1000}[n])
		if !i.ZeroObjects {
			p.Field38 = 0x11223344
		}
		if n == 0 {
			p.Field34 = i.LastFrame
			p.TeamVal.ID = TeamID(i.Team)
		}
		if n == 1 || n == 2 {
			p.TeamVal.ID = TeamID(i.PlayerTeam)
		}
		if n == 1 && i.Dead {
			p.ObjFlags = object.FlagDead
		}
		if n >= 3 {
			p.ObjSubClass = object.SubClass(0x10000)
			p.InitData = unsafe.Pointer(attrs[n-3])
			p.UseData.Ptr = unsafe.Pointer(wands[n-3])
		}
	}
	objs[3].ObjSubClass = object.SubClass(i.ItemSubclass)
	if i.NilUse {
		objs[3].UseData.Ptr = nil
	}
	obelisk.UpdateData = unsafe.Pointer(mana[0])
	objs[1].UpdateData = unsafe.Pointer(uds[0])
	objs[2].UpdateData = unsafe.Pointer(uds[1])
	mana[0].Mana = i.Mana
	mana[1].Mana = i.AlternateMana
	for n, p := range uds {
		p.ManaCur = i.Cur
		p.ManaMax = i.Maximum
		p.EquippedWeapon = objs[3]
		p.Player = players[n]
		if n == 2 {
			p.ManaCur = 0
			p.ManaMax = 20
			p.EquippedWeapon = objs[4]
		}
	}
	for n, p := range players {
		p.PlayerInd = uint8(7 + n)
		p.Info().SetPlayerClass(player.Class(i.PlayerClass))
		p.ProtUnitManaCur = 0x12345678 + uint32(n)*0x11111111
		if n == 2 {
			p.Info().SetPlayerClass(player.Conjurer)
		}
	}
	for n, p := range wands {
		p.Charge = 1
		p.MaxCharge = 20
		p.Progress = 5
		if n == 0 {
			p.Charge = i.Charge
			p.MaxCharge = i.MaxCharge
			p.Progress = uint32(i.Progress)
		}
	}
	for n, p := range effects {
		p.Attack40.Fnc = unsafe.Pointer(rep)
		p.Attack40.Val = i.Rate
		if n == 1 {
			p.Attack40.Val = 3
		}
		if i.Slot == 2 || i.Slot == 3 {
			attrs[n].Modifiers[i.Slot] = p
		}
	}
	allUpdates := []unsafe.Pointer{unsafe.Pointer(mana[0]), unsafe.Pointer(mana[1]), unsafe.Pointer(uds[0]), unsafe.Pointer(uds[1]), unsafe.Pointer(uds[2])}
	ptrIndex := func(p unsafe.Pointer, list []unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, v := range list {
			if v == p {
				return n
			}
		}
		t.Fatal("unknown native pointer")
		return -2
	}
	objIndex := func(p *Object) int {
		if p == nil {
			return -1
		}
		for n, v := range objs {
			if p == v {
				return n
			}
		}
		t.Fatal("unknown native object")
		return -2
	}
	playerIndex := func(p *Player) int {
		if p == nil {
			return -1
		}
		for n, v := range players {
			if p == v {
				return n
			}
		}
		t.Fatal("unknown native player")
		return -2
	}
	wandPointers := []unsafe.Pointer{unsafe.Pointer(wands[0]), unsafe.Pointer(wands[1]), unsafe.Pointer(wands[2])}
	flags := uint32(0)
	if i.Quest {
		flags |= 0x1000
	}
	if i.Online {
		flags |= 0x2000
	}
	frame := i.Frame
	nextPlayer := -1
	if i.Count == 2 {
		nextPlayer = 1
	}
	state := func() obeliskStateOriginal {
		s := obeliskStateOriginal{ObjectWords: make([][3]uint32, 5), ObjectUpdates: make([]int, 5), ObjectUses: make([]int, 5), Mana: make([]uint32, 2), Updates: make([][4]int, 3), Wands: make([][3]uint32, 3), Players: make([][3]uint32, 3), Flags: flags, Frame: frame, NextPlayer: nextPlayer}
		for n, p := range objs {
			s.ObjectWords[n] = [3]uint32{uint32(p.ObjFlags), p.Field34, p.Field38}
			s.ObjectUpdates[n] = ptrIndex(p.UpdateData, allUpdates)
			s.ObjectUses[n] = ptrIndex(p.UseData.Ptr, wandPointers)
		}
		for n, p := range mana {
			s.Mana[n] = uint32(p.Mana)
		}
		for n, p := range uds {
			s.Updates[n] = [4]int{int(p.ManaCur), int(p.ManaMax), objIndex(p.EquippedWeapon), playerIndex(p.Player)}
		}
		for n, p := range wands {
			s.Wands[n] = [3]uint32{uint32(p.Charge), uint32(p.MaxCharge), p.Progress}
		}
		for n, p := range players {
			s.Players[n] = [3]uint32{uint32(p.PlayerInd), uint32(p.PlayerClass()), p.ProtUnitManaCur}
		}
		return s
	}
	clone := func(frames [][]byte) [][]byte {
		out := make([][]byte, len(frames))
		for n, p := range frames {
			out[n] = bytes.Clone(p)
		}
		return out
	}
	wantObjects, wantMana, wantUD, wantWands := clone(objectFrames[:]), clone(manaFrames[:]), clone(udFrames[:]), clone(wandFrames[:])
	wantPlayers, wantAttrs, wantEffects := clone(playerFrames[:]), clone(attrFrames[:]), clone(effectFrames[:])
	wantRep := bytes.Clone(repFrame)
	write32 := func(raw []byte, offset uintptr, value uint32) { binary.LittleEndian.PutUint32(raw[32+offset:], value) }
	write16 := func(raw []byte, offset uintptr, value uint16) { binary.LittleEndian.PutUint16(raw[32+offset:], value) }
	writePtr := func(raw []byte, offset uintptr, value unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(raw[32+offset:], uint64(uintptr(value)))
		} else {
			binary.LittleEndian.PutUint32(raw[32+offset:], uint32(uintptr(value)))
		}
	}
	for n, p := range objs {
		write32(wantObjects[n], unsafe.Offsetof(p.ObjFlags), row.ObjectWords[n][0])
		write32(wantObjects[n], unsafe.Offsetof(p.Field34), row.ObjectWords[n][1])
		write32(wantObjects[n], unsafe.Offsetof(p.Field38), row.ObjectWords[n][2])
		var update, use unsafe.Pointer
		if row.ObjectUpdates[n] >= 0 {
			update = allUpdates[row.ObjectUpdates[n]]
		}
		if row.ObjectUses[n] >= 0 {
			use = wandPointers[row.ObjectUses[n]]
		}
		writePtr(wantObjects[n], unsafe.Offsetof(p.UpdateData), update)
		writePtr(wantObjects[n], unsafe.Offsetof(p.UseData)+unsafe.Offsetof(p.UseData.Ptr), use)
	}
	for n := range mana {
		write32(wantMana[n], 0, row.Mana[n])
	}
	for n, p := range uds {
		w := row.Updates[n]
		write16(wantUD[n], unsafe.Offsetof(p.ManaCur), uint16(w[0]))
		write16(wantUD[n], unsafe.Offsetof(p.ManaMax), uint16(w[1]))
		var weapon, pl unsafe.Pointer
		if w[2] >= 0 {
			weapon = unsafe.Pointer(objs[w[2]])
		}
		if w[3] >= 0 {
			pl = unsafe.Pointer(players[w[3]])
		}
		writePtr(wantUD[n], unsafe.Offsetof(p.EquippedWeapon), weapon)
		writePtr(wantUD[n], unsafe.Offsetof(p.Player), pl)
	}
	for n, p := range wands {
		wantWands[n][32+unsafe.Offsetof(p.Charge)] = byte(row.Wands[n][0])
		wantWands[n][32+unsafe.Offsetof(p.MaxCharge)] = byte(row.Wands[n][1])
		write32(wantWands[n], unsafe.Offsetof(p.Progress), row.Wands[n][2])
	}
	events := make([]obeliskEventOriginal, 0)
	mutated := false
	observe := func(e obeliskEventOriginal) { e.obeliskStateOriginal = state(); events = append(events, e) }
	mutate := func(stage string) {
		if mutated || i.MutationEvent != stage {
			return
		}
		mutated = true
		switch i.MutationAction {
		case "weapon":
			uds[0].EquippedWeapon = objs[4]
		case "player":
			uds[0].Player = players[2]
		case "unit-update":
			objs[1].UpdateData = unsafe.Pointer(uds[2])
		case "obelisk-update":
			obelisk.UpdateData = unsafe.Pointer(mana[1])
		case "wand-use":
			objs[3].UseData.Ptr = unsafe.Pointer(wands[2])
		case "mana-max":
			uds[0].ManaMax = 8
		case "mana-cur":
			uds[0].ManaCur = 12
		case "quest":
			flags = 0x1000
		case "normal":
			flags = 0
		case "frame":
			frame = 200
		case "drop-next":
			nextPlayer = -1
		default:
			t.Fatal("unknown live action")
		}
	}
	units := []*Object{}
	if i.Count >= 1 {
		units = append(units, objs[1])
	}
	if i.Count == 2 {
		units = append(units, objs[2])
	}
	deps := obeliskUpdateNativeDeps53C580{
		firstPlayerUnit: func() *Object {
			if len(units) == 0 {
				return nil
			}
			return units[0]
		},
		nextPlayerUnit: func(p *Object) *Object {
			if len(units) > 1 && p == units[0] && nextPlayer == 1 {
				return units[1]
			}
			return nil
		},
		mapTraceVision: func(o, p *Object) bool {
			if o != obelisk {
				t.Fatal("vision source changed")
			}
			observe(obeliskEventOriginal{Service: "vision", Object: objIndex(p)})
			mutate("vision")
			return i.Vision
		},
		frame: func() uint32 { return frame }, tickRate: func() uint32 { return i.TickRate }, questMode: func() bool { return flags&0x1000 != 0 }, onlineMode: func() bool { return flags&0x2000 != 0 },
		balanceFloat: func(name string) float32 {
			if name != "OblivionStaffRechargeRate" {
				t.Fatal("balance key changed")
			}
			observe(obeliskEventOriginal{Service: "balance", Value: i.Balance})
			mutate("balance")
			return math.Float32frombits(i.Balance)
		},
		questManaMultiplier: func(c player.Class) float32 { return math.Float32frombits(i.Multipliers[int(c)]) }, replenishmentEffect: unsafe.Pointer(rep),
		protectMana: func(token uint32, value int16) {
			observe(obeliskEventOriginal{Service: "protect", Token: token, Value: uint32(uint16(value))})
			mutate("protect")
		},
		protectPlayerHPMana: func(token uint32, value uint16) {
			observe(obeliskEventOriginal{Service: "cap", Token: token, Value: uint32(value)})
			mutate("cap")
		},
		reportCharges: func(index uint8, p *Object, value, maximum uint8) {
			observe(obeliskEventOriginal{Service: "report", Object: objIndex(p), Value: uint32(value), Maximum: uint32(maximum), Index: uint32(index)})
			mutate("report")
		},
		needSync:          func(p *Object) { observe(obeliskEventOriginal{Service: "sync", Object: objIndex(p)}); p.NeedSync() },
		audioManaRecharge: func(p *Object) { observe(obeliskEventOriginal{Service: "audio", Object: objIndex(p)}); mutate("audio") },
	}
	got := int64(0)
	switch i.Kind {
	case "main":
		got = int64(obeliskUpdateNative53C580(obelisk, deps))
	case "charge":
		if obeliskRechargeItem53C520(objs[3], i.Amount) {
			got = 1
		}
	case "rate":
		p := objs[3]
		if i.NilItem {
			p = nil
		}
		got = int64(obeliskRechargeRate53C940(p, deps))
	default:
		t.Fatal("unknown original entry")
	}
	want := int64(int32(row.ReturnRaw))
	if row.ReturnOwner {
		want = int64(int(uintptr(unsafe.Pointer(obelisk))))
	}
	if !row.Preserved || got != want || mutated != row.Mutated || !reflect.DeepEqual(state(), row.obeliskStateOriginal) || !reflect.DeepEqual(events, row.Events) {
		t.Fatalf("native result=%d original=%d mutated=%t/%t state=%+v original=%+v events=%+v original=%+v", got, want, mutated, row.Mutated, state(), row.obeliskStateOriginal, events, row.Events)
	}
	check := func(frames, want [][]byte) {
		for n, p := range frames {
			if !bytes.Equal(p, want[n]) {
				t.Fatalf("native frame %d bytes differ", n)
			}
		}
	}
	check(objectFrames[:], wantObjects)
	check(manaFrames[:], wantMana)
	check(udFrames[:], wantUD)
	check(wandFrames[:], wantWands)
	check(playerFrames[:], wantPlayers)
	check(attrFrames[:], wantAttrs)
	check(effectFrames[:], wantEffects)
	if !bytes.Equal(repFrame, wantRep) {
		t.Fatal("effect token frame changed")
	}
}
