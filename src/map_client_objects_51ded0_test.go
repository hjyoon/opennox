package opennox

import (
	"fmt"
	"image"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func TestSaveClientObjectsNative51DED0OrderAndFields(t *testing.T) {
	third := &client.Drawable{TypeIDVal: 33}
	second := &client.Drawable{TypeIDVal: 22, NextPtr: third}
	first := &client.Drawable{
		TypeIDVal:  11,
		PosVec:     image.Pt(-7, 19),
		NetCode32:  0x81234567,
		ObjFlags:   object.Flags(0x10203040),
		Flags70Val: 0x50607080,
		NextPtr:    second,
	}
	var events []string
	temp := &server.Object{NetCode: 0x99}
	err := saveClientObjectsNative51DED0(saveClientObjectsHooks51DED0{
		first: func() *client.Drawable { return first },
		serverVisible: func(ind int) bool {
			events = append(events, fmt.Sprintf("visible:%d", ind))
			return ind == 22
		},
		saveable: func(ind int) bool {
			events = append(events, fmt.Sprintf("saveable:%d", ind))
			return ind != 33
		},
		typeName: func(ind int) string {
			events = append(events, fmt.Sprintf("name:%d", ind))
			return "InvisibleLight"
		},
		newObject: func(name string) *server.Object {
			events = append(events, "new:"+name)
			return temp
		},
		saveObject: func(obj *server.Object) int {
			events = append(events, "save")
			if obj != temp {
				t.Fatalf("saved object = %p, want %p", obj, temp)
			}
			if obj.PosVec.X != -6.5 || obj.PosVec.Y != 19.5 {
				t.Errorf("position = (%v,%v), want (-6.5,19.5)", obj.PosVec.X, obj.PosVec.Y)
			}
			if obj.NetCode != first.NetCode32 || obj.Extent != first.NetCode32 ||
				obj.ScriptIDVal != int32(first.NetCode32) {
				t.Errorf("identity fields = %#x/%#x/%#x", obj.NetCode, obj.Extent, obj.ScriptIDVal)
			}
			if obj.ObjFlags != first.ObjFlags || obj.Field5 != first.Flags70Val {
				t.Errorf("flags = %#x/%#x, want %#x/%#x", obj.ObjFlags, obj.Field5, first.ObjFlags, first.Flags70Val)
			}
			return 0
		},
		freeObject: func(obj *server.Object) {
			events = append(events, "free")
			if obj.NetCode != 0x99 {
				t.Errorf("restored net code = %#x, want 0x99", obj.NetCode)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"visible:11", "saveable:11", "name:11", "new:InvisibleLight", "save", "free",
		"visible:22", "visible:33", "saveable:33",
	}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
