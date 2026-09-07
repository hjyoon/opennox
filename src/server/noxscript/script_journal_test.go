package noxscript

import (
	"slices"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
	ns "github.com/opennox/noxscript/ns/v4"
)

type journalBuiltinTestObject struct {
	ns.Obj
}

type journalBuiltinTestCall struct {
	obj     ns.Obj
	message ns.StringID
	typ     ns.EntryType
}

type journalBuiltinTestImpl struct {
	ns.Implementation
	trace   *[]string
	edit    journalBuiltinTestCall
	deleted journalBuiltinTestCall
}

func (s *journalBuiltinTestImpl) JournalEdit(obj ns.Obj, message ns.StringID, typ ns.EntryType) {
	*s.trace = append(*s.trace, "journal-edit")
	s.edit = journalBuiltinTestCall{obj: obj, message: message, typ: typ}
}

func (s *journalBuiltinTestImpl) JournalDelete(obj ns.Obj, message ns.StringID) {
	*s.trace = append(*s.trace, "journal-delete")
	s.deleted = journalBuiltinTestCall{obj: obj, message: message}
}

type journalBuiltinTestVM struct {
	VM
	impl    *journalBuiltinTestImpl
	trace   []string
	ints    []int32
	strings []string
	objects []ns.Obj
}

func (s *journalBuiltinTestVM) NoxScript() ns.Implementation {
	return s.impl
}

func (s *journalBuiltinTestVM) PopI32() int32 {
	s.trace = append(s.trace, "pop-i32")
	v := s.ints[0]
	s.ints = s.ints[1:]
	return v
}

func (s *journalBuiltinTestVM) PopString() string {
	s.trace = append(s.trace, "pop-string")
	v := s.strings[0]
	s.strings = s.strings[1:]
	return v
}

func (s *journalBuiltinTestVM) PopObjectNS() ns.Obj {
	s.trace = append(s.trace, "pop-object")
	v := s.objects[0]
	s.objects = s.objects[1:]
	return v
}

func newJournalBuiltinTestVM(obj ns.Obj, message string, typ int32) *journalBuiltinTestVM {
	vm := &journalBuiltinTestVM{
		ints:    []int32{typ},
		strings: []string{message},
		objects: []ns.Obj{obj},
	}
	vm.impl = &journalBuiltinTestImpl{trace: &vm.trace}
	return vm
}

func TestJournalEditBuiltinNativeDispatchAndPopOrder(t *testing.T) {
	obj := &journalBuiltinTestObject{}
	vm := newJournalBuiltinTestVM(obj, "War01AFirstQuest", 0x1234)
	result, ok := CallBuiltin(vm, asm.BuiltinJournalEdit)
	if !ok || result != 0 {
		t.Fatalf("JournalEdit dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-i32", "pop-string", "pop-object", "journal-edit"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("JournalEdit trace = %v, want %v", vm.trace, wantTrace)
	}
	if call := vm.impl.edit; call.obj != obj || call.message != "War01AFirstQuest" || call.typ != 0x1234 {
		t.Fatalf("JournalEdit call = %#v", call)
	}
}

func TestJournalDeleteBuiltinNativeDispatchAndPopOrder(t *testing.T) {
	obj := &journalBuiltinTestObject{}
	vm := newJournalBuiltinTestVM(obj, "War01AFirstQuest", 0)
	result, ok := CallBuiltin(vm, asm.BuiltinJournalDelete)
	if !ok || result != 0 {
		t.Fatalf("JournalDelete dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "pop-object", "journal-delete"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("JournalDelete trace = %v, want %v", vm.trace, wantTrace)
	}
	if call := vm.impl.deleted; call.obj != obj || call.message != "War01AFirstQuest" {
		t.Fatalf("JournalDelete call = %#v", call)
	}
}
