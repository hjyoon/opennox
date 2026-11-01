package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
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
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type heightOriginalInput4D9020 struct {
	Name          string            `json:"name"`
	Recipient     uint32            `json:"recipient"`
	Field5        uint32            `json:"field5"`
	Code          uint32            `json:"net_code"`
	Height        uint32            `json:"height_bits"`
	Velocity      uint32            `json:"velocity_bits"`
	Tail          uint32            `json:"tail_bits"`
	Class         uint32            `json:"unit_class"`
	CodeReturn    uint32            `json:"unit_code_return"`
	SendReturn    uint32            `json:"transport_return"`
	CodeMutations map[string]uint32 `json:"code_mutations"`
	SendMutations map[string]uint32 `json:"send_mutations"`
	Nil           bool              `json:"nil_unit"`
}
type heightOriginalEvent4D9020 struct {
	Service   string            `json:"service"`
	Unit      string            `json:"unit"`
	Before    map[string]uint32 `json:"before"`
	Packet    string            `json:"packet"`
	Recipient uint32            `json:"recipient"`
	Kind      uint32            `json:"kind"`
	Result    uint32            `json:"result"`
}
type heightOriginalRow4D9020 struct {
	Input  heightOriginalInput4D9020 `json:"input"`
	Output struct {
		State     map[string]uint32           `json:"state"`
		Events    []heightOriginalEvent4D9020 `json:"events"`
		Result    uint32                      `json:"result"`
		Heap      bool                        `json:"only_declared_service_heap_writes"`
		CW        bool                        `json:"CW_TOP_preserved"`
		Exception any                         `json:"exception"`
		Invalid   []map[string]any            `json:"invalid"`
	} `json:"output"`
}

func heightOriginalLoad4D9020(t *testing.T) ([]heightOriginalRow4D9020, []heightOriginalRow4D9020) {
	t.Helper()
	file := os.Getenv("NOX_HEIGHT_ORIGINAL_REFERENCE")
	if file == "" {
		file = filepath.Join("..", "server", "testdata", "object_height_original_4d9020.json.gz")
	}
	encoded, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(encoded)) != "62d9899356eba94b9b6f57dbc7097f492ae258bf96de57c91f1010c469429516" {
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
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "0234c77edee395b6245412978d6396c02c19179ac59377177109c6b9fce60855" {
		t.Fatal("decoded original seal")
	}
	var data struct {
		Contexts []heightOriginalRow4D9020 `json:"contexts"`
		Faults   []heightOriginalRow4D9020 `json:"faults"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 370 || len(data.Faults) != 1 {
		t.Fatal("original dimensions")
	}
	for _, row := range append(append([]heightOriginalRow4D9020{}, data.Contexts...), data.Faults...) {
		if !row.Output.Heap || !row.Output.CW {
			t.Fatal("original diagnostics", row.Input.Name)
		}
	}
	return data.Contexts, data.Faults
}

type heightOriginalGuard4D9020 struct {
	Before [32]byte
	Unit   server.Object
	After  [32]byte
}

func TestUnitHeightOriginalNative4D9020(t *testing.T) {
	rows, _ := heightOriginalLoad4D9020(t)
	for _, path := range []string{"Go", "C"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					g, free := alloc.New(heightOriginalGuard4D9020{})
					defer free()
					memory := unsafe.Slice((*byte)(unsafe.Pointer(g)), unsafe.Sizeof(*g))
					for i := range memory {
						memory[i] = 0xa5
					}
					unit := &g.Unit
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
						t.Fatal("native record must exceed4GiB")
					}
					input := row.Input
					unit.ObjClass = object.Class(input.Class)
					unit.Field5 = input.Field5
					unit.NetCode = input.Code
					unit.ZVal = math.Float32frombits(input.Height)
					unit.Field27 = math.Float32frombits(input.Velocity)
					unit.Field29 = input.Tail
					expected := bytes.Clone(memory)
					offset := unsafe.Offsetof(g.Unit)
					offsets := map[string]uintptr{"unit_class": offset + unsafe.Offsetof(unit.ObjClass), "field5": offset + unsafe.Offsetof(unit.Field5), "net_code": offset + unsafe.Offsetof(unit.NetCode), "height_bits": offset + unsafe.Offsetof(unit.ZVal), "velocity_bits": offset + unsafe.Offsetof(unit.Field27), "tail_bits": offset + unsafe.Offsetof(unit.Field29)}
					state := func() map[string]uint32 {
						return map[string]uint32{"unit_class": uint32(unit.ObjClass), "field5": uint32(unit.Field5), "net_code": unit.NetCode, "height_bits": math.Float32bits(unit.ZVal), "velocity_bits": math.Float32bits(unit.Field27), "tail_bits": unit.Field29}
					}
					mutate := func(patch map[string]uint32) {
						for name, value := range patch {
							switch name {
							case "unit_class":
								unit.ObjClass = object.Class(value)
							case "field5":
								unit.Field5 = value
							case "net_code":
								unit.NetCode = value
							case "height_bits":
								unit.ZVal = math.Float32frombits(value)
							case "velocity_bits":
								unit.Field27 = math.Float32frombits(value)
							case "tail_bits":
								unit.Field29 = value
							default:
								t.Fatal("undeclared mutation", name)
							}
							off, ok := offsets[name]
							if !ok {
								t.Fatal("missing native field offset", name)
							}
							for j := uintptr(0); j < 4; j++ {
								expected[off+j] = byte(value >> (8 * j))
							}
						}
					}
					next := 0
					observe := func(service string) heightOriginalEvent4D9020 {
						if next >= len(row.Output.Events) {
							t.Fatalf("extra %s call", service)
						}
						event := row.Output.Events[next]
						next++
						if event.Service != service || !reflect.DeepEqual(state(), event.Before) {
							t.Fatalf("child order/state: %s %#v, want %#v", service, state(), event)
						}
						return event
					}
					runtime := server.UnitHeightRuntime4D9020{
						UnitCode: func(got *server.Object) uint32 {
							event := observe("unit_code")
							if got != unit || event.Unit != "unit" || event.Result != input.CodeReturn {
								t.Fatal("unit-code arguments/return", event)
							}
							mutate(input.CodeMutations)
							return input.CodeReturn
						},
						Submit: func(recipient, kind int32, packet []byte) int32 {
							event := observe("netlist")
							wanted, err := hex.DecodeString(event.Packet)
							if err != nil {
								t.Fatal(err)
							}
							if uint32(recipient) != event.Recipient || uint32(kind) != event.Kind || !bytes.Equal(packet, wanted) || event.Result != input.SendReturn {
								t.Fatalf("netlist args: recipient%08x kind%d packet%x, want %#v", uint32(recipient), kind, packet, event)
							}
							mutate(input.SendMutations)
							return int32(input.SendReturn)
						},
					}
					var result int32
					if path == "Go" {
						result = server.ObjectHeightReportNative4D9020(int32(input.Recipient), unit, runtime)
					} else {
						old := unitHeightRuntimeNative4D9020
						unitHeightRuntimeNative4D9020 = func() server.UnitHeightRuntime4D9020 { return runtime }
						defer func() { unitHeightRuntimeNative4D9020 = old }()
						result = heightOriginalCEntry4D9020(int32(input.Recipient), unit)
					}
					if uint32(result) != row.Output.Result || next != len(row.Output.Events) {
						t.Fatal("full return/child count", uint32(result), next, row.Output.Result, len(row.Output.Events))
					}
					if !reflect.DeepEqual(state(), row.Output.State) {
						t.Fatal("final typed state", state(), row.Output.State)
					}
					if !bytes.Equal(memory, expected) {
						t.Fatal("native object/padding/neighbor guards changed outside declared service writes")
					}
				})
			}
		})
	}
}

func TestUnitHeightOriginalRawFaults4D9020(t *testing.T) {
	_, faults := heightOriginalLoad4D9020(t)
	if !faults[0].Input.Nil || len(faults[0].Output.Invalid) == 0 || faults[0].Output.Exception == nil || len(faults[0].Output.Events) != 0 {
		t.Fatal("original raw nil fault contract")
	}
	if child := os.Getenv("NOX_HEIGHT_RAW_NIL_CHILD"); child != "" {
		t.Log("CALL RAW HEIGHT NIL " + child)
		unitHeightRuntimeNative4D9020 = func() server.UnitHeightRuntime4D9020 { return server.UnitHeightRuntime4D9020{} }
		if child == "Go" {
			server.ObjectHeightReportNative4D9020(3, nil, server.UnitHeightRuntime4D9020{})
		} else if child == "C" {
			heightOriginalCEntry4D9020(3, nil)
		} else {
			t.Fatal("unknown raw path")
		}
		t.Fatal("raw nil unexpectedly returned")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Go", "C"} {
		t.Run(path, func(t *testing.T) {
			cmd := exec.Command(exe, "-test.run=^TestUnitHeightOriginalRawFaults4D9020$", "-test.v")
			cmd.Env = append(os.Environ(), "NOX_HEIGHT_RAW_NIL_CHILD="+path)
			output, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "CALL RAW HEIGHT NIL "+path) || !strings.Contains(string(output), "invalid memory address") {
				t.Fatalf("raw fault boundary: %v\n%s", err, output)
			}
		})
	}
}
