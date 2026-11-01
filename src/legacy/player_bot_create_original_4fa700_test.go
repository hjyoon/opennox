package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerBotCreateRow4FA700 struct {
	Input struct {
		Name      string `json:"name"`
		Class     byte   `json:"cls"`
		FPS       uint32 `json:"fps"`
		RNG       uint32 `json:"rng"`
		Bot       string `json:"bot"`
		Fail      bool   `json:"allocator_fail"`
		Health    uint16 `json:"health"`
		Delta     uint32 `json:"delta"`
		Index     int8   `json:"index"`
		Action    uint32 `json:"action"`
		Status    uint32 `json:"status"`
		Setting   uint32 `json:"setting"`
		World     string `json:"world"`
		Nil       string `json:"nil"`
		Direction uint16 `json:"direction"`
		X         uint32 `json:"x"`
		Y         uint32 `json:"y"`
	} `json:"input"`
	Output struct {
		Result  uint32           `json:"result"`
		RNG     uint32           `json:"rng"`
		Events  []map[string]any `json:"events"`
		Records map[string]struct {
			Bytes int    `json:"bytes"`
			SHA   string `json:"sha256"`
			Hex   string `json:"hex"`
		} `json:"records"`
		Heap    bool             `json:"all_other_heap_bytes_unchanged"`
		Stack   bool             `json:"caller_stack_unchanged"`
		CW      bool             `json:"CW_TOP_preserved"`
		Fault   any              `json:"exception"`
		Invalid []map[string]any `json:"invalid"`
	} `json:"output"`
}

func playerBotCreateRead4FA700(t *testing.T, name, encodedSHA, decodedSHA string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "server", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != encodedSHA {
		t.Fatal("original gzip seal", name)
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	b, err = io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if decodedSHA != "" && fmt.Sprintf("%x", sha256.Sum256(b)) != decodedSHA {
		t.Fatal("original decoded seal", name)
	}
	return b
}

func playerBotCreateLoad4FA700(t *testing.T) ([]playerBotCreateRow4FA700, []playerBotCreateRow4FA700) {
	t.Helper()
	raw := playerBotCreateRead4FA700(t, "player_bot_create_original_4fa700.json.gz", "8313dfa30f1d70450e5b11cc63142eca5efb23818624b51acb6e16f27a92c433", "3926be9f931772b502c6fd52c0346da5dc136184e05a4dd8dbab9b0f8978ec98")
	var d struct {
		Contexts []playerBotCreateRow4FA700 `json:"contexts"`
		Faults   []playerBotCreateRow4FA700 `json:"faults"`
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Contexts) != 945 || len(d.Faults) != 3 {
		t.Fatal("original Create dimensions")
	}
	return d.Contexts, d.Faults
}

type playerBotCreateRNG4FA700 struct {
	Seed  uint32 `json:"seed"`
	Value int32  `json:"value"`
	Next  uint32 `json:"next_index"`
}

func playerBotCreateRNGLoad4FA700(t *testing.T) map[uint32]playerBotCreateRNG4FA700 {
	t.Helper()
	raw := playerBotCreateRead4FA700(t, "player_bot_create_original_rng_4fa700.json.gz", "ac7fe6030fbe81d5f2ec8271861c6108e8a29be7ecb59be44dc36dc840ef0fb4", "")
	var rows []playerBotCreateRNG4FA700
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4097 {
		t.Fatal("original bounded RNG observations")
	}
	out := make(map[uint32]playerBotCreateRNG4FA700)
	for _, r := range rows {
		out[r.Seed] = r
	}
	return out
}

type playerBotCreateWorld4FA700 struct {
	Unit            *server.Object
	PUD, PUD2       *server.PlayerUpdateData
	MUD, MUD2       *server.MonsterUpdateData
	Player, Player2 *server.Player
	Definition      *server.MonsterDef
	Health          *server.HealthData
	Settings        *server.Settings
	Raw, Before     map[string][]byte
	Original        map[string][]byte
	Roles           map[uint32]uintptr
	Events          []map[string]any
	RNG             uint32
	Runtime         server.PlayerBotCreateRuntime4FA700
}

func playerBotCreateWorldNew4FA700(t *testing.T, row playerBotCreateRow4FA700, rng map[uint32]playerBotCreateRNG4FA700) *playerBotCreateWorld4FA700 {
	t.Helper()
	w := &playerBotCreateWorld4FA700{Raw: make(map[string][]byte), Before: make(map[string][]byte), Original: make(map[string][]byte), Roles: map[uint32]uintptr{0: 0}, RNG: row.Input.RNG}
	take := func(name string, size uintptr) unsafe.Pointer {
		p, free := alloc.Malloc(size + 64)
		t.Cleanup(free)
		raw := unsafe.Slice((*byte)(p), size+64)
		for i := range raw {
			raw[i] = 0xa5
		}
		w.Raw[name] = raw
		return unsafe.Add(p, 32)
	}
	w.Unit = (*server.Object)(take("unit", unsafe.Sizeof(server.Object{})))
	w.PUD = (*server.PlayerUpdateData)(take("pud", unsafe.Sizeof(server.PlayerUpdateData{})))
	w.PUD2 = (*server.PlayerUpdateData)(take("pud2", unsafe.Sizeof(server.PlayerUpdateData{})))
	w.MUD = (*server.MonsterUpdateData)(take("mud", unsafe.Sizeof(server.MonsterUpdateData{})))
	w.MUD2 = (*server.MonsterUpdateData)(take("mud2", unsafe.Sizeof(server.MonsterUpdateData{})))
	w.Player = (*server.Player)(take("player", unsafe.Sizeof(server.Player{})))
	w.Player2 = (*server.Player)(take("player2", unsafe.Sizeof(server.Player{})))
	w.Health = (*server.HealthData)(take("health", unsafe.Sizeof(server.HealthData{})))
	w.Settings = (*server.Settings)(take("settings", unsafe.Sizeof(server.Settings{})))
	w.Definition = (*server.MonsterDef)(take("definition", unsafe.Sizeof(server.MonsterDef{})))
	pointers := map[string]unsafe.Pointer{"unit": unsafe.Pointer(w.Unit), "pud": unsafe.Pointer(w.PUD), "pud2": unsafe.Pointer(w.PUD2), "mud": unsafe.Pointer(w.MUD), "mud2": unsafe.Pointer(w.MUD2), "player": unsafe.Pointer(w.Player), "player2": unsafe.Pointer(w.Player2), "health": unsafe.Pointer(w.Health), "settings": unsafe.Pointer(w.Settings), "definition": unsafe.Pointer(w.Definition)}
	original := map[string]uint32{"unit": 0x21001000, "pud": 0x21003000, "pud2": 0x21004000, "mud": 0x21006000, "mud2": 0x21008000, "player": 0x2100a000, "player2": 0x2100c000, "health": 0x2100e000, "settings": 0x2100f000, "definition": 0x21010000}
	for name, p := range pointers {
		if unsafe.Sizeof(p) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native record must exceed4GiB", name)
		}
		w.Roles[original[name]] = uintptr(p)
	}
	sizes := map[string]int{"unit": 800, "pud": 640, "pud2": 640, "mud": 2200, "mud2": 2200, "player": 5000, "player2": 5000, "health": 8, "settings": 128}
	for name, n := range sizes {
		w.Original[name] = bytes.Repeat([]byte{0xa5}, n)
	}
	put := func(name string, off int, v uint32) { binary.LittleEndian.PutUint32(w.Original[name][off:], v) }
	put16 := func(name string, off int, v uint16) { binary.LittleEndian.PutUint16(w.Original[name][off:], v) }
	w.Unit.ObjClass = object.Class(0x81220004)
	w.Unit.ObjSubClass = object.SubClass(0x13579bdf)
	w.Unit.PosVec.X = math.Float32frombits(row.Input.X)
	w.Unit.PosVec.Y = math.Float32frombits(row.Input.Y)
	w.Unit.Direction1 = server.Dir16(row.Input.Direction)
	w.Unit.HealthData = w.Health
	w.Health.Cur = row.Input.Health
	w.Unit.Update = playerBotCreateOriginalUpdateFn4FA700()
	w.Unit.UpdateData = unsafe.Pointer(w.PUD)
	w.PUD.Player = w.Player
	w.PUD2.Player = w.Player2
	w.PUD.Field73 = w.MUD
	w.PUD2.Field73 = w.MUD2
	w.MUD.Field545 = w.PUD
	w.MUD2.Field545 = w.PUD2
	w.Player.Info().SetPlayerClass(player.Class(row.Input.Class))
	w.Player2.Info().SetPlayerClass(player.Class(2))
	w.MUD.AIStackInd = row.Input.Index
	w.MUD2.AIStackInd = 0
	w.MUD.StatusFlags = object.MonsterStatus(row.Input.Status)
	w.MUD2.StatusFlags = w.MUD.StatusFlags
	w.MUD.Field137 = 100 - row.Input.Delta
	w.MUD2.Field137 = w.MUD.Field137
	if row.Input.Index >= 0 {
		w.MUD.AIStack[row.Input.Index].Action = row.Input.Action
		w.MUD2.AIStack[row.Input.Index].Action = 31
	}
	binary.LittleEndian.PutUint32(w.Settings.PlayerSkeletons58[:], row.Input.Setting)
	put("unit", 8, 0x81220004)
	put("unit", 12, 0x13579bdf)
	put("unit", 56, row.Input.X)
	put("unit", 60, row.Input.Y)
	put16("unit", 124, row.Input.Direction)
	put("unit", 556, original["health"])
	put("unit", 744, 0x4fab20)
	put("unit", 748, original["pud"])
	for _, name := range []string{"pud", "pud2"} {
		mm, pl := "mud", "player"
		class := row.Input.Class
		if name == "pud2" {
			mm, pl, class = "mud2", "player2", 2
		}
		put(name, 276, original[pl])
		put(name, 292, original[mm])
		put(mm, 2180, original[name])
		w.Original[pl][2251] = class
		index := row.Input.Index
		if name == "pud2" {
			index = 0
		}
		w.Original[mm][544] = byte(index)
		put(mm, 1440, row.Input.Status)
		put(mm, 548, 100-row.Input.Delta)
		if row.Input.Index >= 0 {
			action := row.Input.Action
			if name == "pud2" {
				action = 31
			}
			put(mm, 552+24*int(row.Input.Index), action)
		}
	}
	put16("health", 0, row.Input.Health)
	put("settings", 58, row.Input.Setting)
	if row.Input.Bot == "nil" {
		w.PUD.Field73 = nil
		put("pud", 292, 0)
	}
	if row.Input.Nil == "pud" {
		w.Unit.UpdateData = nil
		put("unit", 748, 0)
	}
	if row.Input.Nil == "player" {
		w.PUD.Player = nil
		put("pud", 276, 0)
	}
	for name, b := range w.Raw {
		w.Before[name] = bytes.Clone(b)
	}
	w.installRuntime(t, row, rng)
	return w
}

func (w *playerBotCreateWorld4FA700) installRuntime(t *testing.T, row playerBotCreateRow4FA700, rng map[uint32]playerBotCreateRNG4FA700) {
	w.Events = make([]map[string]any, 0)
	role := func(v uint32) string {
		switch v {
		case 0:
			return "nil"
		case 0x21006000:
			return "mud"
		case 0x21010000:
			return "definition"
		}
		return fmt.Sprintf("0x%x", v)
	}
	emit := func(name string, args []any, result uint32) {
		roles := make([]any, len(args))
		for i, arg := range args {
			if v, ok := arg.(uint32); ok {
				roles[i] = role(v)
			} else {
				roles[i] = arg
			}
		}
		w.Events = append(w.Events, map[string]any{"service": name, "args": args, "roles": roles, "result": result, "result_role": role(result)})
	}
	randomCalls := 0
	w.Runtime = server.PlayerBotCreateRuntime4FA700{
		Allocate: func() *server.MonsterUpdateData {
			result := uint32(0x21006000)
			if row.Input.Fail {
				result = 0
			}
			emit("allocate", []any{uint32(1), uint32(2200)}, result)
			if row.Input.Fail {
				return nil
			}
			return w.MUD
		},
		TypeID: func(name string) uint32 {
			if name != "NPC" {
				t.Fatal("original type lookup name", name)
			}
			emit("NPC-name", []any{name}, 77)
			if row.Input.World == "create-replace-pud" {
				w.Unit.UpdateData = unsafe.Pointer(w.PUD2)
			}
			return 77
		},
		Definition: func(id uint32) *server.MonsterDef {
			if id != 77 {
				t.Fatal("original definition typeID", id)
			}
			result := uint32(0x21010000)
			if row.Input.World == "nil-definition" {
				result = 0
			}
			emit("NPC-definition", []any{id}, result)
			if result == 0 {
				return nil
			}
			return w.Definition
		},
		FPS: func() uint32 { return row.Input.FPS },
		Random: func(min, max int32) int32 {
			randomCalls++
			if min != 0 || max != 100 || randomCalls != 1 {
				t.Fatal("original bounded RNG call", min, max, randomCalls)
			}
			observation, ok := rng[w.RNG]
			if !ok {
				t.Fatal("unobserved original RNG input", w.RNG)
			}
			w.RNG = observation.Next
			return observation.Value
		},
	}
}

func playerBotCreatePutNativeWord4FA700(dst []byte, off uintptr, value uintptr) {
	if unsafe.Sizeof(value) == 8 {
		binary.LittleEndian.PutUint64(dst[off:], uint64(value))
	} else {
		binary.LittleEndian.PutUint32(dst[off:], uint32(value))
	}
}

func (w *playerBotCreateWorld4FA700) compare(t *testing.T, row playerBotCreateRow4FA700) {
	t.Helper()
	if !row.Output.Heap || !row.Output.Stack || !row.Output.CW {
		t.Fatal("original heap/stack/FPU observations")
	}
	expected := make(map[string][]byte)
	for name, b := range w.Before {
		expected[name] = bytes.Clone(b)
	}
	for name, record := range row.Output.Records {
		after, err := hex.DecodeString(record.Hex)
		if err != nil || len(after) != record.Bytes || fmt.Sprintf("%x", sha256.Sum256(after)) != record.SHA {
			t.Fatal("complete original record seal", name)
		}
		before := w.Original[name]
		allowed := bytes.Clone(before)
		switch name {
		case "mud":
			if !bytes.Equal(before, after) {
				if binary.LittleEndian.Uint32(after) != 0xdeadface {
					t.Fatal("unexpected original mutation without reset")
				}
				native := playerBotCreateExpandOriginal4FA700(t, after, w.Roles)
				copy(expected[name][32:], native)
				copy(allowed, after)
			}
		case "unit":
			copy(allowed[748:752], after[748:752])
			value := binary.LittleEndian.Uint32(after[748:])
			pointer, ok := w.Roles[value]
			if !ok {
				t.Fatal("unknown original live unit record", value)
			}
			playerBotCreatePutNativeWord4FA700(expected[name][32:], unsafe.Offsetof(server.Object{}.UpdateData), pointer)
		case "pud", "pud2":
			copy(allowed[292:296], after[292:296])
			value := binary.LittleEndian.Uint32(after[292:])
			pointer, ok := w.Roles[value]
			if !ok {
				t.Fatal("unknown original bot link", value)
			}
			playerBotCreatePutNativeWord4FA700(expected[name][32:], unsafe.Offsetof(server.PlayerUpdateData{}.Field73), pointer)
		}
		if !bytes.Equal(allowed, after) {
			t.Fatal("unclassified original changed bytes", name)
		}
	}
	for name, actual := range w.Raw {
		if !bytes.Equal(expected[name], actual) {
			for i, v := range actual {
				if expected[name][i] != v {
					t.Fatalf("native whole record/guard %s byte%d:got%02x original%02x", name, i, v, expected[name][i])
				}
			}
		}
	}
	want := make([]map[string]any, 0, len(row.Output.Events))
	for _, event := range row.Output.Events {
		b, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var copyEvent map[string]any
		if err = json.Unmarshal(b, &copyEvent); err != nil {
			t.Fatal(err)
		}
		if copyEvent["service"] == "NPC-name" {
			copyEvent["args"] = []any{"NPC"}
			copyEvent["roles"] = []any{"NPC"}
		}
		want = append(want, copyEvent)
	}
	expectedJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	actualJSON, err := json.Marshal(w.Events)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expectedJSON, actualJSON) {
		t.Fatalf("ordered original services\ngot%s\nwant%s", actualJSON, expectedJSON)
	}
	if w.RNG != row.Output.RNG {
		t.Fatalf("original RNG final index:got%d want%d", w.RNG, row.Output.RNG)
	}
}

type playerBotCreateGo4FA700 interface {
	PlayerBotCreate4FA700(server.PlayerBotCreateRuntime4FA700) uint32
}

func playerBotCreateCall4FA700(path string, unit *server.Object, rt server.PlayerBotCreateRuntime4FA700) uint32 {
	if path == "C" {
		old := playerBotCreateRuntimeOverride4FA700
		playerBotCreateRuntimeOverride4FA700 = &rt
		defer func() { playerBotCreateRuntimeOverride4FA700 = old }()
		return playerBotCreateOriginalC4FA700(unit)
	}
	return any(unit).(playerBotCreateGo4FA700).PlayerBotCreate4FA700(rt)
}

func TestPlayerBotCreateOriginalNative4FA700(t *testing.T) {
	rows, _ := playerBotCreateLoad4FA700(t)
	rng := playerBotCreateRNGLoad4FA700(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					w := playerBotCreateWorldNew4FA700(t, row, rng)
					cpud, cmud, width := playerBotCreateOriginalControls4FA700(w.Unit, w.PUD)
					if cpud != uintptr(w.Unit.UpdateData) || cmud != uintptr(unsafe.Pointer(w.PUD.Field73)) || width != unsafe.Sizeof(uintptr(0)) {
						t.Fatal("typed C native record controls or bot link width", cpud, cmud, width)
					}
					if got := playerBotCreateCall4FA700(path, w.Unit, w.Runtime); got != row.Output.Result {
						t.Fatalf("raw Create result:got%#x original%#x", got, row.Output.Result)
					}
					w.compare(t, row)
				})
			}
		})
	}
}

func TestPlayerBotCreateOriginalRepresentation4FA700(t *testing.T) {
	mappings := playerBotCreateMappings4FA700(t)
	if len(mappings) == 0 {
		t.Fatal("missing complete original record mapping")
	}
	if reflect.TypeOf(server.MonsterUpdateData{}).Size() < 2200 {
		t.Fatal("native monster record shrank")
	}
}

func TestPlayerBotCreateOriginalRawFaults4FA700(t *testing.T) {
	_, faults := playerBotCreateLoad4FA700(t)
	rng := playerBotCreateRNGLoad4FA700(t)
	if child := os.Getenv("NOX_BOT_CREATE_RAW_FAULT_CHILD"); child != "" {
		parts := strings.SplitN(child, "/", 2)
		for _, row := range faults {
			if row.Input.Name == parts[1] {
				w := playerBotCreateWorldNew4FA700(t, row, rng)
				unit := w.Unit
				if row.Input.Nil == "unit" {
					unit = nil
				}
				defer func() {
					if v := recover(); v != nil {
						w.compare(t, row)
						t.Log("NATIVE PARTIAL RECORDS AND SERVICES MATCH ORIGINAL BEFORE FAULT")
						panic(v)
					}
				}()
				t.Log("CALL RAW BOT CREATE FAULT " + child)
				playerBotCreateCall4FA700(parts[0], unit, w.Runtime)
				t.Fatal("invalid raw Create unexpectedly returned")
			}
		}
		t.Fatal("unknown raw fault input")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"C", "Go"} {
		for _, row := range faults {
			t.Run(path+"/"+row.Input.Name, func(t *testing.T) {
				if row.Output.Fault == nil || len(row.Output.Invalid) == 0 {
					t.Fatal("original raw fault observation")
				}
				child := path + "/" + row.Input.Name
				cmd := exec.Command(exe, "-test.run=^TestPlayerBotCreateOriginalRawFaults4FA700$", "-test.v")
				cmd.Env = append(os.Environ(), "NOX_BOT_CREATE_RAW_FAULT_CHILD="+child)
				raw, err := cmd.CombinedOutput()
				if err == nil || !bytes.Contains(raw, []byte("CALL RAW BOT CREATE FAULT "+child)) || !bytes.Contains(raw, []byte("NATIVE PARTIAL RECORDS AND SERVICES MATCH ORIGINAL BEFORE FAULT")) || !bytes.Contains(raw, []byte("invalid memory address")) {
					t.Fatalf("raw Create fault:%v\n%s", err, raw)
				}
			})
		}
	}
}
