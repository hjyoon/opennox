package legacy

import (
	"reflect"
	"testing"
)

type objectReplayChecksumFixture4E76C0 struct {
	name string
	next *objectReplayChecksumFixture4E76C0
}

func TestObjectReplayChecksumPass4E76C0Empty(t *testing.T) {
	var events []string
	got := objectReplayChecksumPass4E76C0(objectReplayChecksumPassHooks4E76C0[*objectReplayChecksumFixture4E76C0]{
		first: func() *objectReplayChecksumFixture4E76C0 {
			events = append(events, "first")
			return nil
		},
		noop: func(*objectReplayChecksumFixture4E76C0) {
			t.Fatal("no-op called for empty list")
		},
		checksum: func(*objectReplayChecksumFixture4E76C0) uint32 {
			t.Fatal("checksum called for empty list")
			return 0
		},
		next: func(*objectReplayChecksumFixture4E76C0) *objectReplayChecksumFixture4E76C0 {
			t.Fatal("next called for empty list")
			return nil
		},
	})
	if got != nil {
		t.Fatalf("terminal object = %p, want nil", got)
	}
	if want := []string{"first"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectReplayChecksumPass4E76C0OrderAndPostChecksumReload(t *testing.T) {
	a := &objectReplayChecksumFixture4E76C0{name: "a"}
	b := &objectReplayChecksumFixture4E76C0{name: "b"}
	c := &objectReplayChecksumFixture4E76C0{name: "c"}
	a.next = b

	var events []string
	got := objectReplayChecksumPass4E76C0(objectReplayChecksumPassHooks4E76C0[*objectReplayChecksumFixture4E76C0]{
		first: func() *objectReplayChecksumFixture4E76C0 {
			events = append(events, "first")
			return a
		},
		noop: func(obj *objectReplayChecksumFixture4E76C0) {
			events = append(events, "noop:"+obj.name)
		},
		checksum: func(obj *objectReplayChecksumFixture4E76C0) uint32 {
			events = append(events, "checksum:"+obj.name)
			if obj == a {
				obj.next = c
			}
			return uint32(len(obj.name))
		},
		next: func(obj *objectReplayChecksumFixture4E76C0) *objectReplayChecksumFixture4E76C0 {
			events = append(events, "next:"+obj.name)
			return obj.next
		},
	})
	if got != nil {
		t.Fatalf("terminal object = %p, want nil", got)
	}
	want := []string{
		"first",
		"noop:a", "checksum:a", "next:a",
		"noop:c", "checksum:c", "next:c",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectReplayChecksum4E7700FieldWidths(t *testing.T) {
	tests := []struct {
		name string
		set  func(*objectReplayChecksumInput4E7700)
		want uint32
	}{
		{name: "team byte zero extends", set: func(v *objectReplayChecksumInput4E7700) { v.teamID = 0x85 }, want: 0x85},
		{name: "type word zero extends", set: func(v *objectReplayChecksumInput4E7700) { v.typeInd = 0x8001 }, want: 0x8001},
		{name: "script id keeps bits", set: func(v *objectReplayChecksumInput4E7700) { v.scriptID = -2 }, want: 0xfffffffe},
		{name: "position x bits", set: func(v *objectReplayChecksumInput4E7700) { v.posXBits = 0x80000000 }, want: 0x80000000},
		{name: "extent", set: func(v *objectReplayChecksumInput4E7700) { v.extent = 0x01020304 }, want: 0x01020304},
		{name: "network code", set: func(v *objectReplayChecksumInput4E7700) { v.netCode = 0x11121314 }, want: 0x11121314},
		{name: "field five", set: func(v *objectReplayChecksumInput4E7700) { v.field5 = 0x21222324 }, want: 0x21222324},
		{name: "object flags", set: func(v *objectReplayChecksumInput4E7700) { v.objFlags = 0x31323334 }, want: 0x31323334},
		{name: "position y bits", set: func(v *objectReplayChecksumInput4E7700) { v.posYBits = 0x7fc12345 }, want: 0x7fc12345},
		{name: "new position x bits", set: func(v *objectReplayChecksumInput4E7700) { v.newPosXBits = 0x41424344 }, want: 0x41424344},
		{name: "new position y bits", set: func(v *objectReplayChecksumInput4E7700) { v.newPosYBits = 0x51525354 }, want: 0x51525354},
		{name: "previous position x bits", set: func(v *objectReplayChecksumInput4E7700) { v.prevPosXBits = 0x61626364 }, want: 0x61626364},
		{name: "previous position y bits", set: func(v *objectReplayChecksumInput4E7700) { v.prevPosYBits = 0x71727374 }, want: 0x71727374},
		{name: "velocity x bits", set: func(v *objectReplayChecksumInput4E7700) { v.velXBits = 0x81828384 }, want: 0x81828384},
		{name: "velocity y bits", set: func(v *objectReplayChecksumInput4E7700) { v.velYBits = 0x91929394 }, want: 0x91929394},
		{name: "force x bits", set: func(v *objectReplayChecksumInput4E7700) { v.forceXBits = 0xa1a2a3a4 }, want: 0xa1a2a3a4},
		{name: "force y bits", set: func(v *objectReplayChecksumInput4E7700) { v.forceYBits = 0xb1b2b3b4 }, want: 0xb1b2b3b4},
		{name: "position 24 x bits", set: func(v *objectReplayChecksumInput4E7700) { v.pos24XBits = 0xc1c2c3c4 }, want: 0xc1c2c3c4},
		{name: "position 24 y bits", set: func(v *objectReplayChecksumInput4E7700) { v.pos24YBits = 0xd1d2d3d4 }, want: 0xd1d2d3d4},
		{name: "z bits", set: func(v *objectReplayChecksumInput4E7700) { v.zBits = 0xe1e2e3e4 }, want: 0xe1e2e3e4},
		{name: "field 27 bits", set: func(v *objectReplayChecksumInput4E7700) { v.field27Bits = 0xf1f2f3f4 }, want: 0xf1f2f3f4},
		{name: "direction one sign extends", set: func(v *objectReplayChecksumInput4E7700) { v.direction1 = -32768 }, want: 0xffff8000},
		{name: "direction two sign extends", set: func(v *objectReplayChecksumInput4E7700) { v.direction2 = -2 }, want: 0xfffffffe},
		{name: "field 38", set: func(v *objectReplayChecksumInput4E7700) { v.field38 = 0x38383838 }, want: 0x38383838},
		{name: "field 37", set: func(v *objectReplayChecksumInput4E7700) { v.field37 = 0x37373737 }, want: 0x37373737},
		{name: "field 34", set: func(v *objectReplayChecksumInput4E7700) { v.field34 = 0x34343434 }, want: 0x34343434},
		{name: "field 33", set: func(v *objectReplayChecksumInput4E7700) { v.field33 = 0x33333333 }, want: 0x33333333},
		{name: "field 32", set: func(v *objectReplayChecksumInput4E7700) { v.field32 = 0x32323232 }, want: 0x32323232},
		{name: "mass bits", set: func(v *objectReplayChecksumInput4E7700) { v.massBits = 0xff800000 }, want: 0xff800000},
		{name: "buffs", set: func(v *objectReplayChecksumInput4E7700) { v.buffs = 0x85858585 }, want: 0x85858585},
		{name: "field 62", set: func(v *objectReplayChecksumInput4E7700) { v.field62 = 0x62626262 }, want: 0x62626262},
		{name: "health max word", set: func(v *objectReplayChecksumInput4E7700) { v.healthPresent, v.healthMax = true, 0x8001 }, want: 0x8001},
		{name: "health field two word", set: func(v *objectReplayChecksumInput4E7700) { v.healthPresent, v.healthField2 = true, 0x9002 }, want: 0x9002},
		{name: "health current word", set: func(v *objectReplayChecksumInput4E7700) { v.healthPresent, v.healthCur = true, 0xa003 }, want: 0xa003},
		{name: "absent health ignores words", set: func(v *objectReplayChecksumInput4E7700) { v.healthMax, v.healthField2, v.healthCur = 1, 2, 4 }, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var in objectReplayChecksumInput4E7700
			tc.set(&in)
			if got := objectReplayChecksum4E7700(in); got != tc.want {
				t.Fatalf("checksum = %#08x, want %#08x", got, tc.want)
			}
		})
	}
}
