package noxrender

import (
	"encoding/binary"
	"image"
	"unsafe"
)

type drawOp16Func func(dst []uint16, src []byte, val int) (outDst []uint16, outSrc []byte)
type drawOp8Func func(dst []uint16, src []byte, op byte, val int) (outDst []uint16, outSrc []byte)

type drawOps struct {
	draw27 drawOp16Func
	draw4  drawOp8Func
	draw5  drawOp16Func
	draw6  drawOp16Func
}

type Image16 interface {
	Type() int
	Pixdata() []byte
}

type rawImage struct {
	typ  int
	data []byte
}

func (img *rawImage) Type() int {
	return img.typ
}

func (img *rawImage) Pixdata() []byte {
	return img.data
}

func NewRawImage16(typ int, data []byte) Image16 {
	return &rawImage{typ: typ, data: data}
}

func (r *NoxRender) DrawImage16(img Image16, pos image.Point) {
	if img == nil {
		return
	}
	if p, ok := img.(*Image); ok && p == nil {
		return
	}
	switch img.Type() & 0x3F {
	case 2, 7:
		r.nox_client_drawImg_bbb_4C7860(img, pos)
	case 3, 4, 5, 6:
		var ops drawOps
		ops.draw5 = r.pixOpOver4444
		ops.draw6 = func(dst []uint16, src []byte, val int) (_ []uint16, _ []byte) { return dst, src }
		if !r.p.IsAlphaEnabled() {
			if r.p.Multiply14() {
				ops.draw5 = r.pixOpOver4444Multiply
				ops.draw27 = r.pixOpSrcMultiply
				ops.draw4 = r.pixOpSrcMultiplyIndexed
			} else {
				ops.draw27 = pixOpSrc
				ops.draw4 = r.pixOpSrcIndexed
				if r.p.Colorize17() {
					ops.draw27 = r.pixOpSrcColorize
				}
			}
		} else {
			ops.draw5 = r.pixOpOver4444Alpha
			alpha := r.Data().Alpha()
			if r.p.Multiply14() {
				if alpha == 0xFF {
					if r.p.Flag16() {
						ops.draw27 = pixOpSrc
						ops.draw4 = r.pixOpSrcIndexed
					} else {
						ops.draw5 = r.pixOpOver4444Multiply
						ops.draw27 = r.pixOpSrcMultiply
						ops.draw4 = r.pixOpSrcMultiplyIndexed
					}
				} else if alpha == 0x80 {
					ops.draw27 = r.pixOpOverMultiplyAlpha50
					ops.draw4 = r.pixOpOverMultiplyAlpha50Indexed
				} else {
					ops.draw27 = r.pixOpOverMultiplyAlpha
					ops.draw4 = r.pixOpOverMultiplyAlphaIndexed
				}
			} else {
				if alpha == 0xFF {
					ops.draw27 = pixOpSrc
					ops.draw4 = r.pixOpSrcIndexed
				} else if alpha == 0x80 {
					ops.draw27 = r.pixOpOverAlpha50
					ops.draw4 = r.pixOpOverAlpha50Indexed
				} else {
					ops.draw27 = r.pixOpOverAlpha
					ops.draw4 = r.pixOpOverAlphaIndexed
				}
			}
		}
		r.nox_client_drawImg_aaa_4C79F0(&ops, img, pos)
	case 8:
		var ops drawOps
		ops.draw27 = r.pixBlendPremult
		r.nox_client_drawImg_aaa_4C79F0(&ops, img, pos)
	}
}

func (r *NoxRender) nox_client_drawImg_bbb_4C7860(img Image16, pos image.Point) {
	data := img.Pixdata()
	if len(data) < 17 {
		return
	}
	width := int(int32(binary.LittleEndian.Uint32(data[0:])))
	height := int(int32(binary.LittleEndian.Uint32(data[4:])))
	if width <= 0 || height <= 0 {
		return
	}
	data = data[8:]

	offX := int32(binary.LittleEndian.Uint32(data[0:]))
	offY := int32(binary.LittleEndian.Uint32(data[4:]))
	data = data[8:]
	pos.X += int(offX)
	pos.Y += int(offY)

	data = data[1:] // unused

	if r.dword_5d4594_3799484 != 0 {
		crop := int(r.dword_5d4594_3799484)
		if crop >= height {
			return
		}
		height -= crop
		r.dword_5d4594_3799476 = pos.Y + int(height)
	}

	wsz := width
	if r.p.Clip() {
		rc := image.Rectangle{Min: pos, Max: pos.Add(image.Pt(width, height))}
		a1a := rc.Intersect(r.p.ClipRect())
		if a1a.Empty() {
			return
		}
		v11 := a1a.Min.X - rc.Min.X
		v12 := a1a.Min.Y - rc.Min.Y
		wsz = a1a.Dx()
		height = a1a.Dy()
		if a1a.Min.X != rc.Min.X || v12 != 0 {
			pos.X += v11
			data = data[width*v12+2*v11:]
			pos.Y += v12
		}
	}
	xoff := pos.X
	ipitch := 2 * width
	pixbuf := r.PixBuffer()
	pitch := pixbuf.Stride
	for i := 0; i < int(height); i++ {
		dst := pixbuf.Pix[pitch*(pos.Y+1)+xoff:]
		src := data[ipitch*i:]
		copy16b(dst[:wsz], src[:wsz*2])
	}
}

func (r *NoxRender) nox_client_drawImg_aaa_4C79F0(ops *drawOps, img Image16, pos image.Point) {
	src := img.Pixdata()
	if len(src) < 17 {
		return
	}

	width := int(int32(binary.LittleEndian.Uint32(src[0:])))
	height := int(int32(binary.LittleEndian.Uint32(src[4:])))
	if width <= 0 || height <= 0 {
		return
	}
	src = src[8:]

	offX := int32(binary.LittleEndian.Uint32(src[0:]))
	offY := int32(binary.LittleEndian.Uint32(src[4:]))
	src = src[8:]
	pos.X += int(offX)
	pos.Y += int(offY)

	src = src[1:] // unused

	if r.dword_5d4594_3799484 != 0 {
		crop := int(r.dword_5d4594_3799484)
		if crop >= height {
			return
		}
		height -= crop
		r.dword_5d4594_3799476 = pos.Y + int(height)
	}
	if r.HookImageDrawXxx != nil {
		r.HookImageDrawXxx(pos, image.Point{X: width, Y: height})
	}
	if r.p.Clip() {
		rc := image.Rectangle{Min: pos, Max: pos.Add(image.Pt(width, height))}
		a1a := rc.Intersect(r.p.ClipRect())
		if a1a.Empty() {
			return
		}
		if rc != a1a {
			r.nox_client_drawXxx_4C7C80(ops, src, pos, width, a1a)
			return
		}
	}
	r.interlacingY ^= pos.Y & 0x1
	pixbuf := r.PixBuffer()
	pitch := pixbuf.Stride
	for i := 0; i < height; i++ {
		dst := pixbuf.Pix[pitch*(pos.Y+i)+pos.X:]
		if r.interlacing {
			r.interlacingY ^= 1
			if r.interlacingY != 0 {
				if i != 0 {
					copy(dst[:width], pixbuf.Pix[pitch*(pos.Y+i-1)+pos.X:])
				}
				var ok bool
				src, ok = skipPixdata(src, width, 1)
				if !ok {
					return
				}
				continue
			}
		}
		var val int
		for j := 0; j < width; j += val {
			run, _, ok := nextPixdataRun(src)
			if !ok || run.n > width-j || len(dst) < run.n {
				return
			}
			op := run.op
			val = run.n
			src = src[2:]

			if op&0xF == 1 {
				dst = dst[val:]
				continue
			}
			switch op & 0xF {
			case 2, 7:
				if ops.draw27 == nil {
					return
				}
				dst, src = ops.draw27(dst, src, val)
			case 4:
				if ops.draw4 == nil {
					return
				}
				dst, src = ops.draw4(dst, src, op>>4, val)
			case 5:
				if ops.draw5 == nil {
					return
				}
				dst, src = ops.draw5(dst, src, val)
			case 6:
				if ops.draw6 == nil {
					return
				}
				dst, src = ops.draw6(dst, src, val)
			default:
				return
			}
		}
	}
}

func (r *NoxRender) nox_client_drawXxx_4C7C80(ops *drawOps, pix []byte, pos image.Point, width int, clip image.Rectangle) {
	left := clip.Min.X
	right := clip.Max.X
	dy := clip.Min.Y - pos.Y
	height := clip.Dy()
	if r.dword_5d4594_3799484 != 0 {
		height -= int(r.dword_5d4594_3799484)
		if height <= 0 {
			return
		}
		r.dword_5d4594_3799476 = pos.Y + height
	}
	if height == 0 {
		return
	}
	ys := pos.Y
	if dy != 0 {
		ys += dy
		var ok bool
		pix, ok = skipPixdata(pix, width, dy)
		if !ok {
			return
		}
	}
	r.interlacingY ^= ys & 0x1
	pixbuf := r.PixBuffer()
	pitch := pixbuf.Stride
	for i := 0; i < height; i++ {
		yi := ys + i
		if r.interlacing {
			r.interlacingY ^= 1
			if r.interlacingY != 0 {
				if i != 0 {
					src := pixbuf.Pix[pitch*(yi-1)+left:]
					dst := pixbuf.Pix[pitch*(yi+0)+left:]
					w := right - left
					if w > width {
						w = width
					}
					copy(dst[:w], src[:w])
				}
				var ok bool
				pix, ok = skipPixdata(pix, width, 1)
				if !ok {
					return
				}
				continue
			}
		}
		if width <= 0 {
			continue
		}
		row := pixbuf.Pix[pitch*yi : pitch*(yi+1)]
		var n int
		for j := 0; j < width; j += n {
			if len(pix) < 2 {
				return
			}
			op := pix[0]
			n = int(pix[1])
			if n == 0 || n > width-j {
				return
			}
			pix = pix[2:]

			if op&0xF == 1 {
				continue
			}

			var (
				fnc16 drawOp16Func
				fnc8  drawOp8Func
				pmul  int
			)
			switch op & 0xF {
			case 2, 7:
				fnc16 = ops.draw27
				pmul = 2
			case 4:
				fnc8 = ops.draw4
				pmul = 1
			case 5:
				fnc16 = ops.draw5
				pmul = 2
			case 6:
				fnc16 = ops.draw6
				pmul = 2
			default:
				return
			}
			if fnc8 == nil && fnc16 == nil {
				return
			}
			xs := pos.X + j
			xe := xs + n
			xw := n
			if xe <= left || xs >= right {
				need := pmul * n
				if need > len(pix) {
					return
				}
				pix = pix[need:]
				continue
			}

			pix2 := pix
			if xs < left {
				d := left - xs
				xw -= d
				xs = left
				if pmul*d > len(pix2) {
					return
				}
				pix2 = pix2[pmul*d:]
			}
			if xe > right {
				d := xe - right
				xw -= d
			}
			row2 := row[xs:]
			if fnc8 != nil {
				if n > len(pix) {
					return
				}
				_, _ = fnc8(row2, pix2, op>>4, xw)
				pix = pix[n:]
			} else {
				if 2*n > len(pix) {
					return
				}
				_, _ = fnc16(row2, pix2, xw)
				pix = pix[2*n:]
			}
		}
	}
}

func copy16b(dst []uint16, src []byte) int {
	n16 := len(src) / 2
	if n16 == 0 {
		return 0
	}
	src = src[:n16*2]
	src16 := unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), n16)
	return copy(dst, src16)
}

type pixdataRun struct {
	op byte
	n  int
}

func nextPixdataRun(pix []byte) (pixdataRun, []byte, bool) {
	if len(pix) < 2 {
		return pixdataRun{}, nil, false
	}
	op := pix[0]
	n := int(pix[1])
	if n <= 0 {
		return pixdataRun{}, nil, false
	}
	var payload int
	switch op & 0xF {
	case 1:
	case 2, 5, 6, 7:
		payload = 2 * n
	case 4:
		payload = n
	default:
		return pixdataRun{}, nil, false
	}
	if payload > len(pix)-2 {
		return pixdataRun{}, nil, false
	}
	return pixdataRun{op: op, n: n}, pix[2+payload:], true
}

func skipPixdata(pix []byte, width int, skip int) ([]byte, bool) {
	if width <= 0 || skip < 0 {
		return nil, false
	}
	for i := 0; i < skip; i++ {
		for covered := 0; covered < width; {
			run, next, ok := nextPixdataRun(pix)
			if !ok || run.n > width-covered {
				return nil, false
			}
			pix = next
			covered += run.n
		}
	}
	return pix, true
}

type drawU16Func func(old uint16, src uint16) uint16

func (r *NoxRender) drawOpU16(dst []uint16, src []byte, sz int, fnc drawU16Func) (_ []uint16, _ []byte) {
	if sz < 0 {
		panic("negative size")
	}
	dnext := dst[sz:]
	snext := src[2*sz:]
	for i := 0; i < sz; i++ {
		c1 := dst[i]
		c2 := binary.LittleEndian.Uint16(src[2*i:])
		dst[i] = fnc(c1, c2)
	}
	return dnext, snext
}

type drawU8Func func(old uint16, src byte) uint16

func (r *NoxRender) drawOpU8(dst []uint16, src []byte, sz int, fnc drawU8Func) (_ []uint16, _ []byte) {
	if sz < 0 {
		panic("negative size")
	}
	dnext := dst[sz:]
	snext := src[sz:]
	for i := 0; i < sz; i++ {
		c1 := dst[i]
		c2 := src[i]
		dst[i] = fnc(c1, c2)
	}
	return dnext, snext
}

func pixOpSrc(dst []uint16, src []byte, n int) (_ []uint16, _ []byte) {
	if n < 0 {
		panic("negative size")
	}
	copy16b(dst[:n], src[:n*2])
	return dst[n:], src[n*2:]
}

func (r *NoxRender) pixOpOverMultiplyAlpha50(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()

	return r.drawOpU16(dst, src, sz, func(old uint16, src uint16) uint16 {
		c1 := SplitColor16(src)
		c2 := SplitColor16(old)
		return c1.Mult(mul).Over(c2).Make16()
	})
}

func (r *NoxRender) pixOpOver4444(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2, a := SplitColor4444(c2)
		return cc1.OverAlpha(a, cc2).Make16()
	})
}

func (r *NoxRender) pixOpOver4444Multiply(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()

	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2, a := SplitColor4444(c2)
		return cc1.OverAlpha(a, cc2.Mult(mul)).Make16()
	})
}

func (r *NoxRender) pixOpOver4444Alpha(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	a := uint16(r.p.Alpha())

	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2, a2 := SplitColor4444(c2)
		a2 = ((a * a2) >> 8) & 0xff
		return cc1.OverAlpha(a2, cc2).Make16()
	})
}

func (r *NoxRender) pixOpSrcMultiply(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()

	return r.drawOpU16(dst, src, sz, func(_ uint16, c2 uint16) uint16 {
		c := SplitColor16(c2)
		return c.Mult(mul).Make16()
	})
}

func (r *NoxRender) pixOpOverAlpha50(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2 := SplitColor16(c2)
		return cc1.Over(cc2).Make16()
	})
}

func (r *NoxRender) pixOpOverAlpha(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	a := uint16(r.p.Alpha())

	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2 := SplitColor16(c2)
		return cc1.OverAlpha(a, cc2).Make16()
	})
}

func (r *NoxRender) pixOpOverMultiplyAlpha(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	a := uint16(r.p.Alpha())
	mul := r.p.ColorMultA()

	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc1 := SplitColor16(c1)
		cc2 := SplitColor16(c2)
		return cc1.OverAlpha(a, cc2.Mult(mul)).Make16()
	})
}

func (r *NoxRender) pixOpSrcColorize(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()

	return r.drawOpU16(dst, src, sz, func(_ uint16, ci uint16) uint16 {
		var c uint16
		if int(ci) < len(r.colors.revTable) {
			c = uint16(r.colors.revTable[ci])
		}
		return mul.MultI(c).Make16()
	})
}

func (r *NoxRender) pixBlendPremult(dst []uint16, src []byte, sz int) (_ []uint16, _ []byte) {
	return r.drawOpU16(dst, src, sz, func(c1 uint16, c2 uint16) uint16 {
		cc := SplitColor16(c1)
		cc.R += (c2>>8)&0xF8 | 7
		cc.G += (c2>>3)&0xFC | 3
		cc.B += (c2<<3)&0xF8 | 7
		return cc.Saturate().Make16()
	})
}

func (r *NoxRender) pixOpSrcMultiplyIndexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(_ uint16, c byte) uint16 {
		return mul.Mult(pmul.MultI(uint16(c))).Make16()
	})
}

func (r *NoxRender) pixOpSrcIndexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(_ uint16, c byte) uint16 {
		return pmul.MultI(uint16(c)).Make16()
	})
}

func (r *NoxRender) pixOpOverAlpha50Indexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(c1 uint16, c2 byte) uint16 {
		cc1 := SplitColor16(c1)
		return cc1.Over(pmul.MultI(uint16(c2))).Make16()
	})
}

func (r *NoxRender) pixOpOverAlphaIndexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	a := uint16(r.p.Alpha())
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(c1 uint16, c2 byte) uint16 {
		cc1 := SplitColor16(c1)
		return cc1.OverAlpha(a, pmul.MultI(uint16(c2))).Make16()
	})
}

func (r *NoxRender) pixOpOverMultiplyAlpha50Indexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	mul := r.p.ColorMultA()
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(c1 uint16, c2 byte) uint16 {
		cc1 := SplitColor16(c1)
		return cc1.Over(pmul.MultI(uint16(c2)).Mult(mul)).Make16()
	})
}

func (r *NoxRender) pixOpOverMultiplyAlphaIndexed(dst []uint16, src []byte, op byte, sz int) (_ []uint16, _ []byte) {
	a := uint16(r.p.Alpha())
	mul := r.p.ColorMultA()
	pmul := r.p.ColorMultOp(int(op))

	return r.drawOpU8(dst, src, sz, func(c1 uint16, c2 byte) uint16 {
		cc1 := SplitColor16(c1)
		return cc1.OverAlpha(a, pmul.MultI(uint16(c2)).Mult(mul)).Make16()
	})
}
