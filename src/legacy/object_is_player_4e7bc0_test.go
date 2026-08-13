package legacy

import "testing"

type objectIsPlayerFixture4E7BC0 struct {
	class uint32
}

func TestObjectIsPlayer4E7BC0NullSkipsClassLoad(t *testing.T) {
	loads := 0
	got := objectIsPlayer4E7BC0[*objectIsPlayerFixture4E7BC0](nil, func(*objectIsPlayerFixture4E7BC0) uint32 {
		loads++
		return ^uint32(0)
	})
	if got != 0 || loads != 0 {
		t.Fatalf("nil result = %d with %d loads, want 0 with no loads", got, loads)
	}
}

func TestObjectIsPlayer4E7BC0LoadsFullClassOnce(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class uint32
		want  uint32
	}{
		{name: "zero"},
		{name: "lower bits only", class: 0x00000003},
		{name: "player only", class: 0x00000004, want: 1},
		{name: "player with neighbors", class: 0x0000000f, want: 1},
		{name: "all except player", class: 0xfffffffb},
		{name: "all bits", class: 0xffffffff, want: 1},
		{name: "high and player", class: 0x80000004, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &objectIsPlayerFixture4E7BC0{class: tc.class}
			loads := 0
			got := objectIsPlayer4E7BC0(obj, func(gotObj *objectIsPlayerFixture4E7BC0) uint32 {
				loads++
				if gotObj != obj {
					t.Fatalf("class object = %p, want %p", gotObj, obj)
				}
				return gotObj.class
			})
			if got != tc.want || loads != 1 {
				t.Fatalf("class %#08x result = %d with %d loads, want %d with one load", tc.class, got, loads, tc.want)
			}
			if obj.class != tc.class {
				t.Fatalf("class changed from %#08x to %#08x", tc.class, obj.class)
			}
		})
	}
}
