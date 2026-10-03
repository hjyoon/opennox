package opennox

import (
	"fmt"
	"image"
	"time"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2ePlayerStatusAnimation struct {
	kind     string
	buff     server.EnchantID
	unit     *server.Object
	ref      *legacy.ImageRef
	frame    uint32
	duration uint32
	baseline *noximage.Image16
	first    int
}

func e2ePlayerStatusBuff(kind string) (server.EnchantID, bool) {
	switch kind {
	case "stun":
		return server.ENCHANT_HELD, true
	case "confused":
		return server.ENCHANT_CONFUSED, true
	case "nullify":
		return server.ENCHANT_ANTI_MAGIC, true
	case "charming":
		return server.ENCHANT_CHARMING, true
	case "shield":
		return server.ENCHANT_SHIELD, true
	case "slow":
		return server.ENCHANT_SLOWED, true
	default:
		return 0, false
	}
}

func (sc *e2eScenario) CheckPlayerStatusAnimation(kind, name string) {
	sc.checkPlayerStatusAnimation(kind, name, "player status "+kind, 180, func(unit *server.Object, buff server.EnchantID) uint32 {
		// Status preparation uses the normal enchant API; client buffs,
		// rendering, and expiry remain the actual game pipeline.
		asObjectS(unit).ApplyEnchant(buff, 90, 1)
		return 90
	})
}

func (sc *e2eScenario) checkPlayerStatusAnimation(kind, name, screenPrefix string, expiryTimeout time.Duration, apply func(*server.Object, server.EnchantID) uint32) {
	buff, ok := e2ePlayerStatusBuff(kind)
	if !ok {
		e2eError(fmt.Errorf("unknown player status animation %q", kind))
		return
	}
	f := &e2ePlayerStatusAnimation{kind: kind, buff: buff, first: -1}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		dr := noxClient.ClientPlayerUnit()
		return unit != nil && unit.HealthData != nil && unit.HealthData.Cur > 0 && unit.Buffs == 0 &&
			dr != nil && dr.Buffs == 0 && nox_client_isConnected()
	}, func() {
		f.unit = noxServer.Players.HostUnit()
		if f.unit.Buffs != 0 || noxClient.ClientPlayerUnit().Buffs != 0 {
			e2eError(fmt.Errorf("status animation needs an unenchanted baseline: server=%#x client=%#x", f.unit.Buffs, noxClient.ClientPlayerUnit().Buffs))
			return
		}
		// Check both real startup caches before dereferencing them. On LP64 a
		// legacy U32 write cannot initialize memmap's native pointer side slot.
		for _, asset := range []struct {
			name string
			off  uintptr
		}{{"ConfusedBirdies", 1096456}, {"SphericalShieldAnim", 1096460}} {
			want := nox_xxx_gLoadAnim(asset.name)
			got := legacy.AsImageRefP(*memmap.PtrPtr(0x5D4594, asset.off))
			if want == nil || got != want {
				e2eError(fmt.Errorf("status animation startup cache missing/truncated: asset=%s cache=%p loaded=%p packed=%#x", asset.name, got, want, memmap.Uint32(0x5D4594, asset.off)))
				return
			}
			if want.Kind() != 2 || len(want.Field24ptr().Images()) < 2 || want.Field24ptr().AnimType != 2 {
				e2eError(fmt.Errorf("status animation asset is not a multi-frame loop: %s", asset.name))
				return
			}
			if kind != "slow" && (kind == "shield") == (asset.name == "SphericalShieldAnim") {
				f.ref = got
			}
			e2eLog.Printf("STATUS ANIMATION CACHE: asset=%s ref=%p native-high=%t frames=%d delay=%d", asset.name, got, uintptr(unsafe.Pointer(got)) > 0xffffffff, len(got.Field24ptr().Images()), got.Field24ptr().Field_2_1)
		}
		pix := noxClient.r.PixBuffer()
		f.baseline = noximage.NewImage16(pix.Rect)
		copy(f.baseline.Pix, pix.Pix)
		f.frame = noxServer.Frame()
		f.duration = apply(f.unit, buff)
		e2eLog.Printf("STATUS ANIMATION APPLIED: kind=%s enchant=%d frame=%d timer=%d", kind, buff, f.frame, f.unit.EnchantDur(buff))
	})
	for _, sample := range []struct {
		label string
		dt    uint32
	}{{"first", 12}, {"advanced", 24}} {
		sc.addWhen(0, name+" "+sample.label+" visible", 120, func() bool {
			dr := noxClient.ClientPlayerUnit()
			return f.unit != nil && dr != nil && noxServer.Frame() >= f.frame+sample.dt &&
				f.unit.HasEnchant(buff) && dr.HasEnchant(buff)
		}, func() {
			dr := noxClient.ClientPlayerUnit()
			if dr.Buffs != f.unit.Buffs || memmap.Uint32(0x5D4594, 1062540) != f.unit.Buffs {
				e2eError(fmt.Errorf("status animation buff packet is not synchronized: server=%#x drawable=%#x local=%#x", f.unit.Buffs, dr.Buffs, memmap.Uint32(0x5D4594, 1062540)))
				return
			}
			if kind == "slow" {
				visible := e2eVisibleSlowParticles()
				if visible == 0 {
					e2eError(fmt.Errorf("slowed player has no visible YellowBubbleParticle drawables"))
					return
				}
				e2eLog.Printf("STATUS ANIMATION VISIBLE: kind=%s sample=%s frame=%d timer=%d particles=%d", kind, sample.label, noxServer.Frame(), f.unit.EnchantDur(buff), visible)
				return
			}
			index, matched, total := f.matchLivePixels(dr)
			if total < 10 || matched*100 < total*80 {
				e2eError(fmt.Errorf("status effect missing from actual frame: kind=%s sample=%s best-frame=%d matched=%d/%d", kind, sample.label, index, matched, total))
				return
			}
			if sample.label == "first" {
				f.first = index
			} else if index == f.first {
				e2eError(fmt.Errorf("status effect did not animate: kind=%s frame=%d", kind, index))
				return
			}
			e2eLog.Printf("STATUS ANIMATION VISIBLE: kind=%s sample=%s frame=%d timer=%d image-frame=%d matching-pixels=%d/%d", kind, sample.label, noxServer.Frame(), f.unit.EnchantDur(buff), index, matched, total)
		})
		// The live sprite/particle check above is the visual assertion. Keep
		// an actual frame for diagnosis without comparing unrelated animated
		// map objects against an incidental fixed frame or rewriting goldens.
		sc.CaptureMagicFrame(screenPrefix + " " + sample.label)
	}
	sc.addWhen(0, name+" natural expiry", expiryTimeout, func() bool {
		dr := noxClient.ClientPlayerUnit()
		return f.unit != nil && dr != nil && noxServer.Frame() >= f.frame+f.duration+5 &&
			!f.unit.HasEnchant(buff) && !dr.HasEnchant(buff)
	}, func() {
		if f.unit.Buffs != 0 || noxClient.ClientPlayerUnit().Buffs != 0 || memmap.Uint32(0x5D4594, 1062540) != 0 || f.unit.EnchantDur(buff) != 0 {
			e2eError(fmt.Errorf("expired player status remains on the server/client/HUD"))
			return
		}
		if kind == "slow" {
			if visible := e2eVisibleSlowParticles(); visible != 0 {
				e2eError(fmt.Errorf("slowed player particles remain after natural expiry: %d", visible))
				return
			}
		} else {
			_, matched, total := f.matchLivePixels(noxClient.ClientPlayerUnit())
			if total >= 10 && matched*100 >= total*80 {
				e2eError(fmt.Errorf("status effect remains visible after natural expiry: kind=%s pixels=%d/%d", kind, matched, total))
				return
			}
		}
		e2eLog.Printf("STATUS ANIMATION EXPIRED: kind=%s frame=%d elapsed=%d buffs=%#x", kind, noxServer.Frame(), noxServer.Frame()-f.frame, f.unit.Buffs)
	})
	sc.CaptureMagicFrame(screenPrefix + " expired")
}

func e2eVisibleSlowParticles() int {
	typeID := noxClient.Things.IndByID("YellowBubbleParticle")
	visible := 0
	for _, particle := range noxClient.Objs.AllList1() {
		if int(particle.TypeIDVal) == typeID && noxClient.Viewport().ToScreenPos(particle.Pos()).In(noxClient.Viewport().Screen) {
			visible++
		}
	}
	return visible
}

// Compare the already-rendered real game frame with every stock sprite frame.
// Reference sprites are rendered only into a separate scratch buffer, restored
// before returning. This cannot make a missing effect appear in the live image.
func (f *e2ePlayerStatusAnimation) matchLivePixels(dr *client.Drawable) (best, matched, total int) {
	r := noxClient.r
	live := r.PixBuffer()
	data := r.Data()
	before := *data
	defer func() { r.SetPixBuffer(live); *data = before }()
	pos := noxClient.Viewport().ToScreenPos(dr.Pos())
	if f.kind == "shield" {
		pos = pos.Add(image.Pt(-64, -90-dr.Z()))
		data.SetAlphaEnabled(true)
		data.SetAlpha(0x80)
	} else {
		pos = pos.Add(image.Pt(-64, 5-int(int16(dr.ZVal2))-dr.Z()-int(dr.GetZSizeMax())-64))
		if f.kind == "nullify" {
			data.SetColorize17(1)
			sub433E40(nox_color_blue_2650684)
		}
	}
	rect := image.Rect(pos.X, pos.Y, pos.X+128, pos.Y+128).Intersect(live.Rect)
	for i, handle := range f.ref.Field24ptr().Images() {
		pix := noximage.NewImage16(live.Rect)
		copy(pix.Pix, f.baseline.Pix)
		r.SetPixBuffer(pix)
		r.DrawImage16(r.Bag.AsImage(handle), pos)
		m, n := 0, 0
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				ind := pix.PixOffset(x, y)
				if pix.Pix[ind] != f.baseline.Pix[ind] {
					n++
					if live.Pix[ind] == pix.Pix[ind] {
						m++
					}
				}
			}
		}
		if i == 0 || n > 0 && m*total > matched*n {
			best, matched, total = i, m, n
		}
	}
	return best, matched, total
}
