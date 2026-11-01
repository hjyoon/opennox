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
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
)

type timerPercentageRow486640 struct {
	Input struct {
		Name    string `json:"name"`
		Current uint32 `json:"current"`
		Rate    uint32 `json:"rate"`
		Nil     bool   `json:"nil_group"`
	} `json:"input"`
	Output struct {
		Result    uint32           `json:"result"`
		Current   uint32           `json:"current"`
		Record    string           `json:"whole_group_hex"`
		Heap      bool             `json:"whole_heap_unchanged"`
		Stack     bool             `json:"whole_caller_stack_unchanged"`
		CW        bool             `json:"CW_TOP_preserved"`
		Invalid   []map[string]any `json:"invalid"`
		Exception any              `json:"exception"`
	} `json:"output"`
}

func timerPercentageLoad486640(t *testing.T) ([]timerPercentageRow486640, []timerPercentageRow486640) {
	t.Helper()
	file := os.Getenv("NOX_TIMER_PERCENT_ORIGINAL_REFERENCE")
	if file == "" {
		file = filepath.Join("..", "server", "testdata", "timer_percentage_original_486640.json.gz")
	}
	encoded, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(encoded)) != "2b19bbbe0fef266d762ca90a7280d5e0f60e7bc599d6673da4d69c045ddf5ddf" {
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
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "251a9e349276c71661bc4ae49bc41f5543a17198913740f9f8903c93c03ca513" {
		t.Fatal("decoded original seal")
	}
	var data struct {
		Contexts []timerPercentageRow486640 `json:"contexts"`
		Faults   []timerPercentageRow486640 `json:"faults"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 576 || len(data.Faults) != 1 {
		t.Fatal("original dimensions")
	}
	for _, row := range append(append([]timerPercentageRow486640{}, data.Contexts...), data.Faults...) {
		record, err := hex.DecodeString(row.Output.Record)
		if err != nil || len(record) != 96 || !row.Output.Heap || !row.Output.Stack || !row.Output.CW {
			t.Fatal("original read-only/control diagnostics", row.Input.Name)
		}
	}
	return data.Contexts, data.Faults
}
func timerPercentageCall486640(path string, group *timer.TimerGroup, rate int32) int32 {
	switch path {
	case "Go":
		return timer.ScalePercentageNative486640(group, rate)
	case "C":
		return timerPercentageOriginalC486640(group, rate)
	case "LegacyGo":
		return int32(Sub_486640(unsafe.Pointer(group), int(rate)))
	default:
		panic("unknown timer percentage path")
	}
}

type timerPercentageGuard486640 struct {
	Before [32]byte
	Group  timer.TimerGroup
	After  [32]byte
}

func TestTimerPercentageOriginalNative486640(t *testing.T) {
	rows, _ := timerPercentageLoad486640(t)
	for _, path := range []string{"Go", "C", "LegacyGo"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					g, free := alloc.New(timerPercentageGuard486640{})
					defer free()
					memory := unsafe.Slice((*byte)(unsafe.Pointer(g)), unsafe.Sizeof(*g))
					for i := range memory {
						memory[i] = 0xa5
					}
					group := &g.Group
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(group)) <= math.MaxUint32 {
						t.Fatal("native group must exceed4GiB")
					}
					if unsafe.Sizeof(*group) != 96 || unsafe.Offsetof(group.Timers)+unsafe.Sizeof(group.Timers[0])+unsafe.Offsetof(group.Timers[1].Current) != 36 {
						t.Fatal("native typed group layout")
					}
					group.Timers[1].Current = row.Input.Current
					before := bytes.Clone(memory)
					wantRecord, err := hex.DecodeString(row.Output.Record)
					if err != nil {
						t.Fatal(err)
					}
					record := unsafe.Slice((*byte)(unsafe.Pointer(group)), unsafe.Sizeof(*group))
					if !bytes.Equal(record, wantRecord) {
						t.Fatal("original complete group input record")
					}
					got := timerPercentageCall486640(path, group, int32(row.Input.Rate))
					if uint32(got) != row.Output.Result {
						t.Fatalf("return%08x, original%08x", uint32(got), row.Output.Result)
					}
					if !bytes.Equal(record, wantRecord) || !bytes.Equal(memory, before) {
						t.Fatal("complete group/neighbor guards modified")
					}
				})
			}
		})
	}
}
func TestTimerPercentageOriginalRawFaults486640(t *testing.T) {
	_, faults := timerPercentageLoad486640(t)
	if !faults[0].Input.Nil || faults[0].Output.Exception == nil || len(faults[0].Output.Invalid) == 0 {
		t.Fatal("original raw nil fault contract")
	}
	if path := os.Getenv("NOX_TIMER_PERCENT_RAW_NIL_CHILD"); path != "" {
		t.Log("CALL RAW TIMER PERCENT NIL " + path)
		timerPercentageCall486640(path, nil, 44100)
		t.Fatal("raw nil unexpectedly returned")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"Go", "C", "LegacyGo"} {
		t.Run(path, func(t *testing.T) {
			cmd := exec.Command(exe, "-test.run=^TestTimerPercentageOriginalRawFaults486640$", "-test.v")
			cmd.Env = append(os.Environ(), "NOX_TIMER_PERCENT_RAW_NIL_CHILD="+path)
			output, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "CALL RAW TIMER PERCENT NIL "+path) || !strings.Contains(string(output), "invalid memory address") {
				t.Fatalf("nil boundary: %v\n%s", err, output)
			}
		})
	}
}
