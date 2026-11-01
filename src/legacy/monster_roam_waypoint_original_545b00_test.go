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
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type waypointOriginalRow545B00 struct {
	Input struct {
		Name        string `json:"name"`
		Root        string `json:"root"`
		Index       uint32 `json:"index"`
		Slots       []int  `json:"slots"`
		Waypoint    int    `json:"waypoint"`
		Mask        byte   `json:"mask"`
		Flags       uint32 `json:"flags"`
		Flags2      byte   `json:"flags2"`
		Connections []int  `json:"connections"`
		Count       byte   `json:"count"`
		Seed        uint32 `json:"seed"`
		Head        int8   `json:"head"`
		Code        uint32 `json:"code"`
		NilUpdate   bool   `json:"nil_update"`
		NilCurrent  bool   `json:"nil_current"`
		NilUnit     bool   `json:"nil_unit"`
	} `json:"input"`
	Output struct {
		Index          uint32 `json:"index"`
		Slots          []int  `json:"slots"`
		MUDHex         string `json:"mud_hex"`
		MUDSHA         string `json:"mud_sha256"`
		Result         uint32 `json:"result"`
		ResultWaypoint int    `json:"result_waypoint"`
		RNG            uint32 `json:"rng_index"`
		Calls          []struct {
			Min, Max, Result uint32
			Before           uint32 `json:"seed_before"`
			After            uint32 `json:"seed_after"`
		} `json:"rng_calls"`
		Events  []map[string]any `json:"events"`
		Heap    bool             `json:"whole_other_heap_unchanged"`
		CW      bool             `json:"FPU_CW_and_TOP_preserved"`
		Stack   bool             `json:"caller_only_declared_argument_byte_scratch_changed"`
		Fault   any              `json:"exception"`
		Invalid []map[string]any `json:"invalid"`
	} `json:"output"`
}

func waypointOriginalLoad545B00(t *testing.T) ([]waypointOriginalRow545B00, []waypointOriginalRow545B00) {
	t.Helper()
	encoded, err := os.ReadFile(filepath.Join("..", "server", "testdata", "monster_roam_waypoint_original_545b00.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(encoded)) != "1873692bde6653cfba29094021513f6bcbc59091390d5f48d563af4d6784bded" {
		t.Fatal("original gzip seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "08657d083b0e6d772d7e8075234b2c59487812ea0208ef69becd48ae54a3b86e" {
		t.Fatal("original decoded seal")
	}
	var data struct {
		Contexts []waypointOriginalRow545B00
		Faults   []waypointOriginalRow545B00
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 9220 || len(data.Faults) != 7 {
		t.Fatal("original dimensions")
	}
	return data.Contexts, data.Faults
}

type waypointOriginalNative545B00 interface {
	RoamRecordWaypoint545B00(*server.Waypoint)
	RoamPreviousWaypoint545B60(byte) *server.Waypoint
	RoamNextWaypoint545C60(*server.Waypoint, byte, server.MonsterRoamWaypointRuntime545B00) *server.Waypoint
}
type waypointOriginalValid545B00 interface{ RoamValid547EE0(byte) bool }
type waypointOriginalDeadEnd545B00 interface {
	RoamDeadEnd545BB0(*server.Waypoint, server.MonsterRoamWaypointRuntime545B00) bool
}

type waypointOriginalWorld545B00 struct {
	MUD         *server.MonsterUpdateData
	Unit        *server.Object
	Points      [17]*server.Waypoint
	Current     *server.Waypoint
	Raw, Before []byte
	MUDOffset   int
	Runtime     server.MonsterRoamWaypointRuntime545B00
	Events      []map[string]any
	RNG         uint32
	Calls       int
}

func waypointOriginalWorldNew545B00(t *testing.T, row waypointOriginalRow545B00) *waypointOriginalWorld545B00 {
	t.Helper()
	const size = 32768
	base, free := alloc.Malloc(size)
	t.Cleanup(free)
	w := &waypointOriginalWorld545B00{Raw: unsafe.Slice((*byte)(base), size), RNG: row.Input.Seed, Events: make([]map[string]any, 0)}
	for i := range w.Raw {
		w.Raw[i] = 0xa5
	}
	next := 0
	take := func(n uintptr) unsafe.Pointer {
		next = (next + 63) &^ 63
		next += 32
		start := next
		next += int(n) + 32
		if next > size {
			t.Fatal("guarded native allocation size")
		}
		out := unsafe.Add(base, start)
		if unsafe.Sizeof(out) > 4 && uintptr(out) <= math.MaxUint32 {
			t.Fatal("native record must exceed4GiB")
		}
		return out
	}
	w.MUD = (*server.MonsterUpdateData)(take(unsafe.Sizeof(server.MonsterUpdateData{})))
	w.MUDOffset = int(uintptr(unsafe.Pointer(w.MUD)) - uintptr(base))
	w.Unit = (*server.Object)(take(unsafe.Sizeof(server.Object{})))
	for i := 1; i < len(w.Points); i++ {
		w.Points[i] = (*server.Waypoint)(take(unsafe.Sizeof(server.Waypoint{})))
	}
	w.Current = (*server.Waypoint)(take(unsafe.Sizeof(server.Waypoint{})))
	w.MUD.Field91 = row.Input.Index
	for i := range w.MUD.Waypoints {
		role := 0
		if len(row.Input.Slots) > 0 {
			role = row.Input.Slots[i]
		}
		w.MUD.Waypoints[i] = w.Points[role]
	}
	for i := 1; i < len(w.Points); i++ {
		flags := uint32(1 | i<<8)
		if i%4 == 0 {
			flags = 0
		}
		flags2 := byte(1 << ((i - 1) % 8))
		if i%3 == 0 {
			flags2 |= 64
		}
		w.Points[i].Flags = flags
		w.Points[i].Flags2 = flags2
	}
	if row.Input.Root == "Valid" {
		w.Points[1].Flags = row.Input.Flags
		w.Points[1].Flags2 = row.Input.Flags2
	}
	if row.Input.Root == "Next" || row.Input.Root == "DeadEnd" {
		w.MUD.Field2 = 0xdeadbeef
		w.MUD.AIStackInd = row.Input.Head
		head := &w.MUD.AIStack[row.Input.Head]
		head.Action = 10
		head.Args[0] = uintptr(unsafe.Pointer(w.Current))
		head.Args[2] = uintptr(0xa5a5a500 | uint32(row.Input.Mask))
		w.Current.PointsCnt = row.Input.Count
		for i, role := range row.Input.Connections {
			w.Current.Points[i].Waypoint = w.Points[role]
		}
		w.Unit.UpdateData = unsafe.Pointer(w.MUD)
		w.Unit.NetCode = row.Input.Code
		if row.Input.NilUpdate {
			w.Unit.UpdateData = nil
		}
	}
	w.Before = bytes.Clone(w.Raw)
	w.Runtime = server.MonsterRoamWaypointRuntime545B00{
		Random: func(min, max int) int {
			if w.Calls >= len(row.Output.Calls) {
				t.Fatal("unexpected original RNG call")
			}
			c := row.Output.Calls[w.Calls]
			w.Calls++
			if uint32(min) != c.Min || uint32(max) != c.Max || w.RNG != c.Before {
				t.Fatal("original RNG arguments/index", min, max, w.RNG, c)
			}
			if c.Before < 4096 {
				random := prand.New(int(c.Before))
				value := random.IntClamp(min, max)
				if uint32(value) != c.Result || uint32(random.Index()) != c.After {
					t.Fatal("actual production RNG differs from original", value, random.Index(), c)
				}
			}
			w.RNG = c.After
			return int(int32(c.Result))
		},
		UnitName: func(unit *server.Object) string {
			if unit != w.Unit {
				t.Fatal("name unit")
			}
			w.Events = append(w.Events, map[string]any{"service": "unit-name", "args": []any{"unit"}, "result": "unit-name"})
			return "OracleUnit"
		},
		Frame: func() uint32 { return 100 },
		Debug: func(format string, frame uint32, name string, code int32) {
			w.Events = append(w.Events, map[string]any{"service": "debug-printf", "format": format, "frame": float64(frame), "name": name, "code_word": float64(uint32(code))})
		},
		Pop: func(unit *server.Object) int {
			if unit != w.Unit {
				t.Fatal("pop unit")
			}
			w.Events = append(w.Events, map[string]any{"service": "pop-action", "args": []any{"unit"}, "result": float64(37), "mutation": "none in declared world service"})
			return 37
		},
	}
	return w
}

func (w *waypointOriginalWorld545B00) call(t *testing.T, row waypointOriginalRow545B00, path string) uintptr {
	t.Helper()
	data, point, unit := w.MUD, w.Current, w.Unit
	if row.Input.Root == "Record" || row.Input.Root == "Valid" {
		point = w.Points[row.Input.Waypoint]
	}
	if row.Input.NilUpdate {
		data = nil
	}
	if row.Input.NilCurrent {
		point = nil
	}
	if row.Input.NilUnit {
		unit = nil
	}
	if path == "C" {
		old := monsterRoamWaypointRuntimeOverride545B00
		monsterRoamWaypointRuntimeOverride545B00 = &w.Runtime
		defer func() { monsterRoamWaypointRuntimeOverride545B00 = old }()
		return monsterRoamWaypointOriginalC545B00(row.Input.Root, data, point, unit, row.Input.Mask)
	}
	switch row.Input.Root {
	case "Record":
		any(data).(waypointOriginalNative545B00).RoamRecordWaypoint545B00(point)
		return 0
	case "Previous":
		return uintptr(unsafe.Pointer(any(data).(waypointOriginalNative545B00).RoamPreviousWaypoint545B60(row.Input.Mask)))
	case "Next":
		return uintptr(unsafe.Pointer(any(data).(waypointOriginalNative545B00).RoamNextWaypoint545C60(point, row.Input.Mask, w.Runtime)))
	case "Valid":
		if any(point).(waypointOriginalValid545B00).RoamValid547EE0(row.Input.Mask) {
			return 1
		}
		return 0
	case "DeadEnd":
		if any(unit).(waypointOriginalDeadEnd545B00).RoamDeadEnd545BB0(point, w.Runtime) {
			return 1
		}
		return 0
	}
	t.Fatal("unknown original root")
	return 0
}

func waypointOriginalPutPointer545B00(b []byte, value uintptr) {
	if unsafe.Sizeof(value) == 8 {
		binary.LittleEndian.PutUint64(b, uint64(value))
	} else {
		binary.LittleEndian.PutUint32(b, uint32(value))
	}
}

func (w *waypointOriginalWorld545B00) check(t *testing.T, row waypointOriginalRow545B00) {
	t.Helper()
	if !row.Output.Heap || !row.Output.Stack || !row.Output.CW {
		t.Fatal("original complete memory/stack/FPU observations")
	}
	expected := bytes.Clone(w.Before)
	if row.Input.Root == "Record" || row.Input.Root == "DeadEnd" {
		raw, err := hex.DecodeString(row.Output.MUDHex)
		if err != nil || len(raw) != 2200 || fmt.Sprintf("%x", sha256.Sum256(raw)) != row.Output.MUDSHA {
			t.Fatal("original whole MUD seal")
		}
		history := int(unsafe.Offsetof(w.MUD.Waypoints)) + w.MUDOffset
		for i, role := range row.Output.Slots {
			word := binary.LittleEndian.Uint32(raw[300+4*i:])
			want := uint32(0)
			if role != 0 {
				want = 0x21003000 + uint32(role-1)*0x400
			}
			if word != want {
				t.Fatal("original history role")
			}
			waypointOriginalPutPointer545B00(expected[history+i*int(unsafe.Sizeof(uintptr(0))):], uintptr(unsafe.Pointer(w.Points[role])))
		}
		if binary.LittleEndian.Uint32(raw[364:]) != row.Output.Index {
			t.Fatal("original cursor projection")
		}
		binary.LittleEndian.PutUint32(expected[w.MUDOffset+int(unsafe.Offsetof(w.MUD.Field91)):], row.Output.Index)
		if row.Input.Root == "DeadEnd" {
			binary.LittleEndian.PutUint32(expected[w.MUDOffset+int(unsafe.Offsetof(w.MUD.Field2)):], binary.LittleEndian.Uint32(raw[8:]))
			word := binary.LittleEndian.Uint32(raw[556+24*int(row.Input.Head):])
			role := 0
			if word != 0 {
				if word == 0x21008000 {
					role = 17
				} else {
					role = 1 + int((word-0x21003000)/0x400)
					if word != 0x21003000+uint32(role-1)*0x400 || role < 1 || role > 16 {
						t.Fatal("original action waypoint word", word)
					}
				}
			}
			ptr := w.Current
			if role != 17 {
				ptr = w.Points[role]
			}
			offset := w.MUDOffset + int(unsafe.Offsetof(w.MUD.AIStack)) + int(row.Input.Head)*int(unsafe.Sizeof(server.AIStackItem{})) + int(unsafe.Offsetof(server.AIStackItem{}.Args))
			waypointOriginalPutPointer545B00(expected[offset:], uintptr(unsafe.Pointer(ptr)))
		}
	}
	if !bytes.Equal(w.Raw, expected) {
		for i := range expected {
			if w.Raw[i] != expected[i] {
				t.Fatalf("whole native record/padding/neighbor difference at%d got%02x want%02x", i, w.Raw[i], expected[i])
			}
		}
	}
	if w.Calls != len(row.Output.Calls) || w.RNG != row.Output.RNG {
		t.Fatal("original RNG count/final index", w.Calls, w.RNG, row.Output.Calls, row.Output.RNG)
	}
	if !reflect.DeepEqual(w.Events, row.Output.Events) {
		t.Fatalf("original ordered services got%#v want%#v", w.Events, row.Output.Events)
	}
}

func TestMonsterRoamWaypointOriginal545B00(t *testing.T) {
	rows, _ := waypointOriginalLoad545B00(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					w := waypointOriginalWorldNew545B00(t, row)
					got := w.call(t, row, path)
					want := uintptr(row.Output.Result)
					if row.Input.Root == "Previous" || row.Input.Root == "Next" {
						want = uintptr(unsafe.Pointer(w.Points[row.Output.ResultWaypoint]))
					}
					if got != want {
						t.Fatalf("original result got%#x want%#x", got, want)
					}
					w.check(t, row)
				})
			}
		})
	}
}

func TestMonsterRoamWaypointOriginalFaults545B00(t *testing.T) {
	_, rows := waypointOriginalLoad545B00(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					if row.Output.Fault == nil || len(row.Output.Invalid) == 0 {
						t.Fatal("original raw fault")
					}
					w := waypointOriginalWorldNew545B00(t, row)
					panicked := false
					func() { defer func() { panicked = recover() != nil }(); w.call(t, row, path) }()
					if !panicked {
						t.Fatal("original unguarded nil entry must panic")
					}
					w.check(t, row)
				})
			}
		})
	}
}

func TestMonsterRoamWaypointNativeLayout545B00(t *testing.T) {
	got := monsterRoamWaypointOriginalLayout545B00()
	want := [6]uintptr{unsafe.Offsetof(server.MonsterUpdateData{}.Waypoints), unsafe.Offsetof(server.MonsterUpdateData{}.Field91), unsafe.Sizeof(server.Waypoint{}), unsafe.Offsetof(server.Waypoint{}.Flags), unsafe.Offsetof(server.Waypoint{}.Flags2), unsafe.Offsetof(server.Waypoint{}.PointsCnt)}
	if got != want {
		t.Fatal("independent C/Go native layout", got, want)
	}
}
