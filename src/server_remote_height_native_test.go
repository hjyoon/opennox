package opennox

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Object visibility updates already use the native height reporter. The
// remote-player fifteen-second refresh must not re-enter the PE32 C body.
func TestPeriodicRemoteUpdateUsesNativeHeightReporter(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "server.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "nox_xxx_netUpdate_518EE0" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("remote player update missing")
	}
	native := false
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if fn.Sel.Name == "Nox_xxx_netReportUnitHeight_4D9020" {
			t.Error("remote periodic refresh calls the pointer-truncating legacy height reporter")
		}
		if fn.Sel.Name == "netReportUnitHeightNative4D9020" {
			native = true
		}
		return true
	})
	if !native {
		t.Error("remote periodic refresh is not wired to the existing native height reporter")
	}
}

func TestNativeHeightReporterQueuesFixedWidthPacketsForHighObject(t *testing.T) {
	tests := []struct {
		name   string
		status uint32
		z, dz  float32
		tail   float32
		want   []byte
	}{
		{name: "zero", want: []byte{94, 0x45, 0x23, 0}},
		{name: "positive truncation", z: 258.75, want: []byte{94, 0x45, 0x23, 2}},
		{name: "negative truncation", z: -258.75, want: []byte{95, 0x45, 0x23, 2}},
		{name: "ballistic fields", status: 0x20, z: -1.75, dz: 257.5, tail: 511.875, want: []byte{159, 0x45, 0x23, 255, 1, 255}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, free := alloc.New(server.Object{})
			defer free()
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				t.Fatal("native object must exceed 4 GiB")
			}
			obj.ObjClass = object.ClassPlayer
			obj.NetCode = 0x2345
			obj.Field5 = test.status
			obj.ZVal, obj.Field27, obj.Field29 = test.z, test.dz, math.Float32bits(test.tail)
			raw := unsafe.Slice((*byte)(unsafe.Pointer(obj)), unsafe.Sizeof(*obj))
			before := bytes.Clone(raw)
			list := netlist.New()
			list.Init()
			defer list.Free()
			s := &Server{Server: &server.Server{NetList: list}}
			s.netReportUnitHeightNative4D9020(5, obj)
			if got := list.CopyPacketsA(5, netlist.Kind1); !bytes.Equal(got, test.want) {
				t.Fatalf("height packet = % x, want % x", got, test.want)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("height reporting modified the native object")
			}
			if got := list.CopyPacketsA(4, netlist.Kind1); len(got) != 0 {
				t.Fatalf("height packet sent to another recipient: % x", got)
			}
		})
	}
}
