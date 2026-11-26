package legacy

/*
#include "GAME2_3.h"
#include "cgo_const.h"
#include "client__draw__debugdraw.h"

*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/binfile"
)

func init() {
	client.RegisterThingParse("LIGHTDIRECTION", wrapClientThingFunc(nox_parse_thing_light_dir))
	client.RegisterThingParse("LIGHTPENUMBRA", wrapClientThingFunc(nox_parse_thing_light_penumbra))
	client.RegisterThingParse("CLIENTUPDATE", wrapClientThingFunc(nox_parse_thing_client_update))
	client.RegisterThingParse("PRETTYIMAGE", wrapClientThingFunc(nox_parse_thing_pretty_image))
	client.ThingDrawDefault = C.nox_thing_debug_draw
}

type nox_thing = C.nox_thing

//export nox_xxx_getTTByNameSpriteMB_44CFC0
func nox_xxx_getTTByNameSpriteMB_44CFC0(cstr *C.nox_cgo_const_char) int32 {
	id := GoString(cstr)
	return int32(GetClient().Cli().Things.IndByID(id))
}

//export sub_44D330
func sub_44D330(cstr *C.char) *nox_thing {
	id := GoString(cstr)
	return (*C.nox_thing)(GetClient().Cli().Things.TypeByID(id).C())
}

//export nox_get_thing_name
func nox_get_thing_name(i_cgo int32) *C.char {
	i := int(i_cgo)
	t := GetClient().Cli().Things.TypeByInd(i)
	if t == nil {
		return nil
	}
	return (*C.char)(unsafe.Pointer(t.Name))
}

//export nox_get_thing
func nox_get_thing(i_cgo int32) *nox_thing {
	i := int(i_cgo)
	t := GetClient().Cli().Things.TypeByInd(i)
	if t == nil {
		return nil
	}
	return (*nox_thing)(t.C())
}

//export nox_get_thing_pretty_name
func nox_get_thing_pretty_name(i_cgo int32) *wchar2_t {
	i := int(i_cgo)
	t := GetClient().Cli().Things.TypeByInd(i)
	if t == nil {
		return nil
	}
	return (*wchar2_t)(unsafe.Pointer(t.PrettyName))
}

//export nox_get_thing_desc
func nox_get_thing_desc(i_cgo int32) *wchar2_t {
	i := int(i_cgo)
	t := GetClient().Cli().Things.TypeByInd(i)
	if t == nil {
		return nil
	}
	return (*wchar2_t)(unsafe.Pointer(t.Desc))
}

//export nox_get_thing_pretty_image
func nox_get_thing_pretty_image(i_cgo int32) int32 {
	i := int(i_cgo)
	t := GetClient().Cli().Things.TypeByInd(i)
	if t == nil {
		return int32(0)
	}
	return int32(int(t.PrettyImage))
}

//export nox_drawable_link_thing
func nox_drawable_link_thing(a1c *nox_drawable, i_cgo int32) int32 {
	i := int(i_cgo)
	return int32(GetClient().Cli().DrawableLinkThing(asDrawable(a1c), i))
}

func wrapClientThingFunc(parse func(*nox_thing, *nox_memfile, *C.char) C.bool) client.ThingFieldFunc {
	return func(typ *client.ObjectType, f *binfile.MemFile, str string, buf []byte) error {
		StrNCopyBytes(buf, str)
		if !parse((*nox_thing)(typ.C()), (*nox_memfile)(f.C()), (*C.char)(unsafe.Pointer(&buf[0]))) {
			return fmt.Errorf("failed to parse %q", str)
		}
		return nil
	}
}
