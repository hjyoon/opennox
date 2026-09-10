package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterGeneratorUpdate54E930RegisteredNativeSizeAndDispatch(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("MonsterGeneratorUpdate")
	if !ok || callback == nil {
		t.Fatalf("MonsterGeneratorUpdate registration = %p/%v", callback, ok)
	}
	if want := unsafe.Sizeof(server.MonsterGenUpdateData{}); size != want {
		t.Fatalf("MonsterGeneratorUpdate data size = %d, want native size %d", size, want)
	}

	old := monsterGeneratorUpdateCall54E930
	t.Cleanup(func() { monsterGeneratorUpdateCall54E930 = old })
	object := new(server.Object)
	var got *server.Object
	monsterGeneratorUpdateCall54E930 = func(value *server.Object) { got = value }
	server.CallObjectUpdate(callback, object)
	if got != object {
		t.Fatalf("native update argument = %p, want %p", got, object)
	}
}
