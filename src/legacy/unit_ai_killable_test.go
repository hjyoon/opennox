package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterSightKillable528190(t *testing.T) {
	tests := []struct {
		name string
		cur  uint16
		max  uint16
		want int
	}{
		{name: "alive", cur: 1, max: 10, want: 1},
		{name: "zero maximum", cur: 0, max: 0, want: 1},
		{name: "positive with zero maximum", cur: 1, max: 0, want: 1},
		{name: "dead", cur: 0, max: 10, want: 0},
	}
	if got := Nox_xxx_checkIsKillable_528190(&server.Object{}); got != 0 {
		t.Fatalf("nil health result = %d, want 0", got)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj, freeObj := alloc.New(server.Object{})
			defer freeObj()
			health, freeHealth := alloc.New(server.HealthData{})
			defer freeHealth()
			health.Cur, health.Max = tc.cur, tc.max
			obj.HealthData = health
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= uintptr(^uint32(0)) {
				t.Fatalf("test object address = %#x, want native high address", uintptr(unsafe.Pointer(obj)))
			}
			if got := Nox_xxx_checkIsKillable_528190(obj); got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
		})
	}
}
