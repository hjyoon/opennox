package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
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

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerBotActionWord4FABC0 struct {
	Index    int8
	Original int
	Native   uintptr
	Pointer  bool
}

type playerBotActionRow4FABC0 struct {
	Input struct {
		Name   string `json:"name"`
		Index  int8   `json:"index"`
		Action uint32 `json:"action"`
		Status uint32 `json:"status"`
		Nil    string `json:"nil"`
	} `json:"input"`
	Output struct {
		Result    byte             `json:"result"`
		Heap      bool             `json:"all_other_heap_bytes_unchanged"`
		WholeHeap bool             `json:"whole_heap_unchanged"`
		Stack     bool             `json:"caller_stack_unchanged"`
		CW        bool             `json:"CW_TOP_preserved"`
		Fault     any              `json:"exception"`
		Invalid   []map[string]any `json:"invalid"`
	} `json:"output"`
}

func playerBotActionLoad4FABC0(t *testing.T) ([]playerBotActionRow4FABC0, []playerBotActionRow4FABC0) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "server", "testdata", "player_bot_action_original_4fabc0.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "eb376a3965001230049ee05cb3d9be6c1d75d890759740a588d5ce653ebbe63b" {
		t.Fatal("original gzip seal")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	raw, err = io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "98cd36ac315e17fc50689983a5bbebb6d8e137a1d396412dfc14e73bbc2999bc" {
		t.Fatal("original decoded seal")
	}
	var data struct {
		Contexts []playerBotActionRow4FABC0 `json:"contexts"`
		Faults   []playerBotActionRow4FABC0 `json:"faults"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Contexts) != 18768 || len(data.Faults) != 3 {
		t.Fatal("original input dimensions")
	}
	return data.Contexts, data.Faults
}

type playerBotActionGo4FABC0 interface{ PlayerBotActionToState4FABC0() byte }

func playerBotActionCall4FABC0(path string, u *server.Object) byte {
	if path == "C" {
		return playerBotActionOriginalC4FABC0(u)
	}
	return any(u).(playerBotActionGo4FABC0).PlayerBotActionToState4FABC0()
}

type playerBotActionWorld4FABC0 struct {
	Unit    *server.Object
	PUD     *server.PlayerUpdateData
	MUD     *server.MonsterUpdateData
	Records [][]byte
	Before  [][]byte
	Free    []func()
}

func playerBotActionNew4FABC0(t *testing.T, row playerBotActionRow4FABC0) *playerBotActionWorld4FABC0 {
	t.Helper()
	w := new(playerBotActionWorld4FABC0)
	take := func(size uintptr) unsafe.Pointer {
		p, free := alloc.MallocHigh(size + 64)
		w.Free = append(w.Free, free)
		raw := unsafe.Slice((*byte)(p), size+64)
		for i := range raw {
			raw[i] = 0xa5
		}
		w.Records = append(w.Records, raw)
		return unsafe.Add(p, 32)
	}
	w.Unit = (*server.Object)(take(unsafe.Sizeof(server.Object{})))
	w.PUD = (*server.PlayerUpdateData)(take(unsafe.Sizeof(server.PlayerUpdateData{})))
	w.MUD = (*server.MonsterUpdateData)(take(unsafe.Sizeof(server.MonsterUpdateData{})))
	for _, p := range []unsafe.Pointer{unsafe.Pointer(w.Unit), unsafe.Pointer(w.PUD), unsafe.Pointer(w.MUD)} {
		if unsafe.Sizeof(p) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native records must exceed4GiB")
		}
	}
	w.Unit.ObjClass = object.Class(0x81220004)
	w.Unit.UpdateData = unsafe.Pointer(w.PUD)
	w.PUD.Field73 = w.MUD
	w.MUD.AIStackInd = row.Input.Index
	w.MUD.StatusFlags = object.MonsterStatus(row.Input.Status)
	if row.Input.Index != -1 {
		found := false
		for _, m := range playerBotActionProjection4FABC0() {
			if m.Index != row.Input.Index {
				continue
			}
			found = true
			addr := unsafe.Add(unsafe.Pointer(w.MUD), m.Native)
			if m.Pointer {
				*(*uintptr)(addr) = uintptr(row.Input.Action)
			} else {
				*(*uint32)(addr) = row.Input.Action
			}
		}
		if !found {
			t.Fatal("missing independent native field projection", row.Input.Index)
		}
	}
	if row.Input.Nil == "pud" {
		w.Unit.UpdateData = nil
	}
	if row.Input.Nil == "bot" {
		w.PUD.Field73 = nil
	}
	for _, raw := range w.Records {
		w.Before = append(w.Before, bytes.Clone(raw))
	}
	t.Cleanup(func() {
		for i := len(w.Free) - 1; i >= 0; i-- {
			w.Free[i]()
		}
	})
	return w
}

func (w *playerBotActionWorld4FABC0) unchanged(t *testing.T) {
	t.Helper()
	for i, raw := range w.Records {
		if !bytes.Equal(raw, w.Before[i]) {
			t.Fatalf("native record/neighbor guard %d changed", i)
		}
	}
}

func TestPlayerBotActionOriginalNative4FABC0(t *testing.T) {
	rows, _ := playerBotActionLoad4FABC0(t)
	for _, path := range []string{"C", "Go"} {
		t.Run(path, func(t *testing.T) {
			for _, row := range rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					if !row.Output.Heap || !row.Output.WholeHeap || !row.Output.Stack || !row.Output.CW || row.Output.Fault != nil {
						t.Fatal("original read-only observations")
					}
					w := playerBotActionNew4FABC0(t, row)
					if got := playerBotActionCall4FABC0(path, w.Unit); got != row.Output.Result {
						t.Fatalf("got%d original%d", got, row.Output.Result)
					}
					w.unchanged(t)
				})
			}
		})
	}
}

func TestPlayerBotActionOriginalLayout4FABC0(t *testing.T) {
	for _, m := range playerBotActionProjection4FABC0() {
		if m.Original != 24*(int(m.Index)+23) {
			t.Fatal("original offset", m)
		}
		width := uintptr(4)
		if m.Pointer {
			width = unsafe.Sizeof(uintptr(0))
		}
		if m.Native+width > unsafe.Sizeof(server.MonsterUpdateData{}) {
			t.Fatal("native field outside record", m)
		}
		if unsafe.Sizeof(uintptr(0)) == 4 && int(m.Native) != m.Original {
			t.Fatal("PE32 field projection differs from original offset", m)
		}
	}
}

func TestPlayerBotActionPointerWidth4FABC0(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		t.Skip("native pointer expansion requires64bits")
	}
	// Every low word is a recognized action. Its high word makes the complete
	// pointer-sized value an original-default action, rather than that low word.
	for _, path := range []string{"C", "Go"} {
		for _, m := range playerBotActionProjection4FABC0() {
			if !m.Pointer {
				continue
			}
			for _, low := range []uint32{7, 16, 18, 21, 22, 24, 30, 31} {
				t.Run(fmt.Sprintf("%s/i%d/low%d", path, m.Index, low), func(t *testing.T) {
					row := playerBotActionRow4FABC0{}
					row.Input.Index = m.Index
					row.Input.Status = 0x4000
					w := playerBotActionNew4FABC0(t, row)
					*(*uintptr)(unsafe.Add(unsafe.Pointer(w.MUD), m.Native)) = uintptr(uint64(1)<<32 | uint64(low))
					w.Before[2] = bytes.Clone(w.Records[2])
					if got := playerBotActionCall4FABC0(path, w.Unit); got != 13 {
						t.Fatal("pointer word truncated", got)
					}
					w.unchanged(t)
				})
			}
		}
	}
}

func TestPlayerBotActionOriginalRawFaults4FABC0(t *testing.T) {
	_, faults := playerBotActionLoad4FABC0(t)
	if child := os.Getenv("NOX_BOT_ACTION_RAW_FAULT_CHILD"); child != "" {
		parts := strings.SplitN(child, "/", 2)
		for _, row := range faults {
			if row.Input.Name == parts[1] {
				w := playerBotActionNew4FABC0(t, row)
				u := w.Unit
				if row.Input.Nil == "unit" {
					u = nil
				}
				defer func() {
					if v := recover(); v != nil {
						w.unchanged(t)
						t.Log("NATIVE RECORDS UNCHANGED BEFORE FAULT")
						panic(v)
					}
				}()
				t.Log("CALL RAW BOT ACTION FAULT " + child)
				playerBotActionCall4FABC0(parts[0], u)
				t.Fatal("raw invalid input unexpectedly returned")
			}
		}
		t.Fatal("unknown fault child")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"C", "Go"} {
		for _, row := range faults {
			t.Run(path+"/"+row.Input.Name, func(t *testing.T) {
				if row.Output.Fault == nil || len(row.Output.Invalid) == 0 || !row.Output.Heap {
					t.Fatal("original fault observation")
				}
				child := path + "/" + row.Input.Name
				cmd := exec.Command(exe, "-test.run=^TestPlayerBotActionOriginalRawFaults4FABC0$", "-test.v")
				cmd.Env = append(os.Environ(), "NOX_BOT_ACTION_RAW_FAULT_CHILD="+child)
				raw, err := cmd.CombinedOutput()
				if err == nil || !bytes.Contains(raw, []byte("CALL RAW BOT ACTION FAULT "+child)) || !bytes.Contains(raw, []byte("NATIVE RECORDS UNCHANGED BEFORE FAULT")) || !bytes.Contains(raw, []byte("invalid memory address")) {
					t.Fatalf("raw fault:%v\n%s", err, raw)
				}
			})
		}
	}
}
