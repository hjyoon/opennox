package legacy

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// This is a representation-only consumer. It expands observed original words;
// it contains no Create behavior, spell flags, class rules, or timer arithmetic.
type playerBotCreateMapping4FA700 struct {
	Original   int
	Native     uintptr
	Bytes      int
	Pointer    bool
	NativeWord bool
}

func playerBotCreateMappings4FA700(t *testing.T) []playerBotCreateMapping4FA700 {
	t.Helper()
	var out []playerBotCreateMapping4FA700
	var walk func(reflect.Type, int, uintptr) int
	walk = func(typ reflect.Type, original int, native uintptr) int {
		switch typ.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Uintptr:
			out = append(out, playerBotCreateMapping4FA700{original, native, 4, typ.Kind() != reflect.Uintptr, true})
			return 4
		case reflect.Array:
			n := 0
			for i := 0; i < typ.Len(); i++ {
				n += walk(typ.Elem(), original+n, native+uintptr(i)*typ.Elem().Size())
			}
			return n
		case reflect.Struct:
			n := 0
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				n += walk(f.Type, original+n, native+f.Offset)
			}
			return n
		case reflect.Uint8, reflect.Int8, reflect.Uint16, reflect.Int16, reflect.Uint32, reflect.Int32, reflect.Float32:
			n := int(typ.Size())
			out = append(out, playerBotCreateMapping4FA700{original, native, n, false, false})
			return n
		default:
			t.Fatalf("unclassified original record type %s", typ)
			return 0
		}
	}
	typ := reflect.TypeOf(server.MonsterUpdateData{})
	seen := make([]bool, 2200)
	for _, spec := range playerBotCreateOriginalFields4FA700 {
		f, ok := typ.FieldByName(spec.Name)
		if !ok {
			t.Fatal("native field missing", spec.Name)
		}
		start := len(out)
		if n := walk(f.Type, spec.Original, f.Offset); n != spec.Bytes {
			t.Fatalf("original field width %s:got%d original%d", spec.Name, n, spec.Bytes)
		}
		for _, m := range out[start:] {
			nativeBytes := uintptr(m.Bytes)
			if m.NativeWord {
				nativeBytes = unsafe.Sizeof(uintptr(0))
			}
			if m.Native < f.Offset || m.Native+nativeBytes > f.Offset+f.Type.Size() {
				t.Fatal("mapping crosses actual native field", spec.Name, m)
			}
			if unsafe.Sizeof(uintptr(0)) == 4 && int(m.Native) != m.Original {
				t.Fatal("PE32 offset differs from original", spec.Name, m)
			}
			for i := m.Original; i < m.Original+m.Bytes; i++ {
				if i < 0 || i >= len(seen) || seen[i] {
					t.Fatal("overlap or out-of-record original byte", i)
				}
				seen[i] = true
			}
		}
	}
	for i, v := range seen {
		if !v {
			t.Fatal("unclassified original byte", i)
		}
	}
	return out
}

func playerBotCreateExpandOriginal4FA700(t *testing.T, original []byte, roles map[uint32]uintptr) []byte {
	t.Helper()
	if len(original) != 2200 {
		t.Fatal("complete original monster record required")
	}
	native := make([]byte, int(unsafe.Sizeof(server.MonsterUpdateData{})))
	for _, m := range playerBotCreateMappings4FA700(t) {
		src := original[m.Original : m.Original+m.Bytes]
		dst := native[m.Native:]
		if !m.NativeWord {
			copy(dst, src)
			continue
		}
		value := binary.LittleEndian.Uint32(src)
		wide := uintptr(value)
		if m.Pointer {
			var ok bool
			wide, ok = roles[value]
			if !ok {
				t.Fatalf("unclassified original pointer %#x at%d", value, m.Original)
			}
		}
		if unsafe.Sizeof(wide) == 8 {
			binary.LittleEndian.PutUint64(dst, uint64(wide))
		} else {
			binary.LittleEndian.PutUint32(dst, uint32(wide))
		}
	}
	return native
}
