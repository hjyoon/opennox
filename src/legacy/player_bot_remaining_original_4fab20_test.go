package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type playerBotRemainingRow4FAB20 struct {
	Base        playerBotCreateRow4FA700
	Root        string
	Flags, Gate uint32
}

func playerBotRemainingLoad4FAB20(t *testing.T) ([]playerBotRemainingRow4FAB20, []playerBotRemainingRow4FAB20) {
	t.Helper()
	b := playerBotCreateRead4FA700(t, "player_bot_remaining_original_4fab20.json.gz", "279e974a0fde0034d014f7a16538c694b203a638ddd8152adcf0a984e92c3168", "3e49c15803442ac445bc752ea34fddfbffc197aee56ecf2625978f9da7eeb37f")
	var d struct {
		Contexts []json.RawMessage `json:"contexts"`
		Faults   []json.RawMessage `json:"faults"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	parse := func(rows []json.RawMessage) []playerBotRemainingRow4FAB20 {
		out := make([]playerBotRemainingRow4FAB20, 0, len(rows))
		for _, raw := range rows {
			var row playerBotRemainingRow4FAB20
			var extra struct {
				Input struct {
					Root  string `json:"root"`
					Flags uint32 `json:"flags"`
					Gate  uint32 `json:"gate"`
				}
			}
			if err := json.Unmarshal(raw, &row.Base); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &extra); err != nil {
				t.Fatal(err)
			}
			row.Root, row.Flags, row.Gate = extra.Input.Root, extra.Input.Flags, extra.Input.Gate
			out = append(out, row)
		}
		return out
	}
	if len(d.Contexts) != 619 || len(d.Faults) != 7 {
		t.Fatal("remaining original dimensions")
	}
	return parse(d.Contexts), parse(d.Faults)
}

type playerBotRemainingWorld4FAB20 struct {
	*playerBotCreateWorld4FA700
	Runtime server.PlayerBotRuntime4FAB20
}

func playerBotRemainingWorldNew4FAB20(t *testing.T, row playerBotRemainingRow4FAB20, rng map[uint32]playerBotCreateRNG4FA700) *playerBotRemainingWorld4FAB20 {
	t.Helper()
	w := &playerBotRemainingWorld4FAB20{playerBotCreateWorld4FA700: playerBotCreateWorldNew4FA700(t, row.Base, rng)}
	if row.Base.Input.Nil == "bot" {
		w.PUD.Field73 = nil
		binary.LittleEndian.PutUint32(w.Original["pud"][292:], 0)
	}
	if row.Base.Input.Nil == "health" {
		w.Unit.HealthData = nil
		binary.LittleEndian.PutUint32(w.Original["unit"][556:], 0)
	}
	regular := playerBotRegularUpdate4FAB20()
	w.Roles[0x4f8100] = uintptr(regular)
	w.Roles[0x4fab20] = uintptr(w.Unit.Update)
	for name, b := range w.Raw {
		w.Before[name] = bytes.Clone(b)
	}
	role := func(v uint32) string {
		names := map[uint32]string{0: "nil", 0x21001000: "unit", 0x21003000: "pud", 0x21004000: "pud2", 0x21006000: "mud", 0x21008000: "mud2", 0x2100a000: "player", 0x2100c000: "player2", 0x2100e000: "health", 0x2100f000: "settings", 0x21010000: "definition", 0x21001038: "unit.pos", 0x4f8100: "regular-player-update"}
		if name, ok := names[v]; ok {
			return name
		}
		return fmt.Sprintf("0x%x", v)
	}
	oldPointer := func(p unsafe.Pointer) uint32 {
		for old, native := range w.Roles {
			if native == uintptr(p) {
				return old
			}
		}
		t.Fatalf("unknown native service pointer %p", p)
		return 0
	}
	emit := func(name string, args []any, result uint32) map[string]any {
		roles := make([]any, len(args))
		for i, arg := range args {
			if v, ok := arg.(uint32); ok {
				roles[i] = role(v)
			} else {
				roles[i] = arg
			}
		}
		event := map[string]any{"service": name, "args": args, "roles": roles, "result": result, "result_role": role(result)}
		w.Events = append(w.Events, event)
		return event
	}
	var local *types.Pointf
	w.Runtime = server.PlayerBotRuntime4FAB20{
		Create: w.playerBotCreateWorld4FA700.Runtime, Frame: func() uint32 { return 100 },
		LoadSettings: func() *server.Settings {
			emit("settings", []any{}, 0x2100f000)
			if row.Base.Input.World == "settings-replace-pud" {
				w.Unit.UpdateData = unsafe.Pointer(w.PUD2)
			}
			if row.Base.Input.World == "settings-clear-bot" {
				w.PUD.Field73 = nil
			}
			return w.Settings
		}, NetworkMode: func() uint32 { return row.Gate },
		DefaultItems: func(unit *server.Object, first, second int32) {
			if unit != w.Unit || first != 1 || second != 0 {
				t.Fatal("original default items arguments")
			}
			emit("default-items", []any{oldPointer(unsafe.Pointer(unit)), uint32(first), uint32(second)}, 0)
			w.Health.Cur = 20
		},
		RespawnCorpse: func(pos *types.Pointf, direction int16) {
			if pos != &w.Unit.PosVec || direction != int16(row.Base.Input.Direction) {
				t.Fatal("original corpse live position and signed direction")
			}
			emit("respawn-impl", []any{uint32(0x21001038), uint32(int32(direction))}, 0)
		},
		FindStart: func(out *types.Pointf, unit *server.Object) {
			if out == nil || out == &unit.PosVec || unit != w.Unit {
				t.Fatal("original independent local point")
			}
			local = out
			emit("find-start", []any{"local-position", oldPointer(unsafe.Pointer(unit))}, 0)
			out.X = math.Float32frombits(0x41200000)
			out.Y = math.Float32frombits(0xc1a00000)
		},
		Move: func(unit *server.Object, pos *types.Pointf) {
			if unit != w.Unit || pos != local {
				t.Fatal("original moved local point identity")
			}
			event := emit("move", []any{oldPointer(unsafe.Pointer(unit)), "local-position"}, 0)
			event["position_words"] = []uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}
			unit.PosVec = *pos
		},
		Audio: func(id int32, unit *server.Object, first, second int32) {
			if id != 148 || unit != w.Unit || first != 0 || second != 0 {
				t.Fatal("original respawn audio arguments")
			}
			emit("audio", []any{uint32(id), oldPointer(unsafe.Pointer(unit)), uint32(first), uint32(second)}, 0)
		},
		GameFlag: func(mask uint32) bool {
			if mask != 0x2000 {
				t.Fatal("original respawn flag mask", mask)
			}
			return row.Flags&mask != 0
		},
		Buff: func(unit *server.Object, id int32, duration int16, power int8) {
			if unit != w.Unit || id != 23 || power != 5 {
				t.Fatal("original respawn buff arguments")
			}
			event := emit("buff", []any{oldPointer(unsafe.Pointer(unit)), uint32(id), uint32(int32(duration)), uint32(int32(power))}, 0)
			event["semantic_duration_i16"] = int32(duration)
		},
		UpdateMonster: func(unit *server.Object) {
			if unit != w.Unit || !unit.ObjClass.Has(object.ClassMonster) || unit.ObjSubClass != 16 {
				t.Fatal("original transient monster state")
			}
			monster := (*server.MonsterUpdateData)(unit.UpdateData)
			event := emit("monster-update", []any{oldPointer(unsafe.Pointer(unit))}, 0)
			event["live_mud"] = role(oldPointer(unit.UpdateData))
			switch row.Base.Input.World {
			case "mutate", "monster-replace-record", "monster-replace-player":
				monster.Field120_1 = 91
				monster.AIStackInd = 23
				monster.AIStack[23].Action = 22
				unit.PosVec.X = math.Float32frombits(0x7fc54321)
				unit.PosVec.Y = math.Float32frombits(0x80000000)
			}
			if row.Base.Input.World == "monster-replace-record" {
				unit.UpdateData = unsafe.Pointer(w.MUD2)
			}
			if row.Base.Input.World == "monster-replace-player" {
				w.PUD.Player = w.Player2
			}
		}, RegularPlayerUpdate: regular,
	}
	return w
}

func (w *playerBotRemainingWorld4FAB20) compare(t *testing.T, row playerBotRemainingRow4FAB20) {
	t.Helper()
	if !row.Base.Output.Heap || !row.Base.Output.Stack || !row.Base.Output.CW {
		t.Fatal("original whole heap/stack/FPU observations")
	}
	expected := make(map[string][]byte)
	for name, b := range w.Before {
		expected[name] = bytes.Clone(b)
	}
	for name, record := range row.Base.Output.Records {
		after, err := hex.DecodeString(record.Hex)
		if err != nil || len(after) != record.Bytes {
			t.Fatal("complete original record", name)
		}
		before := w.Original[name]
		allowed := bytes.Clone(before)
		scalar := func(old int, native uintptr, n int) {
			copy(allowed[old:old+n], after[old:old+n])
			copy(expected[name][32+native:], after[old:old+n])
		}
		pointer := func(old int, native uintptr) {
			value := binary.LittleEndian.Uint32(after[old:])
			p, ok := w.Roles[value]
			if !ok {
				t.Fatalf("unclassified original pointer %s[%d]=%#x", name, old, value)
			}
			copy(allowed[old:old+4], after[old:old+4])
			playerBotCreatePutNativeWord4FA700(expected[name][32:], native, p)
		}
		switch name {
		case "unit":
			scalar(8, unsafe.Offsetof(server.Object{}.ObjClass), 4)
			scalar(12, unsafe.Offsetof(server.Object{}.ObjSubClass), 4)
			scalar(56, unsafe.Offsetof(server.Object{}.PosVec), 8)
			pointer(744, unsafe.Offsetof(server.Object{}.Update))
			pointer(748, unsafe.Offsetof(server.Object{}.UpdateData))
		case "pud", "pud2":
			scalar(88, unsafe.Offsetof(server.PlayerUpdateData{}.State), 1)
			scalar(236, unsafe.Offsetof(server.PlayerUpdateData{}.Field59_0), 1)
			pointer(276, unsafe.Offsetof(server.PlayerUpdateData{}.Player))
			pointer(292, unsafe.Offsetof(server.PlayerUpdateData{}.Field73))
		case "mud", "mud2":
			if binary.LittleEndian.Uint32(after) == 0xdeadface {
				copy(expected[name][32:], playerBotCreateExpandOriginal4FA700(t, after, w.Roles))
				copy(allowed, after)
			} else {
				scalar(1440, unsafe.Offsetof(server.MonsterUpdateData{}.StatusFlags), 4)
				scalar(481, unsafe.Offsetof(server.MonsterUpdateData{}.Field120_1), 1)
				scalar(544, unsafe.Offsetof(server.MonsterUpdateData{}.AIStackInd), 1)
				for i := 0; i < 24; i++ {
					scalar(552+24*i, unsafe.Offsetof(server.MonsterUpdateData{}.AIStack)+uintptr(i)*unsafe.Sizeof(server.AIStackItem{}), 4)
				}
			}
		case "player", "player2":
			scalar(3632, unsafe.Offsetof(server.Player{}.Pos3632Vec), 8)
		case "health":
			scalar(0, unsafe.Offsetof(server.HealthData{}.Cur), 2)
		}
		if !bytes.Equal(allowed, after) {
			t.Fatal("unclassified original changed bytes", name)
		}
	}
	for name, actual := range w.Raw {
		if !bytes.Equal(expected[name], actual) {
			for i, v := range actual {
				if v != expected[name][i] {
					t.Fatalf("whole native record/guard %s byte%d:got%02x original%02x", name, i, v, expected[name][i])
				}
			}
		}
	}
	want := make([]map[string]any, 0, len(row.Base.Output.Events))
	for _, event := range row.Base.Output.Events {
		raw, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var copyEvent map[string]any
		if err = json.Unmarshal(raw, &copyEvent); err != nil {
			t.Fatal(err)
		}
		if copyEvent["service"] == "NPC-name" {
			copyEvent["args"] = []any{"NPC"}
			copyEvent["roles"] = []any{"NPC"}
		}
		if copyEvent["service"] == "buff" {
			// Original raw caller word is retained in the sealed reference. Its
			// declared int16 callee argument observes only this signed low word.
			semantic, ok := copyEvent["semantic_duration_i16"].(float64)
			if !ok {
				t.Fatal("original int16 buff observation")
			}
			word := uint32(int32(semantic))
			copyEvent["args"].([]any)[2] = word
			role := fmt.Sprintf("0x%x", word)
			if word == 0 {
				role = "nil"
			}
			copyEvent["roles"].([]any)[2] = role
		}
		want = append(want, copyEvent)
	}
	actual, _ := json.Marshal(w.Events)
	original, _ := json.Marshal(want)
	if !bytes.Equal(actual, original) {
		t.Fatalf("ordered world services\ngot%s\noriginal%s", actual, original)
	}
	if w.RNG != row.Base.Output.RNG {
		t.Fatal("original final RNG index", w.RNG, row.Base.Output.RNG)
	}
}

type playerBotRemainingGo4FAB20 interface {
	PlayerBotUpdate4FAB20(server.PlayerBotRuntime4FAB20) uint32
	PlayerBotRespawn4FAC70(server.PlayerBotRuntime4FAB20) uint32
}

func playerBotRemainingCall4FAB20(path, root string, unit *server.Object, rt server.PlayerBotRuntime4FAB20) uint32 {
	if path == "C" {
		old := playerBotRuntimeOverride4FAB20
		playerBotRuntimeOverride4FAB20 = &rt
		defer func() { playerBotRuntimeOverride4FAB20 = old }()
		which := 0
		if root == "Respawn" {
			which = 1
		}
		return playerBotRemainingOriginalC4FAB20(which, unit)
	}
	native := any(unit).(playerBotRemainingGo4FAB20)
	if root == "Respawn" {
		return native.PlayerBotRespawn4FAC70(rt)
	}
	return native.PlayerBotUpdate4FAB20(rt)
}

func TestPlayerBotRemainingOriginalNative4FAB20(t *testing.T) {
	rows, _ := playerBotRemainingLoad4FAB20(t)
	rng := playerBotCreateRNGLoad4FA700(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Root+"/"+row.Base.Input.Name, func(t *testing.T) {
					w := playerBotRemainingWorldNew4FAB20(t, row, rng)
					if got := playerBotRemainingCall4FAB20(path, row.Root, w.Unit, w.Runtime); got != row.Base.Output.Result {
						t.Fatalf("raw remaining result:got%#x original%#x", got, row.Base.Output.Result)
					}
					w.compare(t, row)
				})
			}
		})
	}
}

func TestPlayerBotRemainingOriginalRawFaults4FAB20(t *testing.T) {
	_, faults := playerBotRemainingLoad4FAB20(t)
	rng := playerBotCreateRNGLoad4FA700(t)
	if child := os.Getenv("NOX_BOT_REMAINING_RAW_FAULT_CHILD"); child != "" {
		parts := strings.SplitN(child, "/", 3)
		for _, row := range faults {
			if row.Root == parts[1] && row.Base.Input.Name == parts[2] {
				w := playerBotRemainingWorldNew4FAB20(t, row, rng)
				unit := w.Unit
				if row.Base.Input.Nil == "unit" {
					unit = nil
				}
				defer func() {
					if v := recover(); v != nil {
						w.compare(t, row)
						t.Log("NATIVE REMAINING PARTIAL RECORDS AND SERVICES MATCH ORIGINAL BEFORE FAULT")
						panic(v)
					}
				}()
				t.Log("CALL RAW BOT REMAINING FAULT " + child)
				playerBotRemainingCall4FAB20(parts[0], row.Root, unit, w.Runtime)
				t.Fatal("invalid raw entry unexpectedly returned")
			}
		}
		t.Fatal("unknown remaining raw fault")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"C", "Go"} {
		for _, row := range faults {
			t.Run(path+"/"+row.Root+"/"+row.Base.Input.Name, func(t *testing.T) {
				if row.Base.Output.Fault == nil || len(row.Base.Output.Invalid) == 0 {
					t.Fatal("original raw remaining fault")
				}
				child := path + "/" + row.Root + "/" + row.Base.Input.Name
				cmd := exec.Command(exe, "-test.run=^TestPlayerBotRemainingOriginalRawFaults4FAB20$", "-test.v")
				cmd.Env = append(os.Environ(), "NOX_BOT_REMAINING_RAW_FAULT_CHILD="+child)
				raw, err := cmd.CombinedOutput()
				if err == nil || !bytes.Contains(raw, []byte("CALL RAW BOT REMAINING FAULT "+child)) || !bytes.Contains(raw, []byte("NATIVE REMAINING PARTIAL RECORDS AND SERVICES MATCH ORIGINAL BEFORE FAULT")) || !bytes.Contains(raw, []byte("invalid memory address")) {
					t.Fatalf("raw remaining fault:%v\n%s", err, raw)
				}
			})
		}
	}
}
