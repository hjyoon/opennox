package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
)

// A real SCRIPT03 stream: two GLOBAL records and one callable function.
// Metadata vectors below are derived from the original loader/CRT instructions,
// not from asm.ReadScript's derived fields.
func scriptQualifierStreamR317(name string) []byte {
	var out bytes.Buffer
	u32 := func(value uint32) { _ = binary.Write(&out, binary.LittleEndian, value) }
	str := func(value string) { u32(uint32(len(value))); out.WriteString(value) }
	out.WriteString("SCRIPT03STRG")
	u32(0)
	out.WriteString("CODE")
	u32(3)
	for index, fname := range []string{"GLOBAL", "GLOBAL", name} {
		out.WriteString("FUNC")
		str(fname)
		u32(0)
		u32(0)
		out.WriteString("SYMB")
		count := uint32(0)
		if index == 1 {
			count = 4
		}
		u32(count)
		u32(0)
		for i := uint32(0); i < count; i++ {
			u32(1)
		}
		out.WriteString("DATA")
		u32(4)
		u32(uint32(asm.OpReturn))
	}
	out.WriteString("DONE")
	return out.Bytes()
}

func TestScriptQualifier505360ReadScriptOriginalVectorsR317(t *testing.T) {
	qualified := "%7%13%-29"
	for _, tc := range []struct {
		label, name, suffix string
		dx, dy              int
	}{
		{"unqualified", "Probe", "", 0, 0},
		{"independent_axes", "Probe%7%13%-29", "%7", 13, -29},
		{"reversed_signs", "Probe%0%-91%37", "%0", -91, 37},
		{"zero", "Probe%00%0%0", "%00", 0, 0},
		{"consecutive_delimiters", "Probe%%7%%%13%%-29%", "%7", 13, -29},
		{"leading_delimiters", "%%Probe%7%13%-29", "%7", 13, -29},
		{"extra_tokens_ignored", "Probe%7%13%-29%123%456", "%7", 13, -29},
		{"text_ordinal", "Probe%tag%+13% -29junk%extra", "%tag", 13, -29},
		{"ordinal_is_not_trimmed", "Probe%  tag  %13%-29", "%  tag  ", 13, -29},
		{"ascii_whitespace", "Probe%7%\t\n\v\f\r +13%\t -29tail", "%7", 13, -29},
		{"decimal_prefix", "Probe%7%0013end%-0029tail", "%7", 13, -29},
		{"nonnumeric_and_sign_only", "Probe%7%junk%+", "%7", 0, 0},
		{"double_sign", "Probe%7%--1%+-2", "%7", 0, 0},
		{"signed_limits", "Probe%7%2147483647%-2147483648", "%7", 2147483647, -2147483648},
		{"signed_overflow", "Probe%7%2147483648%-2147483649", "%7", -2147483648, 2147483647},
		{"unsigned_overflow", "Probe%7%4294967295%-4294967296", "%7", -1, 0},
		{"multiple_wraps", "Probe%7%18446744073709551617%-18446744073709551617", "%7", 1, -1},
		{"unicode_name_and_ordinal", "마법사%방%13%-29", "%방", 13, -29},
		{"name_1023_bytes", strings.Repeat("F", 1023-len(qualified)) + qualified, "%7", 13, -29},
	} {
		t.Run(tc.label, func(t *testing.T) {
			var vm NoxScriptVM
			if err := vm.ReadScript(bytes.NewReader(scriptQualifierStreamR317(tc.name))); err != nil {
				t.Fatal(err)
			}
			funcs := vm.Funcs()
			if len(funcs) != 3 || funcs[0].Name() != "GLOBAL" || funcs[1].Name() != "GLOBAL" || len(funcs[1].Values) != 4 {
				t.Fatal("SCRIPT03 global/function structure changed")
			}
			got := funcs[2]
			if got.Name() != tc.name || got.NamePref != tc.suffix || got.PosOff != image.Pt(tc.dx, tc.dy) {
				t.Fatalf("full name/suffix/offset = %q/%q/%v, want %q/%q/%v", got.Name(), got.NamePref, got.PosOff, tc.name, tc.suffix, image.Pt(tc.dx, tc.dy))
			}
			if got.Return != 0 || got.FuncDef.Args != 0 || len(got.Code) != 1 || got.Code[0] != uint32(asm.OpReturn) {
				t.Fatal("qualifier decoding changed callable code/arity")
			}
		})
	}
}

func TestScriptQualifier505360BuiltinContextR317(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "builtin_error"}[fail], func(t *testing.T) {
			var vm NoxScriptVM
			if err := vm.ReadScript(bytes.NewReader(scriptQualifierStreamR317("Probe%tag%13%-29"))); err != nil {
				t.Fatal(err)
			}
			fn := &vm.vm.funcs[2]
			called := false
			sentinel := errors.New("controlled builtin error")
			vm.CallBuiltinNative = func(fi asm.Builtin) error {
				called = true
				if fi != asm.BuiltinMoveObject || vm.NameSuff() != "%tag" || vm.DPos() != image.Pt(13, -29) {
					t.Fatalf("builtin saw wrong qualifier context: %v/%q/%v", fi, vm.NameSuff(), vm.DPos())
				}
				if fail {
					return sentinel
				}
				return nil
			}
			err := vm.callBuiltin(fn, asm.BuiltinMoveObject)
			if !called || (fail && err != sentinel) || (!fail && err != nil) {
				t.Fatal("builtin dispatch outcome changed")
			}
			if vm.NameSuff() != "" || vm.DPos() != (image.Point{}) {
				t.Fatal("builtin left temporary qualifier context")
			}
			if fn.Name() != "Probe%tag%13%-29" || fn.PosOff != image.Pt(13, -29) {
				t.Fatal("builtin changed persistent function metadata")
			}
		})
	}
}

func TestScriptQualifier505360KeepsExistingNameBoundaryR317(t *testing.T) {
	var vm NoxScriptVM
	err := vm.ReadScript(bytes.NewReader(scriptQualifierStreamR317(strings.Repeat("F", 1024))))
	if err == nil || err.Error() != "function name too long: 1024" {
		t.Fatalf("name boundary error = %v", err)
	}
}
