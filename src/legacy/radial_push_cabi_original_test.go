package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func radialCFrameOriginal(t *testing.T, size uintptr) []byte {
	t.Helper()
	p, free := alloc.Make([]byte(nil), int(size)+64)
	t.Cleanup(free)
	for n := range p {
		p[n] = 0xa5
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&p[32])) <= math.MaxUint32 {
		t.Fatal("native C frame below 4 GiB")
	}
	return p
}

func TestRadialPushRootCNativeMapOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/radial_push_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name  string `json:"name"`
				Order []int  `json:"order"`
			} `json:"input"`
			Objects [][7]uint32 `json:"objects"`
			Origin  [2]uint32   `json:"origin"`
			Events  []struct {
				Service string   `json:"service"`
				Args    []uint32 `json:"args"`
			} `json:"events"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	rows := map[int]int{}
	for n, row := range f.Rows {
		if row.Mode == "actual-startup" && strings.HasPrefix(row.Input.Name, "native-map-active-missiles/") {
			rows[row.Input.Order[1]] = n
		}
	}
	if len(rows) != 2 {
		t.Fatal("missing original native-map orders")
	}
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse-%t", reverse), func(t *testing.T) {
			srv := server.New(nil, nil, strman.New())
			t.Cleanup(srv.Close)
			srv.Map.Init()
			t.Cleanup(srv.Map.Free)
			if srv.Walls.Init() == 0 {
				t.Fatal("native empty walls initialization")
			}
			t.Cleanup(srv.Walls.Free)
			oldServer := GetServer
			GetServer = func() Server { return &powderCService53C9A0{srv: srv} }
			t.Cleanup(func() { GetServer = oldServer })
			var objects [2]*server.Object
			var frames [2][]byte
			for n := range objects {
				p := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
				frames[n] = p
				o := (*server.Object)(unsafe.Pointer(&p[32]))
				objects[n] = o
				*o = server.Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, Mass: 1, PosVec: types.Ptf(305.1+float32(n)*5, 300.1), Field37: 0x55667788, Field38: 0x11223344}
				o.NewPos = o.PosVec
				o.PrevPos = o.PosVec
				o.Shape.Kind = server.ShapeKindCircle
				o.Shape.Circle.R = 1
				o.Shape.Circle.R2 = 1
			}
			order := []int{0, 1}
			if reverse {
				order = []int{1, 0}
			}
			for _, index := range order {
				srv.Map.AddObjectToIndex(objects[index])
			}
			defer func() {
				for _, o := range objects {
					srv.Map.RemoveObjectFromIndex(o)
				}
			}()
			pointFrame := radialCFrameOriginal(t, unsafe.Sizeof(types.Pointf{}))
			point := (*types.Pointf)(unsafe.Pointer(&pointFrame[32]))
			*point = types.Ptf(300.1, 300.1)
			_, probeSize := radialPushCSize52E0E0()
			probeFrame := radialCFrameOriginal(t, probeSize)
			probe := unsafe.Pointer(&probeFrame[32])
			radialPushCProbeInit52E0E0(probe)
			radialPushCProbeOrigin52E0E0(probe, point)
			before := [][]byte{bytes.Clone(frames[0]), bytes.Clone(frames[1])}
			radialPushCRoot52E040(point, 50, 10, 100, nil, radialPushCProbePointer52E0E0(), uintptr(probe))
			offsets := radialPushCProbeOffsets52E0E0()
			readWord := func(offset uintptr) uintptr {
				if unsafe.Sizeof(uintptr(0)) == 8 {
					return uintptr(binary.LittleEndian.Uint64(probeFrame[32+offset:]))
				}
				return uintptr(binary.LittleEndian.Uint32(probeFrame[32+offset:]))
			}
			if binary.LittleEndian.Uint32(probeFrame[32+offsets[0]:]) != 2 {
				t.Fatalf("native spatial query callbacks=%d, want 2", binary.LittleEndian.Uint32(probeFrame[32+offsets[0]:]))
			}
			last := -1
			for n, o := range objects {
				if readWord(offsets[1]) == uintptr(unsafe.Pointer(o)) {
					last = n
				}
			}
			if last < 0 {
				t.Fatal("foreign C callback object")
			}
			row := f.Rows[rows[last]]
			for n, o := range objects {
				binary.LittleEndian.PutUint32(before[n][32+unsafe.Offsetof(o.ForceVec):], row.Objects[n][2])
				binary.LittleEndian.PutUint32(before[n][36+unsafe.Offsetof(o.ForceVec):], row.Objects[n][3])
				if !bytes.Equal(frames[n], before[n]) {
					t.Fatalf("native indexed object %d differs from original order %v", n, row.Input.Order)
				}
			}
			if math.Float32bits(point.X) != row.Origin[0] || math.Float32bits(point.Y) != row.Origin[1] || readWord(offsets[3]) != uintptr(probe) {
				t.Fatal("original live point or native argument word changed")
			}
			for _, e := range row.Events {
				if e.Service == "before:1" && binary.LittleEndian.Uint32(probeFrame[32+offsets[2]:]) != e.Args[1] {
					t.Fatal("second original distance word changed")
				}
			}
		})
	}
}

func TestRadialPushCArgumentWordsAndPointReference(t *testing.T) {
	pointFrame := radialCFrameOriginal(t, unsafe.Sizeof(types.Pointf{}))
	sourceFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	point := (*types.Pointf)(unsafe.Pointer(&pointFrame[32]))
	source := (*server.Object)(unsafe.Pointer(&sourceFrame[32]))
	*point = types.Ptf(300, 400)
	*source = server.Object{}
	callback := radialPushCProbePointer52E0E0()
	arg := uintptr(unsafe.Pointer(source))
	if unsafe.Sizeof(uintptr(0)) > 4 && callback <= math.MaxUint32 {
		t.Fatal("callback code below 4 GiB")
	}
	old := mapPushUnitsAroundCall52E040
	t.Cleanup(func() { mapPushUnitsAroundCall52E040 = old })
	calls := 0
	mapPushUnitsAroundCall52E040 = func(p *types.Pointf, outer, inner, force float32, s *server.Object, cb, data int) {
		calls++
		if p != point || s != source || uintptr(cb) != callback || uintptr(data) != arg || outer != 100 || inner != 30 || force != 60 {
			t.Fatal("native C pointer or callback word changed")
		}
		p.X = 301
	}
	radialPushCRoot52E040(point, 100, 30, 60, source, callback, arg)
	if calls != 1 || point.X != 301 {
		t.Fatal("C entry copied the borrowed point")
	}
}

func TestRadialPushCandidateCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/radial_push_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string `json:"original_sha256"`
		Contexts   int    `json:"contexts"`
		Executions int    `json:"root_executions"`
		Rows       []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind     string      `json:"kind"`
				Name     string      `json:"name"`
				Origin   [2]uint32   `json:"origin"`
				Outer    uint32      `json:"outer"`
				Inner    uint32      `json:"inner"`
				Force    uint32      `json:"force"`
				Objects  [][7]uint32 `json:"objects"`
				Callback bool        `json:"callback"`
			} `json:"input"`
			Objects [][7]uint32 `json:"objects"`
			Events  []struct {
				Service string   `json:"service"`
				Args    []uint32 `json:"args"`
			} `json:"events"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Contexts != 454 || f.Executions != 908 || len(f.Rows) != 908 {
		t.Fatal("incomplete original radial fixture")
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	if srv.Walls.Init() == 0 {
		t.Fatal("empty native wall index initialization")
	}
	t.Cleanup(srv.Walls.Free)
	oldServer := GetServer
	GetServer = func() Server { return &powderCService53C9A0{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	argsSize, probeSize := radialPushCSize52E0E0()
	callback := radialPushCProbePointer52E0E0()
	offsets := radialPushCProbeOffsets52E0E0()
	if argsSize != map[uintptr]uintptr{4: 28, 8: 48}[unsafe.Sizeof(uintptr(0))] {
		t.Fatal("native callback record size")
	}
	count := 0
	for _, row := range f.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || i.Kind != "candidate" || len(i.Name) < 19 || i.Name[:19] != "candidate-geometry/" || i.Objects[0][5] != 1 {
			continue
		}
		count++
		t.Run(i.Name, func(t *testing.T) {
			objFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
			pointFrame := radialCFrameOriginal(t, unsafe.Sizeof(types.Pointf{}))
			argsFrame := radialCFrameOriginal(t, argsSize)
			probeFrame := radialCFrameOriginal(t, probeSize)
			obj := (*server.Object)(unsafe.Pointer(&objFrame[32]))
			point := (*types.Pointf)(unsafe.Pointer(&pointFrame[32]))
			data := unsafe.Pointer(&argsFrame[32])
			probe := unsafe.Pointer(&probeFrame[32])
			d := i.Objects[0]
			*obj = server.Object{PosVec: rigidNativePoint([2]uint32{d[0], d[1]}), ForceVec: rigidNativePoint([2]uint32{d[2], d[3]}), Mass: math.Float32frombits(d[4]), ObjClass: object.Class(d[5]), ObjFlags: object.Flags(d[6]), Field37: 0x55667788, Field38: 0x11223344}
			*point = rigidNativePoint(i.Origin)
			radialPushCProbeInit52E0E0(probe)
			radialPushCInit52E0E0(data, point, math.Float32frombits(i.Outer), math.Float32frombits(i.Inner), math.Float32frombits(i.Force), nil, callback, uintptr(probe))
			wantObj, wantPoint, wantArgs, wantProbe := bytes.Clone(objFrame), bytes.Clone(pointFrame), bytes.Clone(argsFrame), bytes.Clone(probeFrame)
			binary.LittleEndian.PutUint32(wantObj[32+unsafe.Offsetof(obj.ForceVec):], row.Objects[0][2])
			binary.LittleEndian.PutUint32(wantObj[36+unsafe.Offsetof(obj.ForceVec):], row.Objects[0][3])
			word := func(p []byte, off uintptr, value uintptr) {
				if unsafe.Sizeof(uintptr(0)) == 8 {
					binary.LittleEndian.PutUint64(p[32+off:], uint64(value))
				} else {
					binary.LittleEndian.PutUint32(p[32+off:], uint32(value))
				}
			}
			called := false
			for _, e := range row.Events {
				if e.Service == "wake:0" {
					t.Fatal("C subset unexpectedly needs activation")
				}
				if e.Service == "before:0" {
					called = true
					if e.Args[2] != 0x12345678 {
						t.Fatal("opaque callback payload")
					}
					binary.LittleEndian.PutUint32(wantProbe[32+offsets[0]:], 1)
					word(wantProbe, offsets[1], uintptr(unsafe.Pointer(obj)))
					binary.LittleEndian.PutUint32(wantProbe[32+offsets[2]:], e.Args[1])
					word(wantProbe, offsets[3], uintptr(probe))
				}
			}
			radialPushCCandidate52E0E0(obj, data)
			if !bytes.Equal(objFrame, wantObj) || !bytes.Equal(pointFrame, wantPoint) || !bytes.Equal(argsFrame, wantArgs) || !bytes.Equal(probeFrame, wantProbe) {
				t.Fatalf("native C frames differ: object=%t origin=%t args=%t probe=%t expected callback=%t", bytes.Equal(objFrame, wantObj), bytes.Equal(pointFrame, wantPoint), bytes.Equal(argsFrame, wantArgs), bytes.Equal(probeFrame, wantProbe), called)
			}
		})
	}
	if count != 32 {
		t.Fatalf("finite missile callback conditions=%d", count)
	}
}
