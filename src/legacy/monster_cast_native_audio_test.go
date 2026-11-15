package legacy

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Compare actual queued events with the unchanged public C audio entry.
// Queue outside the audio loop so arbitrary signed IDs never index sound data.
func TestMonsterCastNativeDefaultAudioSignedCABI(t *testing.T) {
	s := new(server.Server)
	old := GetServer
	GetServer = func() Server { return &resourceSpellLegacyServer{srv: s} }
	t.Cleanup(func() { GetServer = old })
	owner := resourceSpellCompositionOwner(t, server.Object{})
	ids := []uint32{0, 41, 0x7fffffff, 0x80000000, 0x81234567, 0xffffffff}
	rt := monsterCastDefaultRuntime541300()
	for i, id := range ids {
		rt.Audio(id, owner)
		simplePolypOriginalAudio54CAE0(id, owner)
		queued := reflect.ValueOf(&s.Audio).Elem().FieldByName("delayedObj")
		if queued.Len() != 2*(i+1) {
			t.Fatal("actual default/C queued event count", queued.Len())
		}
		for j := 2 * i; j < 2*i+2; j++ {
			event := queued.Index(j)
			if event.FieldByName("ID").Int() != int64(int32(id)) || event.FieldByName("Obj").Pointer() != uintptr(unsafe.Pointer(owner)) || event.FieldByName("Kind").Int() != 0 || event.FieldByName("Code").Uint() != 0 {
				t.Fatalf("actual default/C event%d id=%d want%d owner=%#x", j, event.FieldByName("ID").Int(), int32(id), event.FieldByName("Obj").Pointer())
			}
		}
	}
}
