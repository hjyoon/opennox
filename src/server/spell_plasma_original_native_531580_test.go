package server

import (
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

//go:embed testdata/spell_plasma_native_original_531580.json
var plasmaOriginalData531580 []byte

type plasmaOriginalState531580 struct {
	Caster        int          `json:"caster"`
	Target        int          `json:"target"`
	Weapon        int          `json:"weapon"`
	Ray           int          `json:"ray"`
	Grace         uint32       `json:"grace"`
	Expiry        uint32       `json:"expiry"`
	Flags         uint32       `json:"flags"`
	Position      [2]uint32    `json:"position"`
	FPS           uint32       `json:"fps"`
	Frame         uint32       `json:"frame"`
	Caches        [2]uint32    `json:"caches"`
	Selected      int          `json:"selected"`
	Nearest       uint32       `json:"nearest"`
	Types         [8]uint16    `json:"types"`
	Classes       [8]uint32    `json:"classes"`
	Subclasses    [8]uint32    `json:"subclasses"`
	ObjectFlags   [8]uint32    `json:"object_flags"`
	Positions     [8][2]uint32 `json:"positions"`
	Directions    [8]uint16    `json:"directions"`
	Updates       [8]int       `json:"updates"`
	UseData       [8]int       `json:"use_data"`
	Equipped      [8]int       `json:"equipped"`
	Cursors       [8]int       `json:"cursors"`
	PlayerLinks   [8]int       `json:"player_links"`
	PlayerIndices [8]uint8     `json:"player_indices"`
	Wands         [8][4]uint32 `json:"wands"`
}
type plasmaOriginalInput531580 struct {
	plasmaOriginalState531580
	Name           string            `json:"name"`
	Kind           string            `json:"kind"`
	Lookup         [2]uint32         `json:"lookup"`
	PositionResult uint32            `json:"position_result"`
	Facing         [8][8]uint32      `json:"facing"`
	Enemy          [8][8]uint32      `json:"enemy"`
	LOS            [8][8]uint32      `json:"los"`
	DistanceBits64 [8][8]string      `json:"distance_bits64"`
	Balances       map[string]string `json:"balances"`
	Candidates     []int             `json:"candidates"`
	Mutation       struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}
type plasmaOriginalEvent531580 struct {
	Kind  string                    `json:"kind"`
	Args  []json.RawMessage         `json:"args"`
	State plasmaOriginalState531580 `json:"state"`
}
type plasmaOriginalResult531580 struct {
	Ret    *uint32                     `json:"ret"`
	State  plasmaOriginalState531580   `json:"state"`
	Events []plasmaOriginalEvent531580 `json:"events"`
}

func plasmaOriginalPoint531580(p [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(p[0]), math.Float32frombits(p[1]))
}
func plasmaOriginalBits531580(p types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
}
func plasmaOriginalDouble531580(s string) float64 {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
}

// The producer executes every byte of the original roots, candidate and
// converter under the actual startup FPU profile. Services observe complete
// state before applying the declared mutation. Native-width identity maps
// compare links without reusing PE32 addresses. The duration-owner removal of
// the weapon sidecar after destroy is explicitly checked as a native lifetime
// adaptation: the original root retains that dying record's weapon link.
func TestSpellPlasmaOriginalNative531580(t *testing.T) {
	digest := sha256.Sum256(plasmaOriginalData531580)
	if hex.EncodeToString(digest[:]) != "2e3906893ead709e19ac7aec289d7e41ff76116de7ea0915d5b7ce7aab6b24c9" {
		t.Fatal("sealed original fixture hash")
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		ControlWord    int    `json:"control_word"`
		Rows           []struct {
			Input  plasmaOriginalInput531580  `json:"input"`
			Result plasmaOriginalResult531580 `json:"result"`
		} `json:"rows"`
	}
	if e := json.Unmarshal(plasmaOriginalData531580, &fixture); e != nil {
		t.Fatal(e)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 581 {
		t.Fatal("original identity/profile/contexts")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			objs := new([8]Object)
			updates := new([8]PlayerUpdateData)
			players := new([8]Player)
			wands := new([8]WandUseData)
			obj := func(i int) *Object {
				if i < 0 {
					return nil
				}
				return &objs[i]
			}
			idx := func(o *Object) int {
				if o == nil {
					return -1
				}
				for i := range objs {
					if o == &objs[i] {
						return i
					}
				}
				panic("foreign object")
			}
			ui := func(p unsafe.Pointer) int {
				for i := range updates {
					if p == unsafe.Pointer(&updates[i]) {
						return i
					}
				}
				panic("foreign update")
			}
			wi := func(p unsafe.Pointer) int {
				for i := range wands {
					if p == unsafe.Pointer(&wands[i]) {
						return i
					}
				}
				panic("foreign wand")
			}
			pi := func(p *Player) int {
				for i := range players {
					if p == &players[i] {
						return i
					}
				}
				panic("foreign player")
			}
			for i := range objs {
				objs[i] = Object{TypeInd: in.Types[i], ObjClass: object.Class(in.Classes[i]), ObjSubClass: object.SubClass(in.Subclasses[i]), ObjFlags: object.Flags(in.ObjectFlags[i]), PosVec: plasmaOriginalPoint531580(in.Positions[i]), Direction1: Dir16(in.Directions[i]), UpdateData: unsafe.Pointer(&updates[in.Updates[i]])}
				objs[i].UseData.Ptr = unsafe.Pointer(&wands[in.UseData[i]])
				updates[i].EquippedWeapon = obj(in.Equipped[i])
				updates[i].CursorObj = obj(in.Cursors[i])
				updates[i].Player = &players[in.PlayerLinks[i]]
				players[i].PlayerInd = in.PlayerIndices[i]
				wands[i] = WandUseData{Flags: in.Wands[i][0], Charge: uint8(in.Wands[i][1]), MaxCharge: uint8(in.Wands[i][2]), Progress: in.Wands[i][3]}
			}
			r := &DurSpell{Caster16: obj(in.Caster), Target48: obj(in.Target), Field76: uintptr(in.Grace), Frame68: in.Expiry, Flags88: in.Flags, Pos: plasmaOriginalPoint531580(in.Position)}
			weapon, ray := obj(in.Weapon), obj(in.Ray)
			fps, frame, caches := in.FPS, in.Frame, in.Caches
			selected, nearest := obj(in.Selected), math.Float32frombits(in.Nearest)
			rt := SpellPlasmaRuntime531580{}
			// Reflection keeps this exact comparator runnable against the pre-port
			// baseline. Only that baseline lacks the native selection scratch/Balance64
			// fields; its unavailable scratch observations are projected out below.
			scratchField := reflect.ValueOf(&rt).Elem().FieldByName("Selection")
			scratch := reflect.Value{}
			if scratchField.IsValid() {
				scratch = reflect.New(scratchField.Type().Elem())
				scratchField.Set(scratch)
			}
			storeScratch := func() {
				if scratch.IsValid() {
					scratch.Elem().FieldByName("Target").Set(reflect.ValueOf(selected))
					scratch.Elem().FieldByName("Nearest").SetFloat(float64(nearest))
				}
			}
			loadScratch := func() {
				if scratch.IsValid() {
					selected = scratch.Elem().FieldByName("Target").Interface().(*Object)
					nearest = float32(scratch.Elem().FieldByName("Nearest").Float())
				}
			}
			storeScratch()
			snapshot := func() plasmaOriginalState531580 {
				loadScratch()
				s := plasmaOriginalState531580{Caster: idx(r.Caster16), Target: idx(r.Target48), Weapon: idx(weapon), Ray: idx(ray), Grace: uint32(r.Field76), Expiry: r.Frame68, Flags: r.Flags88, Position: plasmaOriginalBits531580(r.Pos), FPS: fps, Frame: frame, Caches: caches, Selected: idx(selected), Nearest: math.Float32bits(nearest)}
				for i := range objs {
					o, d, w := &objs[i], &updates[i], &wands[i]
					s.Types[i] = o.TypeInd
					s.Classes[i] = uint32(o.ObjClass)
					s.Subclasses[i] = uint32(o.ObjSubClass)
					s.ObjectFlags[i] = uint32(o.ObjFlags)
					s.Positions[i] = plasmaOriginalBits531580(o.PosVec)
					s.Directions[i] = uint16(o.Direction1)
					s.Updates[i] = ui(o.UpdateData)
					s.UseData[i] = wi(o.UseData.Ptr)
					s.Equipped[i] = idx(d.EquippedWeapon)
					s.Cursors[i] = idx(d.CursorObj)
					s.PlayerLinks[i] = pi(d.Player)
					s.PlayerIndices[i] = players[i].PlayerInd
					s.Wands[i] = [4]uint32{w.Flags, uint32(w.Charge), uint32(w.MaxCharge), w.Progress}
				}
				return s
			}
			events := make([]plasmaOriginalEvent531580, 0)
			calls := map[string]int{}
			changed := false
			event := func(kind string, args ...any) {
				raw := make([]json.RawMessage, len(args))
				for i, v := range args {
					raw[i], _ = json.Marshal(v)
				}
				events = append(events, plasmaOriginalEvent531580{kind, raw, snapshot()})
				at := calls[kind]
				calls[kind]++
				m := in.Mutation
				if changed || m.Stage != kind || m.At != at {
					return
				}
				changed = true
				switch m.Action {
				case "caster-b":
					r.Caster16 = obj(1)
				case "caster-nil":
					r.Caster16 = nil
				case "target-b":
					r.Target48 = obj(3)
				case "target-nil":
					r.Target48 = nil
				case "weapon-b":
					weapon = obj(7)
				case "weapon-nil":
					weapon = nil
				case "ray-b":
					ray = obj(4)
				case "ray-nil":
					ray = nil
				case "grace":
					r.Field76 = 2
				case "grace-zero":
					r.Field76 = 0
				case "expiry":
					r.Frame68 = 0x13579bdf
				case "record-flags":
					r.Flags88 ^= 0x20
				case "record-position":
					r.Pos = types.Ptf(99.25, -88.5)
				case "fps":
					fps = 60
				case "frame":
					frame = 0xfffffff9
				case "cache-hec":
					caches[0] = 20
				case "cache-orb":
					caches[1] = 20
				case "selected-b":
					selected = obj(5)
					storeScratch()
				case "nearest":
					nearest = 4
					storeScratch()
				case "caster-update-b":
					objs[0].UpdateData = unsafe.Pointer(&updates[1])
				case "equipped-b":
					updates[0].EquippedWeapon = obj(7)
				case "cursor-b":
					updates[0].CursorObj = obj(3)
				case "player-b":
					updates[0].Player = &players[1]
				case "caster-position":
					objs[0].PosVec = types.Ptf(3, 4)
				case "target-position":
					objs[2].PosVec = types.Ptf(1.25, -2.5)
				case "target-dead":
					objs[2].ObjFlags |= object.Flags(0x8020)
				case "wand-flags":
					wands[6].Flags = 0x12345678
				case "wand-charge":
					wands[6].Charge = 1
				case "wand-max":
					wands[6].MaxCharge = 9
				case "wand-data-b":
					objs[6].UseData.Ptr = unsafe.Pointer(&wands[7])
				default:
					t.Fatalf("unknown mutation %q", m.Action)
				}
			}
			rt.HecubahType = &caches[0]
			rt.HecubahOrbType = &caches[1]
			rt.LookupType = func(key string) uint32 {
				event("lookup", key)
				if key == "Hecubah" {
					return in.Lookup[0]
				}
				return in.Lookup[1]
			}
			rt.Frame = func() uint32 { return frame }
			rt.TickRate = func() uint32 { return fps }
			balance := func(key string) float64 { event("balance", key); return plasmaOriginalDouble531580(in.Balances[key]) }
			rt.Balance = func(key string) float32 { return float32(balance(key)) }
			if f := reflect.ValueOf(&rt).Elem().FieldByName("Balance64"); f.IsValid() {
				f.Set(reflect.ValueOf(balance))
			}
			rt.PositionDelta = func(o *Object, p *types.Pointf) int32 {
				v := plasmaOriginalBits531580(*p)
				event("position", idx(o), v[0], v[1])
				return int32(in.PositionResult)
			}
			rt.Facing = func(a, b *Object) int {
				i, j := idx(a), idx(b)
				pa, pb := plasmaOriginalBits531580(a.PosVec), plasmaOriginalBits531580(b.PosVec)
				direction := int16(a.Direction1)
				event("facing", i, int32(direction), j, pa[0], pa[1], pb[0], pb[1])
				return int(in.Facing[i][j])
			}
			rt.Distance = func(a, b *Object) float64 {
				i, j := idx(a), idx(b)
				event("distance", i, j)
				return plasmaOriginalDouble531580(in.DistanceBits64[i][j])
			}
			rt.IsEnemy = func(a, b *Object) bool { i, j := idx(a), idx(b); event("enemy", i, j); return in.Enemy[i][j] != 0 }
			rt.CanInteract = func(a, b *Object) bool { i, j := idx(a), idx(b); event("los", i, j); return in.LOS[i][j] != 0 }
			rt.ObjectsInCircle = func(p types.Pointf, radius float32, visit func(*Object) bool) {
				// The original passes a fixed caster pointer as callback context even
				// when a service mutates record.Caster16 during enumeration.
				caster := r.Caster16
				v := plasmaOriginalBits531580(p)
				event("circle", idx(caster), v[0], v[1], math.Float32bits(radius))
				for _, i := range in.Candidates {
					event("visit", i, idx(caster))
					if !visit(obj(i)) {
						break
					}
				}
			}
			rt.StopRay = func(_ *DurSpell, o *Object) { event("stop", idx(o)) }
			rt.StartRay = func(_ *DurSpell) { event("start") }
			rt.Damage = func(a, b *Object, amount int32) { event("damage", idx(a), idx(b), amount) }
			rt.PointFX = func(code uint8, p types.Pointf) {
				owner := -1
				if in.Kind != "create" {
					owner = idx(r.Target48)
				}
				v := plasmaOriginalBits531580(p)
				event("fx", code, owner, v[0], v[1])
			}
			rt.SetPlayerState = func(o *Object, state PlayerState) { event("state", idx(o), state) }
			rt.Audio = func(id uint16, o *Object) { event("audio", id, idx(o)) }
			rt.ReportCharges = func(d *PlayerUpdateData, o *Object, c, m uint8) { event("report", d.Player.PlayerInd, idx(o), c, m) }
			rt.LoadWeapon = func(*DurSpell) *Object { return weapon }
			rt.StoreWeapon = func(_ *DurSpell, o *Object) { weapon = o }
			rt.LoadRayTarget = func(*DurSpell) *Object { return ray }
			rt.StoreRayTarget = func(_ *DurSpell, o *Object) { ray = o }
			var ret *uint32
			switch in.Kind {
			case "create":
				v := uint32(SpellPlasmaCreate531580(r, rt))
				ret = &v
			case "update":
				v := uint32(SpellPlasmaUpdate531600(r, rt))
				ret = &v
			case "destroy":
				SpellPlasmaDestroy5319E0(r, rt)
			default:
				t.Fatal("unknown root")
			}
			if in.Mutation.Stage != "" && !changed {
				t.Fatal("declared mutation not reached")
			}
			got := plasmaOriginalResult531580{ret, snapshot(), events}
			want := row.Result
			if in.Kind == "destroy" {
				if want.State.Weapon != in.Weapon {
					t.Fatal("original destroy must retain weapon link")
				}
				want.State.Weapon = -1
			}
			if !scratch.IsValid() {
				t.Log("pre-port baseline: native scratch fields unavailable; only scratch observations projected out, selection/event behavior remains compared")
				got.State.Selected = 0
				got.State.Nearest = 0
				want.State.Selected = 0
				want.State.Nearest = 0
				for i := range got.Events {
					got.Events[i].State.Selected = 0
					got.Events[i].State.Nearest = 0
				}
				for i := range want.Events {
					want.Events[i].State.Selected = 0
					want.Events[i].State.Nearest = 0
				}
			}
			if !reflect.DeepEqual(got, want) {
				actual, _ := json.Marshal(got)
				expected, _ := json.Marshal(want)
				t.Fatalf("got %s\nwant %s", actual, expected)
			}
		})
	}
}
