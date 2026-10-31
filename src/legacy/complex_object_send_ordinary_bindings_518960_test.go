package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func senderOrdinaryRows518960(t *testing.T) []struct {
	Name string
	Row  senderRow518960
} {
	t.Helper()
	root := os.Getenv("NOX_COMPLEX_OBJECT_SEND_REFERENCE_DIR")
	if root == "" {
		root = "../server/testdata/complex_object_send_518960"
	}
	raw, err := os.ReadFile(root + "/phaseA-startup-v2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "f80a75388bee334824454bfc763f2819dea1573e822e6de5ddf1b639daf7f174" {
		t.Fatal("original sender seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Rows []senderRow518960 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.Rows) != 14110 {
		t.Fatal("original sender dimensions")
	}
	byName := make(map[string]senderRow518960)
	for _, r := range ref.Rows {
		byName[r.Name] = r
	}
	modes := []struct{ Name, Row string }{
		{"east", "00000/all-directions"},
		{"northeast edge", "00046/all-directions"},
		{"west edge", "00111/all-directions"},
		{"direction wrap", "00255/all-directions"},
		{"fractional coordinates", "00037/all-directions"},
		{"cached Mimic idle", "02426/warm-selector-matrix"},
		{"cached Plant idle", "02407/warm-selector-matrix"},
		{"cached burning Zombie", "03938/warm-selector-matrix"},
		{"NPC chance below threshold", "08856/actual-RNG-table-and-wrap"},
		{"NPC random frame", "08855/actual-RNG-table-and-wrap"},
		{"NPC RNG table wrap", "12950/actual-RNG-table-and-wrap"},
		{"NPC zero frames", "12957/NPC-frame-output-boundaries"},
		{"NPC frame255", "13002/NPC-frame-output-boundaries"},
		{"NPC duration zero", "12987/NPC-frame-output-boundaries"},
		{"report with nil health", "13669/health-admission"},
		{"actual health damage report", "13670/health-admission"},
		{"health healing updates sample", "13287/health-delta-and-frame-wrap"},
		{"health delta low-word wrap", "13545/health-delta-and-frame-wrap"},
		{"health age two skips", "13076/health-delta-and-frame-wrap"},
		{"unsigned frame subtraction wrap", "13379/health-delta-and-frame-wrap"},
		{"enqueue returns false", "14106/full-network-return-word"},
		{"invalid coordinates signed DWORD", "14100/actual-FISTP-coordinate-boundaries"},
		{"alternate network code", "13717/actual-network-code-boundaries"},
	}
	out := make([]struct {
		Name string
		Row  senderRow518960
	}, 0, len(modes))
	for _, m := range modes {
		r, ok := byName[m.Row]
		if !ok {
			t.Fatal("missing original ordinary mode", m.Row)
		}
		out = append(out, struct {
			Name string
			Row  senderRow518960
		}{m.Name, r})
	}
	return out
}

func TestComplexObjectSendOrdinaryBindings518960(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	rows := senderOrdinaryRows518960(t)
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for i := 1; i <= 44; i++ {
		name := fmt.Sprintf("ComplexSendBindingType%d", i)
		for j, v := range []int{11, 22, 33, 44} {
			if i == v {
				name = animationNames533790[j]
			}
		}
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	if srv.Types.MimicID() != 11 {
		t.Fatal("real Go Mimic cache")
	}
	types := srv.TypePredicateRuntime534A10()
	types.Write(0, 22)
	types.Write(1, 33)
	types.Write(2, 44)
	oldMimic := statePredicateOriginalSwapMimicCache534840(11)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldMimic) })
	b := &animationFixture533790{t: t, srv: srv}
	unit := animationAllocate533790[server.Object](t, b)
	update := animationAllocate533790[server.MonsterUpdateData](t, b)
	health := animationAllocate533790[server.HealthData](t, b)
	enemy := animationAllocate533790[server.Object](t, b)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantListHeadsC(0, 0)
	setImportantAllocClassC(nil)
	setImportantRecipientMaskC(0)
	owner := Nox_netlist_addToMsgListSrv
	t.Cleanup(func() { Nox_netlist_addToMsgListSrv = owner })
	calls := [3]int{}
	for index, mode := range rows {
		t.Run(mode.Name, func(t *testing.T) {
			var n senderNetworkInput518960
			var s animationInput533790
			if err := json.Unmarshal(mode.Row.Input, &n); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(mode.Row.Input, &s); err != nil {
				t.Fatal(err)
			}
			draw := things.PlayerDraw{Anims: map[player.AnimType]*things.PlayerAnim{player.AnimElectricZap: {FramesPerDir: byte(n.Frames), Field8: byte(n.Duration)}}}
			if err := srv.Types.ReadObjectType(&things.Thing{Name: fmt.Sprintf("ComplexSendFrames%d", index), Draw: draw}); err != nil {
				t.Fatal(err)
			}
			if a, b := srv.PlayerAnimFrames(50); a != int(n.Frames) || b != int(n.Duration) {
				t.Fatal("real PlayerDraw index50")
			}
			for path := 0; path < 3; path++ {
				t.Run([]string{"C", "publicGo", "packetGo"}[path], func(t *testing.T) {
					*unit = animationGuard533790[server.Object]{}
					*update = animationGuard533790[server.MonsterUpdateData]{}
					*health = animationGuard533790[server.HealthData]{}
					*enemy = animationGuard533790[server.Object]{}
					u, d, h := &unit.Value, &update.Value, &health.Value
					*u = server.Object{TypeInd: s.Type, ObjClass: object.Class(n.Class), ObjSubClass: object.SubClass(n.Subclass), ObjFlags: object.Flags(0x1000000), UpdateData: unsafe.Pointer(d), NetCode: n.Primary, Extent: n.Secondary, Frame134: n.LastFrame, Direction1: server.Dir16(n.Direction)}
					u.PosVec.X, u.PosVec.Y = math.Float32frombits(n.X), math.Float32frombits(n.Y)
					*h = server.HealthData{Cur: n.CurrentHP, Max: 65535}
					if n.HP {
						u.HealthData = h
					}
					*d = server.MonsterUpdateData{AIStackInd: s.Head, StatusFlags: object.MonsterStatus(s.Status), Field120_1: n.Sprite, Field523_2: n.NPC}
					if s.Enemy {
						d.CurrentEnemy = &enemy.Value
					}
					item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&d.AIStack[0]), int(s.Head)*int(unsafe.Sizeof(d.AIStack[0]))))
					item.Action = s.Action
					d.HealthGraph103[n.Player] = n.PreviousHP
					srv.SetFrame(n.Frame)
					srv.Rand.Logic.Reset(int(n.RNG))
					beforeU := animationSnapshot533790(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
					beforeD := animationSnapshot533790(unsafe.Pointer(update), unsafe.Sizeof(*update))
					beforeH := animationSnapshot533790(unsafe.Pointer(health), unsafe.Sizeof(*health))
					beforeE := animationSnapshot533790(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))
					mimic := animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)
					events := []map[string]any{}
					drain := func() {
						recipient, data := complexObjectTakeImportantPacketC518960()
						if data != nil {
							events = append(events, map[string]any{"service": "0x4e5420", "player": uint32(recipient), "packet": hex.EncodeToString(data)})
						}
						first, last := importantNativeListC()
						if first != nil || last != nil {
							t.Fatal("extra health publication")
						}
					}
					Nox_netlist_addToMsgListSrv = func(player ntype.PlayerInd, packet []byte) bool {
						drain()
						events = append(events, map[string]any{"service": "0x40ef40", "player": uint32(player), "packet": hex.EncodeToString(packet)})
						return n.Return != 0
					}
					var got int32
					switch path {
					case 0:
						got = complexObjectSendPublicC518960(n.Player, u, int32(n.Report))
					case 1:
						got = NetSendComplexObject518960(n.Player, u, int32(n.Report))
					case 2:
						packet := srv.ComplexObjectPacket518960(u)
						events = append(events, map[string]any{"service": "0x40ef40", "player": uint32(n.Player), "packet": hex.EncodeToString(packet[:])})
					}
					calls[path]++
					drain()
					want := []map[string]any{}
					for _, e := range mode.Row.Output.Events {
						if e["service"] == "0x40ef40" || (path != 2 && e["service"] == "0x4e5420") {
							want = append(want, e)
						}
					}
					if !bytes.Equal(mustSenderJSON518960(events), mustSenderJSON518960(want)) {
						t.Errorf("publication/packet got=%v original=%v", events, want)
					}
					if path != 2 {
						if uint32(got) != mode.Row.Output.Return {
							t.Errorf("sender return=%#x original=%#x", uint32(got), mode.Row.Output.Return)
						}
						binary.LittleEndian.PutUint16(beforeD[unsafe.Offsetof(update.Value)+unsafe.Offsetof(d.HealthGraph103)+uintptr(2*n.Player):], mode.Row.Output.Previous)
					}
					if uint32(srv.Rand.Logic.Index()) != mode.Row.Output.RNG {
						t.Errorf("RNG=%d original=%d", srv.Rand.Logic.Index(), mode.Row.Output.RNG)
					}
					if !bytes.Equal(beforeU, animationSnapshot533790(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) || !bytes.Equal(beforeD, animationSnapshot533790(unsafe.Pointer(update), unsafe.Sizeof(*update))) || !bytes.Equal(beforeH, animationSnapshot533790(unsafe.Pointer(health), unsafe.Sizeof(*health))) || !bytes.Equal(beforeE, animationSnapshot533790(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))) {
						t.Error("whole native guarded unit/MUD/health/enemy differs beyond original HP sample")
					}
					if !bytes.Equal(mimic, animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)) {
						t.Error("Mimic cache/neighbors changed")
					}
					if types.Read(0) != 22 || types.Read(1) != 33 || types.Read(2) != 44 || srv.Frame() != n.Frame {
						t.Error("type caches or global frame changed")
					}
				})
			}
		})
	}
	if calls != [3]int{23, 23, 23} {
		t.Fatal("ordinary sender call dimensions", calls)
	}
	t.Log("ordinary sender:23 original modes each actual-C/publicGo/pureGoPacket; actual default factories,PlayerDraw50,registry,RNG and health publication; no provider substitution")
}
func mustSenderJSON518960(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
