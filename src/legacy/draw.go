package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME5_2.h"
#include "common__system__team.h"
#include "client__gui__guiquit.h"
#include "client__draw__debugdraw.h"
#include "client__draw__fx.h"
*/
import "C"
import (
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

var (
	Nox_xxx_cliUpdateCameraPos_435600    func(x, y int)
	Sub_437260                           func()
	Get_nox_client_texturedFloors_154956 func() bool
	Sub_480860                           func(dst, src []uint16, w int, a4p, a5p []uint32)
	Sub_473970                           func(a1 image.Point) image.Point
	Nox_client_isConnected               func() bool
	Nox_video_inFadeTransition_44E0D0    func() int
)

var (
	_ = [1]struct{}{}[unsafe.Sizeof(C.nox_render_data_t{})-unsafe.Sizeof(noxrender.RenderData{})]
	_ = [1]struct{}{}[unsafe.Sizeof(C.nox_render_mat_t{})-unsafe.Sizeof(noxrender.RenderMat{})]
	_ = [1]struct{}{}[unsafe.Sizeof(C.nox_draw_viewport_t{})-unsafe.Sizeof(noxrender.Viewport{})]
)

type nox_draw_viewport_t = C.nox_draw_viewport_t

func asViewportP(p unsafe.Pointer) *noxrender.Viewport {
	return (*noxrender.Viewport)(p)
}

func asViewport(p *nox_draw_viewport_t) *noxrender.Viewport {
	return asViewportP(unsafe.Pointer(p))
}

//export get_nox_client_texturedFloors_154956
func get_nox_client_texturedFloors_154956() C.bool {
	return C.bool(Get_nox_client_texturedFloors_154956())
}

//export sub_4C42A0
func sub_4C42A0(a1 *C.int2, a2 *C.int2, a3_cgo *int32, a4_cgo *int32) int32 {
	a3, a3_cgo_finish := cgoABIIntPtr(a3_cgo)
	defer a3_cgo_finish()
	a4, a4_cgo_finish := cgoABIIntPtr(a4_cgo)
	defer a4_cgo_finish()
	return GetClient().Sub4C42A0(AsPoint(unsafe.Pointer(a1)), AsPoint(unsafe.Pointer(a2)), a3, a4)
}

//export sub_4C5630
func sub_4C5630(a1_cgo int32, a2_cgo int32, a3_cgo int32) int32 {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	return int32(GetClient().Sub4C5630(a1, a2, a3))
}

//export nox_draw_getViewport_437250
func nox_draw_getViewport_437250() *nox_draw_viewport_t {
	return (*nox_draw_viewport_t)(GetClient().Viewport().C())
}

//export nox_xxx_getSomeCoods_435670
func nox_xxx_getSomeCoods_435670(a1 *C.int2) {
	p := GetClient().Viewport().World.Max
	a1.field_0 = C.int(p.X)
	a1.field_4 = C.int(p.Y)
}

//export nox_xxx_cliUpdateCameraPos_435600
func nox_xxx_cliUpdateCameraPos_435600(x_cgo, y_cgo int32) {
	x := int(x_cgo)
	y := int(y_cgo)
	Nox_xxx_cliUpdateCameraPos_435600(x, y)
}

//export sub_437260
func sub_437260() {
	Sub_437260()
}

//export nox_draw_splitColor_435280
func nox_draw_splitColor_435280(cl C.short, pr, pg, pb *C.uchar) {
	c := noxrender.SplitColor(noxcolor.RGBA5551(cl))
	*pr = C.uchar(c.R)
	*pg = C.uchar(c.G)
	*pb = C.uchar(c.B)
}

//export nox_draw_setMaterial_4340A0
func nox_draw_setMaterial_4340A0(ind_cgo, r_cgo, g_cgo, b_cgo int32) {
	ind := int(ind_cgo)
	r := int(r_cgo)
	g := int(g_cgo)
	b := int(b_cgo)
	GetClient().R2().Data().SetMaterialRGB(ind, r, g, b)
}

//export nox_draw_setMaterial_4341D0
func nox_draw_setMaterial_4341D0(ind_cgo, cl_cgo int32) {
	ind := int(ind_cgo)
	cl := int(cl_cgo)
	GetClient().R2().Data().SetMaterial(ind, noxcolor.RGBA5551(cl))
}

//export sub_434080
func sub_434080(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetField262(a1)
}

//export nox_xxx_drawSetTextColor_434390
func nox_xxx_drawSetTextColor_434390(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetTextColor(noxcolor.RGBA5551(a1))
}

//export nox_xxx_drawSetColor_4343E0
func nox_xxx_drawSetColor_4343E0(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetColor(noxcolor.RGBA5551(a1))
}

//export nox_client_drawSetColor_434460
func nox_client_drawSetColor_434460(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetColor2(noxcolor.RGBA5551(a1))
}

//export nox_client_drawEnableAlpha_434560
func nox_client_drawEnableAlpha_434560(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetAlphaEnabled(a1 != 0)
}

//export sub_4345F0
func sub_4345F0(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetMultiply14(a1)
}

//export nox_xxx_draw_434600
func nox_xxx_draw_434600(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Data().SetColorize17(a1)
}

//export sub_434990
func sub_434990(r_cgo, g_cgo, b_cgo int32) {
	r := int(r_cgo)
	g := int(g_cgo)
	b := int(b_cgo)
	GetClient().R2().Data().SetLightColor(noxrender.RGB{
		R: r,
		G: g,
		B: b,
	})
}

//export sub_4349C0
func sub_4349C0(a1 *C.uint) {
	arr := unsafe.Slice(a1, 3)
	GetClient().R2().Data().SetLightColor(noxrender.RGB{
		R: int(arr[0]),
		G: int(arr[1]),
		B: int(arr[2]),
	})
}

//export sub_47D370
func sub_47D370(a1_cgo int32) {
	a1 := int(a1_cgo)
	GetClient().R2().Set_dword_5d4594_3799484(a1)
}

//export sub_47D400
func sub_47D400(a1_cgo int32, a2 C.char) {
	a1 := int(a1_cgo)
	GetClient().R2().SetInterlacing(a1 != 0, int(a2))
}

//export sub_49F7C0_def
func sub_49F7C0_def() {
	GetClient().R2().Sub_49F7C0_def_go()
}

//export nox_client_drawSetAlpha_434580
func nox_client_drawSetAlpha_434580(a C.uchar) {
	GetClient().R2().Data().SetAlpha(byte(a))
}

//export nox_draw_enableTextSmoothing_43F670
func nox_draw_enableTextSmoothing_43F670(v_cgo int32) {
	v := int(v_cgo)
	GetClient().R2().SetTextSmooting(v != 0)
}

//export nox_client_drawResetPoints_49F5A0
func nox_client_drawResetPoints_49F5A0() {
	GetClient().R2().ClearPoints()
}

//export nox_client_drawAddPoint_49F500
func nox_client_drawAddPoint_49F500(x_cgo, y_cgo int32) {
	x := int(x_cgo)
	y := int(y_cgo)
	GetClient().R2().AddPoint(image.Pt(x, y))
}

//export nox_xxx_rasterPointRel_49F570
func nox_xxx_rasterPointRel_49F570(x_cgo, y_cgo int32) {
	x := int(x_cgo)
	y := int(y_cgo)
	GetClient().R2().AddPointRel(image.Pt(x, y))
}

//export nox_client_drawLineFromPoints_49E4B0
func nox_client_drawLineFromPoints_49E4B0() int32 {
	r := GetClient().R2()
	return int32(bool2int(r.DrawLineFromPoints(r.Data().Color2())))
}

//export sub_49E4F0
func sub_49E4F0(a1_cgo int32) int32 {
	a1 := int(a1_cgo)
	return int32(bool2int(GetClient().R2().DrawParticles49ED80(a1)))
}

//export sub_480860
func sub_480860(a1, a2 *C.ushort, w_cgo int32, a4, a5 *C.int) {
	w := int(w_cgo)
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(a1)), w)
	src := unsafe.Slice((*uint16)(unsafe.Pointer(a2)), w)
	a4p := unsafe.Slice((*uint32)(unsafe.Pointer(a4)), 3)
	a5p := unsafe.Slice((*uint32)(unsafe.Pointer(a5)), 3)
	Sub_480860(dst, src, w, a4p, a5p)
}

//export nox_draw_setColorMultAndIntensityRGB_433CD0
func nox_draw_setColorMultAndIntensityRGB_433CD0(r, g, b C.uchar) int32 {
	return int32(int(GetClient().R2().SetColorMultAndIntensityRGB(byte(r), byte(g), byte(b))))
}

//export nox_draw_set54RGB32_434040
func nox_draw_set54RGB32_434040(cl_cgo int32) {
	cl := int(cl_cgo)
	c := noxrender.SplitColor(noxcolor.RGBA5551(cl))
	GetClient().R2().Data().SetColorInt54(noxrender.RGB{
		R: int(c.R),
		G: int(c.G),
		B: int(c.B),
	})
}

//export nox_draw_setColorMultAndIntensity_433E40
func nox_draw_setColorMultAndIntensity_433E40(cl_cgo int32) int32 {
	cl := int(cl_cgo)
	c := noxcolor.RGBA5551(cl).ColorNRGBA()
	return int32(int(GetClient().R2().SetColorMultAndIntensityRGB(c.R, c.G, c.B)))
}

//export sub_437290
func sub_437290() {
	GetClient().R2().SetRectFullScreen()
}

//export nox_client_drawRectFilledOpaque_49CE30
func nox_client_drawRectFilledOpaque_49CE30(a1_cgo, a2_cgo, a3_cgo, a4_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	r := GetClient().R2()
	r.DrawRectFilledOpaque(a1, a2, a3, a4, r.Data().Color2())
}

//export nox_client_drawRectFilledAlpha_49CF10
func nox_client_drawRectFilledAlpha_49CF10(a1_cgo, a2_cgo, a3_cgo, a4_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	GetClient().R2().DrawRectFilledAlpha(a1, a2, a3, a4)
}

//export nox_client_drawBorderLines_49CC70
func nox_client_drawBorderLines_49CC70(a1_cgo, a2_cgo, a3_cgo, a4_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	r := GetClient().R2()
	r.DrawBorder(a1, a2, a3, a4, r.Data().Color2())
}

//export nox_client_drawPixel_49EFA0
func nox_client_drawPixel_49EFA0(a1_cgo, a2_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	r := GetClient().R2()
	r.DrawPixel(image.Pt(a1, a2), r.Data().Color2())
}

//export nox_client_drawPoint_4B0BC0
func nox_client_drawPoint_4B0BC0(a1_cgo, a2_cgo, a3_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	r := GetClient().R2()
	r.DrawPointRad(image.Pt(a1, a2), a3, r.Data().Color2())
}

//export nox_xxx_drawPointMB_499B70
func nox_xxx_drawPointMB_499B70(a1_cgo, a2_cgo, a3_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	r := GetClient().R2()
	r.DrawPoint(image.Pt(a1, a2), a3, r.Data().Color2())
}

//export nox_xxx_guiFontHeightMB_43F320
func nox_xxx_guiFontHeightMB_43F320(fnt unsafe.Pointer) int32 {
	r := GetClient().R2()
	return int32(r.FontHeight(r.GetFonts().AsFont(fnt)))
}

//export nox_draw_setTabWidth_43FE20
func nox_draw_setTabWidth_43FE20(v_cgo int32) int32 {
	v := int(v_cgo)
	old := GetClient().R2().TabWidth()
	GetClient().R2().SetTabWidth(v)
	return int32(old)
}

//export nox_xxx_drawGetStringSize_43F840
func nox_xxx_drawGetStringSize_43F840(font unsafe.Pointer, sp *wchar2_t, outW, outH *C.int, maxW_cgo int32) int32 {
	maxW := int(maxW_cgo)
	r := GetClient().R2()
	sz := r.GetStringSizeWrapped(r.GetFonts().AsFont(font), GoWString(sp), maxW)
	if outW != nil {
		*outW = C.int(sz.X)
	}
	if outH != nil {
		*outH = C.int(sz.Y)
	}
	return int32(bool2int(sz != (image.Point{})))
}

//export nox_xxx_bookGetStringSize_43FA80
func nox_xxx_bookGetStringSize_43FA80(font unsafe.Pointer, sp *wchar2_t, outW, outH *C.int, maxW_cgo int32) int32 {
	maxW := int(maxW_cgo)
	r := GetClient().R2()
	sz := r.GetStringSizeWrappedStyle(r.GetFonts().AsFont(font), GoWString(sp), maxW)
	if outW != nil {
		*outW = C.int(sz.X)
	}
	if outH != nil {
		*outH = C.int(sz.Y)
	}
	return int32(bool2int(sz != (image.Point{})))
}

//export nox_xxx_drawString_43F6E0
func nox_xxx_drawString_43F6E0(font unsafe.Pointer, sp *wchar2_t, x_cgo, y_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	r := GetClient().R2()
	return int32(r.DrawString(r.GetFonts().AsFont(font), GoWString(sp), image.Point{X: x, Y: y}))
}

//export nox_draw_drawStringHL_43F730
func nox_draw_drawStringHL_43F730(font unsafe.Pointer, sp *wchar2_t, x_cgo, y_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	r := GetClient().R2()
	return int32(r.DrawStringHL(r.GetFonts().AsFont(font), GoWString(sp), image.Point{X: x, Y: y}))
}

//export nox_xxx_drawStringWrap_43FAF0
func nox_xxx_drawStringWrap_43FAF0(font unsafe.Pointer, sp *wchar2_t, x_cgo, y_cgo, maxW_cgo, maxH_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	maxW := int(maxW_cgo)
	maxH := int(maxH_cgo)
	r := GetClient().R2()
	return int32(r.DrawStringWrapped(r.GetFonts().AsFont(font), GoWString(sp), image.Rect(x, y, x+maxW, y+maxH)))
}

//export nox_xxx_drawStringWrapHL_43FD00
func nox_xxx_drawStringWrapHL_43FD00(font unsafe.Pointer, sp *wchar2_t, x_cgo, y_cgo, maxW_cgo, maxH_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	maxW := int(maxW_cgo)
	maxH := int(maxH_cgo)
	r := GetClient().R2()
	return int32(r.DrawStringWrappedHL(r.GetFonts().AsFont(font), GoWString(sp), image.Rect(x, y, x+maxW, y+maxH)))
}

//export nox_xxx_bookDrawString_43FA80_43FD80
func nox_xxx_bookDrawString_43FA80_43FD80(font unsafe.Pointer, s *wchar2_t, x_cgo, y_cgo, maxW_cgo, maxH_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	maxW := int(maxW_cgo)
	maxH := int(maxH_cgo)
	r := GetClient().R2()
	return int32(r.DrawStringWrappedStyle(r.GetFonts().AsFont(font), GoWString(s), image.Rect(x, y, x+maxW, y+maxH)))
}

//export nox_xxx_drawStringStyle_43F7B0
func nox_xxx_drawStringStyle_43F7B0(font unsafe.Pointer, sp *wchar2_t, x_cgo, y_cgo int32) int32 {
	x := int(x_cgo)
	y := int(y_cgo)
	r := GetClient().R2()
	return int32(r.DrawStringStyle(r.GetFonts().AsFont(font), GoWString(sp), image.Point{X: x, Y: y}))
}

//export nox_video_drawAnimatedImageOrCursorAt_4BE6D0
func nox_video_drawAnimatedImageOrCursorAt_4BE6D0(a1 uintptr, a2_cgo, a3_cgo int32) {
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	GetClient().Nox_video_drawAnimatedImageOrCursorAt(AsImageRefP(unsafe.Pointer(a1)), image.Point{X: a2, Y: a3})
}

//export sub_484C60
func sub_484C60(a1 C.float) int32 {
	return int32(client.LightRadius(float32(a1)))
}

//export sub_469920
func sub_469920(p *C.nox_point) *C.char {
	dst := GetClient().Sub469920(AsPoint(unsafe.Pointer(p)))
	return (*C.char)(unsafe.Pointer(&dst[0]))
}

//export nox_video_drawCircleColored_4C3270
func nox_video_drawCircleColored_4C3270(a1_cgo, a2_cgo, a3_cgo, a4_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	a4 := int(a4_cgo)
	GetClient().R2().DrawCircle(a1, a2, a3, noxcolor.RGBA5551(a4))
}

//export nox_video_drawCircle_4B0B90
func nox_video_drawCircle_4B0B90(a1_cgo, a2_cgo, a3_cgo int32) {
	a1 := int(a1_cgo)
	a2 := int(a2_cgo)
	a3 := int(a3_cgo)
	GetClient().R2().DrawCircle(a1, a2, a3, GetClient().R2().Data().Color2())
}

//export nox_client_drawImageAt_47D2C0
func nox_client_drawImageAt_47D2C0(img *nox_video_bag_image_t, x_cgo, y_cgo int32) {
	x := int(x_cgo)
	y := int(y_cgo)
	GetClient().R2().DrawImageAt(asImage(img), image.Point{X: x, Y: y})
}

//export nox_draw_imageMeta_47D5C0
func nox_draw_imageMeta_47D5C0(img *nox_video_bag_image_t, px, py, pw, ph *C.uint) int32 {
	if img == nil {
		return int32(0)
	}
	if pw != nil {
		*pw = 0
	}
	if ph != nil {
		*ph = 0
	}
	off, sz, ok := asImage(img).Meta()
	if !ok {
		return int32(0)
	}
	if px != nil {
		*px += C.uint(off.X)
	}
	if py != nil {
		*py += C.uint(off.Y)
	}
	if pw != nil {
		*pw = C.uint(sz.X)
	}
	if ph != nil {
		*ph = C.uint(sz.Y)
	}
	return int32(1)
}

//export nox_video_getImagePixdata_42FB30
func nox_video_getImagePixdata_42FB30(img *nox_video_bag_image_t) unsafe.Pointer {
	data := asImage(img).Pixdata()
	if len(data) == 0 {
		return nil
	}
	return unsafe.Pointer(&data[0])
}

//export sub_4AE6F0
func sub_4AE6F0(cx_cgo, cy_cgo, rad_cgo, ang_cgo, ccl_cgo int32) {
	cx := int(cx_cgo)
	cy := int(cy_cgo)
	rad := int(rad_cgo)
	ang := int(ang_cgo)
	ccl := int(ccl_cgo)
	GetClient().R2().DrawCircleSegment(cx, cy, rad, ang, noxcolor.RGBA5551(ccl))
}

//export sub_473970
func sub_473970(a1, a2p *C.int2) {
	a2 := Sub_473970(image.Pt(int(a1.field_0), int(a1.field_4)))
	a2p.field_0 = C.int(a2.X)
	a2p.field_4 = C.int(a2.Y)
}

//export nox_client_isConnected_43C700
func nox_client_isConnected_43C700() int32 {
	return int32(bool2int(Nox_client_isConnected()))
}

//export nox_video_stopAllFades_44E040
func nox_video_stopAllFades_44E040() {
	GetClient().Nox_video_stopAllFades44E040()
}

//export nox_video_inFadeTransition_44E0D0
func nox_video_inFadeTransition_44E0D0() int32 {
	return int32(Nox_video_inFadeTransition_44E0D0())
}

//export nox_video_fadeInScreen_44DAB0
func nox_video_fadeInScreen_44DAB0(a1, a2 C.int, fnc unsafe.Pointer) {
	GetClient().R2().FadeInScreen(int(a1), a2 != 0, func() {
		ccall.CallVoidVoid(fnc)
	})
}

//export nox_video_fadeOutScreen_44DB30
func nox_video_fadeOutScreen_44DB30(a1, a2 C.int, fnc unsafe.Pointer) {
	GetClient().R2().FadeOutScreen(int(a1), a2 != 0, func() {
		ccall.CallVoidVoid(fnc)
	})
}

//export sub_4B6720
func sub_4B6720(a1 *C.int2, a2, a3 C.int, a4 C.char) {
	GetClient().R2().DrawGlow(AsPoint(unsafe.Pointer(a1)), noxcolor.RGBA5551(a2), int(a3), int(a4))
}

func toRect(cr *C.nox_rect) image.Rectangle {
	return image.Rect(int(cr.min_x), int(cr.min_y), int(cr.max_x), int(cr.max_y))
}

func setRect(cr *C.nox_rect, r image.Rectangle) {
	cr.min_x = C.intptr_t(r.Min.X)
	cr.min_y = C.intptr_t(r.Min.Y)
	cr.max_x = C.intptr_t(r.Max.X)
	cr.max_y = C.intptr_t(r.Max.Y)
}

func Sub_437180() {
	C.sub_48D990((*nox_draw_viewport_t)(GetClient().Viewport().C()))
}

func Sub_476AE0(vp *noxrender.Viewport, dr *client.Drawable) {
	C.sub_476AE0((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))
}

func Nox_xxx_drawShinySpot_4C4F40(vp *noxrender.Viewport, dr *client.Drawable) {
	C.nox_xxx_drawShinySpot_4C4F40((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))
}

func Sub_499F60(p int, pos image.Point, a4 int, a5, a6, a7, a8, a9 int, a10 int) {
	C.sub_499F60(C.int(p), C.int(pos.X), C.int(pos.Y), C.short(a4), C.char(a5), C.char(a6), C.char(a7), C.char(a8), C.char(a9), C.int(a10))
}

func Get_sub_480250() unsafe.Pointer {
	return C.sub_480250
}
func Get_sub_480220() unsafe.Pointer {
	return C.sub_480220
}
func Get_nox_xxx_tileDraw_4815E0() unsafe.Pointer {
	return C.nox_xxx_tileDraw_4815E0
}
func Get_nox_xxx_drawTexEdgesProbably_481900() unsafe.Pointer {
	return C.nox_xxx_drawTexEdgesProbably_481900
}
func Get_sub_481770() unsafe.Pointer {
	return C.sub_481770
}
func Get_nullsub_8() unsafe.Pointer {
	return C.nullsub_8
}
func Sub_435120(a1 unsafe.Pointer, a2 unsafe.Pointer) {
	C.sub_435120(a1, a2)
}
func Sub_435040() {
	C.sub_435040()
}
func Sub_435150(a1 unsafe.Pointer, a2 unsafe.Pointer) {
	C.sub_435150((*C.uchar)(a1), (*C.char)(a2))
}
func Nox_xxx_wndDraw_49F7F0() {
	C.nox_xxx_wndDraw_49F7F0()
}
func Sub_49F780(a1 int, a2 int) {
	C.sub_49F780(C.int(a1), C.int(a2))
}
func Sub_49F860() {
	C.sub_49F860()
}
func Nox_xxx_drawEnergyBolt_499710(a1 int, a2 int, a3 int, a4 int) {
	C.nox_xxx_drawEnergyBolt_499710(C.int(a1), C.int(a2), C.short(a3), C.int(a4))
}
func Nox_xxx_drawShield_499810(vp *noxrender.Viewport, dr *client.Drawable) {
	C.nox_xxx_drawShield_499810((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))
}
func Sub_474B40(dr *client.Drawable) int {
	return int(C.sub_474B40((*nox_drawable)(dr.C())))
}
func Sub_495BB0(dr *client.Drawable, vp *noxrender.Viewport) {
	C.sub_495BB0((*nox_drawable)(dr.C()), (*nox_draw_viewport_t)(vp.C()))
}
