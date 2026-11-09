package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func memDumpTestTable413E30(t *testing.T) []byte {
	t.Helper()
	b := memmap.Slice(0x5D4594, 252284)[:memDumpRowSize413E30*memDumpRowCount413E30]
	old := bytes.Clone(b)
	h := memDumpHead413E30()
	oldHead := *h
	t.Cleanup(func() { copy(b, old); *h = oldHead })
	for i := range b {
		b[i] = 0xA5
	}
	return b
}

func TestGAME1MemDump413E30NativeCNodeLayout(t *testing.T) {
	var n memDumpNode413E30
	want := [4]uintptr{unsafe.Sizeof(n), unsafe.Offsetof(n.Next), unsafe.Offsetof(n.Count), unsafe.Offsetof(n.Name)}
	if got := memDumpFixtureLayout413E30(); got != want {
		t.Fatalf("C native layout %v Go %v", got, want)
	}
	if unsafe.Sizeof(n.Count) != 4 || len(n.Name) != 80 {
		t.Fatal("scalar or name widened with pointers")
	}
}

func TestGAME1MemDump413E30ActualCEntryHighNativeLinks(t *testing.T) {
	table := memDumpTestTable413E30(t)
	var nodes []*memDumpNode413E30
	for i := 0; i < 4; i++ {
		p := memDumpFixtureNew413E30()
		if p == nil {
			t.Fatal("C node allocation")
		}
		t.Cleanup(func() { memDumpFixtureFree413E30(p) })
		nodes = append(nodes, (*memDumpNode413E30)(p))
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= uintptr(^uint32(0)) {
			t.Fatal("C node did not exercise high pointer bits")
		}
	}
	for i, x := range []struct {
		name  string
		count uint32
	}{{"alpha", 4}, {"beta", 10}, {"alpha", 3}, {"gamma", 1}} {
		copy(nodes[i].Name[:], x.name)
		nodes[i].Count = x.count
		if i+1 < len(nodes) {
			nodes[i].Next = nodes[i+1]
		}
	}
	*memDumpHead413E30() = nodes[0]
	memDumpFixturePublic413E30()
	for i, x := range []struct {
		name  string
		count uint32
	}{{"gamma", 1}, {"alpha", 7}, {"beta", 10}} {
		r := table[i*84 : (i+1)*84]
		if got := string(memDumpName413E30(r[:80])); got != x.name || binary.LittleEndian.Uint32(r[80:]) != x.count {
			t.Fatalf("row%d=%q/%d", i, got, binary.LittleEndian.Uint32(r[80:]))
		}
	}
	for i := 3; i < 1024; i++ {
		r := table[i*84 : (i+1)*84]
		if binary.LittleEndian.Uint32(r[80:]) != 0 || !bytes.Equal(r[:80], bytes.Repeat([]byte{0xA5}, 80)) {
			t.Fatalf("unused record%d changed", i)
		}
	}
	*memDumpHead413E30() = nil
}

func TestGAME1MemDump413E30WrappingCountsAndGoOwners(t *testing.T) {
	table := memDumpTestTable413E30(t)
	first, second := new(memDumpNode413E30), new(memDumpNode413E30)
	copy(first.Name[:], "wrap")
	copy(second.Name[:], "wrap")
	first.Count, second.Count, first.Next = ^uint32(1), 9, second
	*memDumpHead413E30() = first
	for i := 0; i < 3; i++ {
		runtime.GC()
	}
	memDumpFixturePublic413E30()
	if got := binary.LittleEndian.Uint32(table[80:]); got != 7 {
		t.Fatalf("DWORD addition=%d", got)
	}
	runtime.KeepAlive(first)
}

func TestGAME1MemDump413E30PackedHeadIsNotNativeAddress(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 4 {
		return
	}
	table := memDumpTestTable413E30(t)
	packed := memmap.PtrUint32(0x5D4594, 338304)
	old := *packed
	t.Cleanup(func() { *packed = old })
	*packed = 0xF1234567
	*memDumpHead413E30() = nil
	memDumpFixturePublic413E30()
	if *packed != 0xF1234567 || binary.LittleEndian.Uint32(table[80:]) != 0 {
		t.Fatal("packed scalar interpreted or overwritten as native pointer")
	}
}

func TestGAME1MemDump413E30InvalidListPoliciesBoundOutput(t *testing.T) {
	for _, kind := range []string{"cycle", "unterminated", "capacity"} {
		t.Run(kind, func(t *testing.T) {
			const capacity = 3
			storage := bytes.Repeat([]byte{0x5A}, capacity*84+16)
			table := storage[:capacity*84]
			var nodes []*memDumpNode413E30
			count := 1
			if kind == "capacity" {
				count = 5
			}
			for i := 0; i < count; i++ {
				n := new(memDumpNode413E30)
				copy(n.Name[:], fmt.Sprintf("name%d", i))
				n.Count = uint32(i + 1)
				nodes = append(nodes, n)
				if i > 0 {
					nodes[i-1].Next = n
				}
			}
			switch kind {
			case "cycle":
				nodes[0].Next = nodes[0]
			case "unterminated":
				for i := range nodes[0].Name {
					nodes[0].Name[i] = 'x'
				}
			}
			got := memDumpAggregate413E30(table, nodes[0])
			want := 1
			if kind == "capacity" {
				want = capacity
			}
			if got != want || !bytes.Equal(storage[capacity*84:], bytes.Repeat([]byte{0x5A}, 16)) {
				t.Fatal("invalid-list policy escaped fixed table")
			}
			if kind == "unterminated" && len(memDumpName413E30(table[:80])) != 79 {
				t.Fatal("unterminated name crossed count")
			}
		})
	}
}
