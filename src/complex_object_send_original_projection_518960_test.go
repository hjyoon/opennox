package opennox

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
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func senderOrdinaryRows518960(t *testing.T) []struct {
	Name string
	Row  senderRow518960
} {
	t.Helper()
	root := os.Getenv("NOX_COMPLEX_OBJECT_SEND_REFERENCE_DIR")
	if root == "" {
		root = "server/testdata/complex_object_send_518960"
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

type senderNetworkInput518960 struct {
	Player      int32  `json:"player"`
	Report      uint32 `json:"report"`
	HP          bool   `json:"hp_present"`
	CurrentHP   uint16 `json:"hp"`
	PreviousHP  uint16 `json:"previous_hp"`
	Frame       uint32 `json:"frame"`
	LastFrame   uint32 `json:"last_frame"`
	Class       uint32 `json:"object_class"`
	Subclass    uint32 `json:"subclass"`
	Primary     uint32 `json:"primary_code"`
	Secondary   uint32 `json:"secondary_code"`
	X           uint32 `json:"x"`
	Y           uint32 `json:"y"`
	Direction   int16  `json:"direction"`
	Sprite      uint8  `json:"sprite_frame"`
	OtherSprite uint8  `json:"other_sprite_frame"`
	NPC         uint8  `json:"NPC_random"`
	OtherNPC    uint8  `json:"other_NPC_random"`
	RNG         uint32 `json:"rng"`
	Frames      int32  `json:"player_frames"`
	Duration    int32  `json:"player_duration"`
	Return      uint32 `json:"queue_return"`
}

type senderRow518960 struct {
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Output struct {
		Return    uint32           `json:"return_word"`
		Returns   []uint32         `json:"returns"`
		Events    []map[string]any `json:"events"`
		Mutations []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
			Call  int             `json:"call"`
		} `json:"mutations"`
		Previous uint16    `json:"previous_hp_after"`
		RNG      uint32    `json:"RNG_index_after"`
		Caches   [4]uint32 `json:"caches_after"`
		Guards   bool      `json:"full_heap_stack_global_neighbors_nonvolatile_and_FP_control_guards_preserved"`
	} `json:"output"`
}

type animationInput533790 struct {
	Head        int8      `json:"head"`
	Action      uint32    `json:"action"`
	Status      uint32    `json:"status"`
	Enemy       bool      `json:"enemy"`
	OtherStatus uint32    `json:"other_status"`
	OtherEnemy  bool      `json:"other_enemy"`
	Class       uint32    `json:"class"`
	Flags       uint32    `json:"flags"`
	Type        uint16    `json:"type_index"`
	Caches      [4]uint32 `json:"caches"`
	Lookup      [4]uint16 `json:"lookup"`
	Found       [4]bool   `json:"found"`
	Repeat      int       `json:"repeat"`
	Mutation    *struct {
		Name       string `json:"name"`
		Occurrence int    `json:"occurrence"`
		Effects    []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
		} `json:"effects"`
	} `json:"mutation"`
}

func TestComplexObjectPacketExistingGoProjection518960(t *testing.T) {
	rows := senderOrdinaryRows518960(t)
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	s := &Server{Server: srv}
	for i := 1; i <= 44; i++ {
		name := fmt.Sprintf("ComplexSendProjectionType%d", i)
		for j, v := range []int{11, 22, 33, 44} {
			if i == v {
				name = []string{"Mimic", "CarnivorousPlant", "Zombie", "VileZombie"}[j]
			}
		}
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	if srv.Types.MimicID() != 11 {
		t.Fatal("real Mimic cache")
	}
	types := srv.TypePredicateRuntime534A10()
	types.Write(0, 22)
	types.Write(1, 33)
	types.Write(2, 44)
	unit, freeU := alloc.New(server.Object{})
	t.Cleanup(freeU)
	update, freeD := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(freeD)
	health, freeH := alloc.New(server.HealthData{})
	t.Cleanup(freeH)
	enemy, freeE := alloc.New(server.Object{})
	t.Cleanup(freeE)
	for index, mode := range rows {
		t.Run(mode.Name, func(t *testing.T) {
			var n senderNetworkInput518960
			var c animationInput533790
			if err := json.Unmarshal(mode.Row.Input, &n); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(mode.Row.Input, &c); err != nil {
				t.Fatal(err)
			}
			draw := things.PlayerDraw{Anims: map[player.AnimType]*things.PlayerAnim{player.AnimElectricZap: {FramesPerDir: byte(n.Frames), Field8: byte(n.Duration)}}}
			if err := srv.Types.ReadObjectType(&things.Thing{Name: fmt.Sprintf("ComplexProjectionFrames%d", index), Draw: draw}); err != nil {
				t.Fatal(err)
			}
			*unit = server.Object{TypeInd: c.Type, ObjClass: object.Class(n.Class), ObjSubClass: object.SubClass(n.Subclass), ObjFlags: object.Flags(0x1000000), UpdateData: unsafe.Pointer(update), NetCode: n.Primary, Extent: n.Secondary, Frame134: n.LastFrame, Direction1: server.Dir16(n.Direction)}
			unit.PosVec.X, unit.PosVec.Y = math.Float32frombits(n.X), math.Float32frombits(n.Y)
			*health = server.HealthData{Cur: n.CurrentHP, Max: 65535}
			unit.HealthData = nil
			if n.HP {
				unit.HealthData = health
			}
			*update = server.MonsterUpdateData{AIStackInd: c.Head, StatusFlags: object.MonsterStatus(c.Status), Field120_1: n.Sprite, Field523_2: n.NPC}
			*enemy = server.Object{}
			if c.Enemy {
				update.CurrentEnemy = enemy
			}
			item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&update.AIStack[0]), int(c.Head)*int(unsafe.Sizeof(update.AIStack[0]))))
			item.Action = c.Action
			update.HealthGraph103[n.Player] = n.PreviousHP
			srv.SetFrame(n.Frame)
			srv.Rand.Logic.Reset(int(n.RNG))
			snapshot := func(ptr unsafe.Pointer, n uintptr) []byte {
				return append([]byte(nil), unsafe.Slice((*byte)(ptr), n)...)
			}
			beforeU, beforeD, beforeH, beforeE := snapshot(unsafe.Pointer(unit), unsafe.Sizeof(*unit)), snapshot(unsafe.Pointer(update), unsafe.Sizeof(*update)), snapshot(unsafe.Pointer(health), unsafe.Sizeof(*health)), snapshot(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))
			got := s.complexObjectPacketNative518960(unit)
			want := ""
			for _, e := range mode.Row.Output.Events {
				if e["service"] == "0x40ef40" {
					want = e["packet"].(string)
				}
			}
			if want == "" {
				t.Fatal("missing original packet")
			}
			if hex.EncodeToString(got[:]) != want {
				t.Errorf("packet=%x original=%s", got, want)
			}
			if uint32(srv.Rand.Logic.Index()) != mode.Row.Output.RNG {
				t.Errorf("RNG=%d original=%d", srv.Rand.Logic.Index(), mode.Row.Output.RNG)
			}
			if !bytes.Equal(beforeU, snapshot(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) || !bytes.Equal(beforeD, snapshot(unsafe.Pointer(update), unsafe.Sizeof(*update))) || !bytes.Equal(beforeH, snapshot(unsafe.Pointer(health), unsafe.Sizeof(*health))) || !bytes.Equal(beforeE, snapshot(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))) {
				t.Error("packet-only native input memory changed")
			}
		})
	}
	t.Log("23 frozen original-mode packet/RNG projection only; excludes health publication,enqueue return,helper entry and global C caches; no rounding mask")
}
