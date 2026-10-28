package opennox

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectAIParametersNativeScriptAccess(t *testing.T) {
	properties := []struct {
		name  string
		value float32
		set   func(*Object, float32)
		get   func(*Object) float32
	}{
		{"aggression", 0.625, (*Object).SetAggression, (*Object).Aggression},
		{"retreat", 0.25, (*Object).SetRetreatLevel, (*Object).RetreatLevel},
		{"regroup", 0.875, (*Object).SetRegroupLevel, (*Object).RegroupLevel},
	}
	classes := []struct {
		name    string
		class   object.Class
		monster bool
	}{
		{"monster", object.ClassMonster, true},
		{"monster-with-other-classes", object.ClassMonster | object.ClassLight | object.ClassComplex, true},
		{"player", object.ClassPlayer | object.ClassLight | object.ClassComplex, false},
		{"simple-object", object.ClassSimple, false},
		{"missile", object.ClassMissile | object.ClassSimple, false},
	}
	for _, p := range properties {
		for _, c := range classes {
			t.Run(p.name+"/"+c.name, func(t *testing.T) {
				o, freeObject := alloc.New(server.Object{ObjClass: c.class})
				t.Cleanup(freeObject)
				*o = server.Object{ObjClass: c.class}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= 0xffffffff {
					t.Fatal("native object must be above4GiB")
				}
				var data *server.MonsterUpdateData
				if c.monster {
					var freeData func()
					data, freeData = alloc.New(server.MonsterUpdateData{})
					*data = server.MonsterUpdateData{}
					t.Cleanup(func() { o.UpdateData = nil; freeData() })
					o.UpdateData = unsafe.Pointer(data)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(o.UpdateData) <= 0xffffffff {
						t.Fatal("native monster update must be above4GiB")
					}
				}
				// Use the public script-facing wrapper, which also backs Lua's
				// aggression/retreat/regroup properties. Non-monsters have no
				// monster update record and must not dereference one.
				unit := asObjectS(o)
				objectBytes := unsafe.Slice((*byte)(unsafe.Pointer(o)), int(unsafe.Sizeof(*o)))
				beforeSet := bytes.Clone(objectBytes)
				p.set(unit, p.value)
				if !c.monster && !bytes.Equal(objectBytes, beforeSet) {
					t.Fatal("non-monster setter modified the native object")
				}
				beforeRead := bytes.Clone(objectBytes)
				var updateBytes, beforeUpdate []byte
				if data != nil {
					updateBytes = unsafe.Slice((*byte)(unsafe.Pointer(data)), int(unsafe.Sizeof(*data)))
					beforeUpdate = bytes.Clone(updateBytes)
				}
				defer func() {
					if panicValue := recover(); panicValue != nil {
						t.Errorf("script property dereferenced non-monster update data: %v", panicValue)
					}
				}()
				got := p.get(unit)
				want := float32(0)
				if c.monster {
					want = p.value
				}
				if got != want {
					t.Errorf("script property after setter = %g, want %g", got, want)
				}
				if !bytes.Equal(objectBytes, beforeRead) || !bytes.Equal(updateBytes, beforeUpdate) {
					t.Error("script property read modified native object/update bytes")
				}
			})
		}
	}
}
