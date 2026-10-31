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

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spriteInput518AE0 struct {
	Class         uint32            `json:"object_class"`
	Subclass      uint32            `json:"subclass"`
	Type          uint16            `json:"type_index"`
	Flags         uint32            `json:"flags"`
	Primary       uint32            `json:"primary_code"`
	Secondary     uint32            `json:"secondary_code"`
	Caches        [3]uint32         `json:"caches"`
	Update        bool              `json:"update_present"`
	Link          bool              `json:"link_present"`
	LinkedUpdate  bool              `json:"linked_update_present"`
	Unit          bool              `json:"unit_present"`
	Value         uint32            `json:"value"`
	Recipient     uint32            `json:"recipient"`
	DirectReturn  uint32            `json:"direct_return"`
	QueueReturn   uint32            `json:"queue_return"`
	SpecialReturn uint32            `json:"special_return"`
	Sequence      []json.RawMessage `json:"sequence"`
}
type spriteRow518AE0 struct {
	Name   string            `json:"name"`
	Input  spriteInput518AE0 `json:"input"`
	Output struct {
		Returns []uint32         `json:"returns"`
		Events  []map[string]any `json:"events"`
		Fault   json.RawMessage  `json:"fault"`
		Guards  bool             `json:"heap_global_neighbors_and_FP_guards_preserved"`
	} `json:"output"`
}
type spriteGuard518AE0[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}
type spriteRegion518AE0 struct {
	Pointer unsafe.Pointer
	Size    uintptr
}
type spriteFixture518AE0 struct {
	t                       *testing.T
	unit, linked, recipient *spriteGuard518AE0[server.Object]
	update, linkedUpdate    *spriteGuard518AE0[[64]byte]
	regions                 []spriteRegion518AE0
	before                  [][]byte
	input                   spriteInput518AE0
	runtime                 server.SpriteUpdateNativeRuntime518AE0
	events                  []map[string]any
	call                    int
}

func spriteAllocate518AE0[T comparable](t *testing.T, f *spriteFixture518AE0) *spriteGuard518AE0[T] {
	g, free := alloc.New(spriteGuard518AE0[T]{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(g)) <= math.MaxUint32 {
		t.Fatal("native sprite allocation must exceed4GiB")
	}
	for i := range g.Before {
		g.Before[i] = 0x6d
		g.After[i] = 0xf3
	}
	f.regions = append(f.regions, spriteRegion518AE0{unsafe.Pointer(g), unsafe.Sizeof(*g)})
	return g
}
func spriteCanonical518AE0(v any) any {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	var out any
	if e = json.Unmarshal(b, &out); e != nil {
		panic(e)
	}
	return out
}
func (f *spriteFixture518AE0) event(e map[string]any) {
	e["call"] = f.call
	f.events = append(f.events, e)
}
func spriteNewFixture518AE0(t *testing.T) *spriteFixture518AE0 {
	npcReflectLegacyServer4E17B0(t)
	f := &spriteFixture518AE0{t: t}
	f.unit = spriteAllocate518AE0[server.Object](t, f)
	f.linked = spriteAllocate518AE0[server.Object](t, f)
	f.recipient = spriteAllocate518AE0[server.Object](t, f)
	f.update = spriteAllocate518AE0[[64]byte](t, f)
	f.linkedUpdate = spriteAllocate518AE0[[64]byte](t, f)
	f.runtime = server.SpriteUpdateNativeRuntime518AE0{
		UnitCode: func(unit *server.Object) uint32 {
			if unit != f.unitArg() {
				f.t.Fatal("wrong actual scalar input pointer")
			}
			var label any
			if unit != nil {
				label = "unit"
			}
			f.event(map[string]any{"entry": "0x578ac0", "unit": label})
			result := complexObjectUnitCodeC518960(unit)
			f.event(map[string]any{"child_return": "0x578ac0", "return_word": result})
			return result
		},
		ReadType: func(index int) uint32 {
			if index < 0 || index >= 3 {
				f.t.Fatal("wrong cache index")
			}
			return f.input.Caches[index]
		},
		ElevatorLink: func(unit *server.Object, data unsafe.Pointer) *server.Object {
			if unit != &f.unit.Value || data != unsafe.Pointer(&f.update.Value) {
				f.t.Fatal("shaft resolver lost native object or exact cached update")
			}
			if f.input.Link {
				return &f.linked.Value
			}
			return nil
		},
		Important: func(player int32, packet [4]byte, zero, reliable int32) int32 {
			f.event(map[string]any{"service": "0x4e5390", "recipient_word": uint32(player), "length": len(packet), "zero": zero, "reliable": reliable, "packet": hex.EncodeToString(packet[:])})
			return int32(f.input.DirectReturn)
		},
		Queue: func(player, kind int32, packet [4]byte) int32 {
			f.event(map[string]any{"service": "0x40ebc0", "recipient_word": uint32(player), "kind": kind, "length": len(packet), "packet": hex.EncodeToString(packet[:])})
			return int32(f.input.QueueReturn)
		},
		Special: func(recipient, unit *server.Object) int32 {
			if recipient != &f.recipient.Value || unit != &f.unit.Value {
				f.t.Fatal("special callback lost full native recipient or target")
			}
			f.event(map[string]any{"service": "0x527e50", "recipient": "recipient", "unit": "unit"})
			return int32(f.input.SpecialReturn)
		},
	}
	old := spriteUpdateRuntimeNative518AE0
	spriteUpdateRuntimeNative518AE0 = func() server.SpriteUpdateNativeRuntime518AE0 { return f.runtime }
	t.Cleanup(func() { spriteUpdateRuntimeNative518AE0 = old })
	return f
}
func (f *spriteFixture518AE0) unitArg() *server.Object {
	if f.input.Unit {
		return &f.unit.Value
	}
	return nil
}
func (f *spriteFixture518AE0) reset(input spriteInput518AE0) {
	f.input = input
	f.events = nil
	f.call = 0
	f.unit.Value = server.Object{}
	f.linked.Value = server.Object{}
	f.recipient.Value = server.Object{}
	f.update.Value = [64]byte{}
	f.linkedUpdate.Value = [64]byte{}
}
func (f *spriteFixture518AE0) fields() {
	c := f.input
	u := &f.unit.Value
	u.TypeInd = c.Type
	u.ObjClass = object.Class(c.Class)
	u.ObjSubClass = object.SubClass(c.Subclass)
	u.ObjFlags = object.Flags(c.Flags)
	u.NetCode = c.Primary
	u.Extent = c.Secondary
	u.UpdateData = nil
	if c.Update {
		u.UpdateData = unsafe.Pointer(&f.update.Value)
	}
	for _, offset := range []int{0, 12, 16, 20} {
		binary.LittleEndian.PutUint32(f.update.Value[offset:], c.Value)
	}
	// Native link resolution replaces the obsolete PE32 pointer word. Poison
	// that word while retaining a distinct native linked frame at byte16.
	binary.LittleEndian.PutUint32(f.update.Value[4:], 0xdeadbeef)
	f.linked.Value.UpdateData = nil
	if c.LinkedUpdate {
		f.linked.Value.UpdateData = unsafe.Pointer(&f.linkedUpdate.Value)
	}
	binary.LittleEndian.PutUint32(f.linkedUpdate.Value[16:], c.Value^0xa5)
}
func (f *spriteFixture518AE0) snapshot() {
	f.before = nil
	for _, r := range f.regions {
		f.before = append(f.before, append([]byte(nil), unsafe.Slice((*byte)(r.Pointer), int(r.Size))...))
	}
}
func (f *spriteFixture518AE0) check() {
	for i, r := range f.regions {
		if !bytes.Equal(f.before[i], unsafe.Slice((*byte)(r.Pointer), int(r.Size))) {
			f.t.Fatalf("complete native allocation/neighbor guard changed: region=%d pointer=%p", i, r.Pointer)
		}
	}
}
func (f *spriteFixture518AE0) invoke(native bool) int32 {
	if native {
		return spriteUpdatePublicC518AE0(&f.recipient.Value, int32(f.input.Recipient), f.unitArg())
	}
	return server.SpriteUpdateNative518AE0(&f.recipient.Value, int32(f.input.Recipient), f.unitArg(), f.runtime)
}
func spriteLoad518AE0(t *testing.T) ([]spriteRow518AE0, []spriteRow518AE0) {
	t.Helper()
	path := os.Getenv("NOX_SPRITE_UPDATE_ORIGINAL_FIXTURE")
	if path == "" {
		path = filepath.Join("..", "server", "testdata", "sprite_update_original_518ae0.json.gz")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	b, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != "b5cc4c06911b6f0e096083c81402e3c6ea645efa6931335de78619cbf1e72205" {
		t.Fatalf("sealed unchanged original fixture hash: %s", got)
	}
	var data struct {
		Contexts []spriteRow518AE0 `json:"contexts"`
		Faults   []spriteRow518AE0 `json:"expected_fault_contexts"`
	}
	if e = json.Unmarshal(b, &data); e != nil {
		t.Fatal(e)
	}
	if len(data.Contexts) != 6127 || len(data.Faults) != 7 {
		t.Fatal("incomplete original input inventory")
	}
	return data.Contexts, data.Faults
}
func spriteRunOriginal518AE0(t *testing.T, native bool) {
	rows, faults := spriteLoad518AE0(t)
	f := spriteNewFixture518AE0(t)
	for index, row := range rows {
		t.Run(fmt.Sprintf("%06d/%s", index, row.Name), func(t *testing.T) {
			f.t = t
			f.reset(row.Input)
			var returns []uint32
			steps := append([]json.RawMessage{json.RawMessage(`{}`)}, row.Input.Sequence...)
			for call, changes := range steps {
				f.call = call
				if e := json.Unmarshal(changes, &f.input); e != nil {
					t.Fatal(e)
				}
				f.fields()
				f.snapshot()
				returns = append(returns, uint32(f.invoke(native)))
				f.check()
			}
			if !row.Output.Guards || !reflect.DeepEqual(returns, row.Output.Returns) || !reflect.DeepEqual(spriteCanonical518AE0(f.events), spriteCanonical518AE0(row.Output.Events)) {
				t.Fatalf("native=%t original return/events mismatch\nreturns=%#v want=%#v\nevents=%#v\nwant=%#v", native, returns, row.Output.Returns, f.events, row.Output.Events)
			}
		})
	}
	for index, row := range faults {
		t.Run(fmt.Sprintf("fault/%d/%s", index, row.Name), func(t *testing.T) {
			f.t = t
			f.reset(row.Input)
			f.fields()
			f.snapshot()
			panicked := func() (panicked bool) { defer func() { panicked = recover() != nil }(); f.invoke(native); return false }()
			f.check()
			if !panicked || !reflect.DeepEqual(spriteCanonical518AE0(f.events), spriteCanonical518AE0(row.Output.Events)) {
				t.Fatalf("native=%t expected original invalid-pointer fault with exact preceding scalar events, panicked=%t events=%#v want=%#v", native, panicked, f.events, row.Output.Events)
			}
		})
	}
}
func TestSpriteUpdateOriginalNative518AE0(t *testing.T) { spriteRunOriginal518AE0(t, true) }
func TestSpriteUpdateOriginalGo518AE0(t *testing.T)     { spriteRunOriginal518AE0(t, false) }
