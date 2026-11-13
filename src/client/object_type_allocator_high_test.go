package client

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestObjectTypeExplicitHighOwnerIsolation(t *testing.T) {
	var ordinary, high clientObjTypes
	high.highAddresses = true
	for _, owner := range []*clientObjTypes{&ordinary, &high} {
		owner.Nox_things_free_44C580_B()
		t.Cleanup(owner.Nox_things_free_44C580_B)
		name := "NativeOwner"
		raw := append([]byte{byte(len(name))}, name...)
		raw = append(raw, 0)
		data, _ := alloc.CloneSliceHigh(raw)
		file := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
		err := owner.ReadType(file, nil)
		file.Free()
		if err != nil {
			t.Fatal(err)
		}
	}
	a, b := ordinary.TypeByID("NativeOwner"), high.TypeByID("NativeOwner")
	if a == nil || b == nil || a == b || a.UsesHighNativeMemory() || !b.UsesHighNativeMemory() {
		t.Fatal("native type owners lost independent allocation policies")
	}
	if a.Name == b.Name || alloc.GoString(a.Name) != alloc.GoString(b.Name) {
		t.Fatal("native type names lost separate foreign ownership or text")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(b.C()) <= 0xffffffff || uintptr(unsafe.Pointer(b.Name)) <= 0xffffffff) {
		t.Fatal("explicit type or name lacks high native address bits")
	}
	runtime.GC()
	if high.TypeByID("NativeOwner") != b || b.Index() != a.Index() {
		t.Fatal("GC changed the actual native type registry")
	}
	high.Nox_things_free_44C580_B()
	if b.UsesHighNativeMemory() || ordinary.TypeByID("NativeOwner") != a {
		t.Fatal("native type release retained its policy or affected another owner")
	}
}
