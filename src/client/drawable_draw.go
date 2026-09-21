package client

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
)

func init() {
	RegisterThingParse("DRAW", parseThingDraw)

	RegisterDraw("NoDraw", nil, 0, nil)
}

type thingsDraw struct {
	Name  string
	Draw  unsafe.Pointer
	Kind  int
	Parse ThingFieldFunc
}

var (
	drawFuncs    = make(map[string]*thingsDraw)
	drawFuncPtrs = make(map[unsafe.Pointer]struct{})
)

func RegisterDraw(name string, draw unsafe.Pointer, kind int, parse ThingFieldFunc) {
	if _, ok := drawFuncs[name]; ok {
		panic("already registered")
	}
	drawFuncs[name] = &thingsDraw{Name: name, Draw: draw, Kind: kind, Parse: parse}
	drawFuncPtrs[draw] = struct{}{}
}

// IsRegisteredDrawFunc reports whether fn belongs to the DRAW callback
// registry. Drawable callback slots may be damaged by legacy PE32 offset
// writes on native-width builds; callers must not invoke an arbitrary value
// merely because it is non-nil.
func IsRegisteredDrawFunc(fn unsafe.Pointer) bool {
	_, ok := drawFuncPtrs[fn]
	return ok
}

func DrawableDataKind(fnc unsafe.Pointer) int {
	if fnc == nil {
		return 0
	}
	for _, v := range drawFuncs {
		if v.Draw == fnc {
			return v.Kind
		}
	}
	return 0
}

func parseThingDraw(obj *ObjectType, f *binfile.MemFile, str string, buf []byte) error {
	name, _ := f.ReadString8()
	// TODO: After cleanup: Figure out if this value has any significance to the data in the file, or if the file was
	_ = f.ReadU64Align()
	item := drawFuncs[name]
	if item == nil {
		thingsLog.Printf("unsupported draw function: %q", name)
		return nil
	}
	if item.Parse != nil {
		if err := item.Parse(obj, f, str, buf); err != nil {
			thingsLog.Printf("failed to parse draw %q: %s", name, err)
		}
	}
	obj.DrawFunc = item.Draw
	return nil
}
