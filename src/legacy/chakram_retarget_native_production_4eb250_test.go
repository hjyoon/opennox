package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestChakramRetargetActualCProduction4EB250(t *testing.T) {
	modes := []string{"driver", "empty", "excluded-owner", "excluded-last", "invisible", "rejected-flags", "invalid-class", "far", "tie", "boundary400", "boundary400-circle", "callback", "callback-point-alias", "callback-excluded"}
	if mode := os.Getenv("OPENNOX_CHAKRAM_RETARGET_PRODUCTION_CHILD"); mode != "" {
		if !slices.Contains(modes, mode) || chakramRetargetNativeRuntimeOverride4EB250 != nil || chakramRetargetOriginalABI4EB250(0) != 64 || chakramRetargetOriginalABI4EB250(1) != 64 {
			t.Fatal("actual native C entries and default runtime required", mode)
		}
		TestChakramRetargetNativeLayout4EB250(t)
		s, source, oldUpdate, owner := monsterMainBlockLegacyFixture547210(t)
		update, freeUpdate := alloc.New(server.ChakramUpdateData{})
		t.Cleanup(freeUpdate)
		last, freeLast := alloc.New(server.Object{})
		t.Cleanup(freeLast)
		target, freeTarget := alloc.New(server.Object{})
		t.Cleanup(freeTarget)
		source.ObjClass, source.ObjOwner, source.UpdateData = object.ClassMissile, owner, unsafe.Pointer(update)
		source.PosVec, source.NewPos, source.SpeedCur = types.Ptf(10, 20), types.Ptf(10, 20), 11
		source.VelVec = types.Ptf(6, 5)
		update.LastHit, update.ReturnState = last, 0xa5
		*target = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagActive | object.FlagEnabled, PosVec: types.Ptf(19, 32), NewPos: types.Ptf(19, 32)}
		target.Shape.Kind = server.ShapeKindCenter
		accepted := true
		switch mode {
		case "empty", "excluded-owner", "excluded-last", "invisible", "rejected-flags", "invalid-class", "far", "boundary400", "callback-excluded":
			accepted = false
		}
		switch mode {
		case "excluded-owner", "callback-excluded":
			source.ObjOwner = target
		case "excluded-last":
			update.LastHit = target
		case "invisible":
			target.Buffs = 1
		case "rejected-flags":
			target.ObjFlags |= object.Flags(0x20)
		case "invalid-class":
			target.ObjClass = object.ClassMissile
		case "far":
			target.PosVec, target.NewPos = types.Ptf(900, 900), types.Ptf(900, 900)
		case "boundary400", "boundary400-circle":
			target.PosVec, target.NewPos = types.Ptf(410, 20), types.Ptf(410, 20)
			if mode == "boundary400-circle" {
				target.Shape.Kind = server.ShapeKindCircle
				target.Shape.Circle.R, target.Shape.Circle.R2 = 1, 1
			}
		case "callback-point-alias":
			source.VelVec = source.PosVec
		}
		var tied *server.Object
		if mode == "tie" {
			value, free := alloc.New(server.Object{})
			t.Cleanup(free)
			tied = value
			*tied = *target
			s.Map.AddObjectToIndex(tied)
			t.Cleanup(func() { s.Map.RemoveObjectFromIndex(tied) })
		}
		if mode != "empty" {
			s.Map.AddObjectToIndex(target)
			t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
		}
		// Independently traverse the real index to establish the first equal
		// target without injecting a selection into either production entry.
		var enumerated []*server.Object
		s.Map.EachObjInRect(types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(410, 420)}, func(it *server.Object) bool {
			if it == target || it == tied {
				enumerated = append(enumerated, it)
			}
			return true
		})
		actualRect := types.Rectf{Min: types.Ptf(-390, -380), Max: types.Ptf(410, 420)}
		actualCEnumerated := chakramRetargetProductionRectControl4EB250(actualRect)
		if accepted && !slices.Contains(actualCEnumerated, target) {
			t.Fatal("actual C rectangle did not admit the valid target", mode)
		}
		if mode == "boundary400" && slices.Contains(actualCEnumerated, target) {
			t.Fatal("center exactly on the rectangle edge must retain actual spatial admission")
		}
		wantTarget := target
		if mode == "tie" {
			if len(enumerated) != 2 {
				t.Fatal("real tie controls", enumerated)
			}
			wantTarget = enumerated[0]
		}
		if accepted && (!chakramRetargetOriginalMap4EB250(wantTarget, source) || wantTarget.HasEnchant(0)) {
			t.Fatal("real typed C visibility/buff positive controls")
		}
		controls := chakramRetargetOriginalControls4EB250(source)
		if controls != [3]uintptr{uintptr(unsafe.Pointer(source.ObjOwner)), uintptr(unsafe.Pointer(update)), uintptr(unsafe.Pointer(update.LastHit))} {
			t.Fatal("real typed C source chain controls", controls)
		}
		oldScratch := chakramRetargetNativeScratch4EB250
		t.Cleanup(func() { chakramRetargetNativeScratch4EB250 = oldScratch })
		chakramRetargetNativeScratch4EB250 = server.ChakramRetargetState4EB250{Source: source, Owner: source.ObjOwner, LastHit: update.LastHit, BestDistance: math.Float32frombits(0x4b189680)}
		type record struct {
			ptr  unsafe.Pointer
			size uintptr
		}
		records := []record{{unsafe.Pointer(source), unsafe.Sizeof(*source)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(owner), unsafe.Sizeof(*owner)}, {unsafe.Pointer(last), unsafe.Sizeof(*last)}, {unsafe.Pointer(target), unsafe.Sizeof(*target)}, {unsafe.Pointer(oldUpdate), unsafe.Sizeof(*oldUpdate)}, {unsafe.Pointer(source.HealthData), unsafe.Sizeof(*source.HealthData)}}
		if tied != nil {
			records = append(records, record{unsafe.Pointer(tied), unsafe.Sizeof(*tied)})
		}
		want := make([][]byte, len(records))
		for i, r := range records {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("actual production records must remain above4GiB")
			}
			want[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		callback := strings.HasPrefix(mode, "callback")
		var result *server.Object
		if callback {
			point := &source.PosVec
			if mode == "callback-point-alias" {
				point = &source.VelVec
			}
			chakramRetargetOriginalC4EB250(1, target, point)
		} else {
			result = chakramRetargetOriginalC4EB250(0, source, nil)
		}
		if !accepted {
			wantTarget = nil
		}
		if !callback && result != wantTarget {
			t.Fatalf("actual C target=%p want%p mode=%s", result, wantTarget, mode)
		}
		state := chakramRetargetNativeScratch4EB250
		if state.Source != source || state.Owner != source.ObjOwner || state.LastHit != update.LastHit || state.Selected != wantTarget {
			t.Fatal("actual C shared native pointer identities", state, wantTarget)
		}
		wantBest := uint32(0x4b189680)
		if accepted {
			wantBest = 1130430464
			if mode == "boundary400-circle" {
				wantBest = 0x481c4000
			}
		}
		if math.Float32bits(state.BestDistance) != wantBest {
			t.Fatal("actual C best distance", math.Float32bits(state.BestDistance), wantBest)
		}
		if accepted && !callback {
			// Exact original PC53/toward-zero result for driver/route/3.
			velocity := [2]uint32{1087491364, 1091296792}
			if mode == "boundary400-circle" {
				velocity = [2]uint32{1093661885, 0}
			}
			binary.LittleEndian.PutUint32(want[0][unsafe.Offsetof(source.VelVec):], velocity[0])
			binary.LittleEndian.PutUint32(want[0][unsafe.Offsetof(source.VelVec)+4:], velocity[1])
			want[1][unsafe.Offsetof(update.ReturnState)] = 2
		}
		// The unchanged real rectangle/vision services update only their
		// Field62 visitation tokens on indexed records. Keep exact expected
		// tokens as declared world effects; every other byte remains guarded.
		visited := slices.Contains([]string{"driver", "excluded-owner", "excluded-last", "invisible", "tie", "boundary400", "boundary400-circle", "callback", "callback-point-alias"}, mode)
		if visited {
			for _, index := range []int{4, 7} {
				if index >= len(records) {
					continue
				}
				offset := unsafe.Offsetof(target.Field62)
				previous := binary.LittleEndian.Uint32(want[index][offset:])
				binary.LittleEndian.PutUint32(want[index][offset:], previous+1)
				if !callback && accepted {
					inner := uint32(1)
					if mode == "tie" {
						inner = 2
					}
					binary.LittleEndian.PutUint32(want[index][offset+4:], inner)
				}
			}
		}
		for i, r := range records {
			if !bytes.Equal(want[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				for at, value := range want[i] {
					if actual := unsafe.Slice((*byte)(r.ptr), int(r.size))[at]; value != actual {
						t.Fatalf("complete actual native record differs mode=%s record=%d offset=%x got=%02x want=%02x", mode, i, at, actual, value)
					}
				}
			}
		}
		t.Logf("CHAKRAM_RETARGET_ACTUAL_C_PRODUCTION mode=%s source=%p selected=%p default-C-rectangle-callback-buff-vision=true full-records=%d", mode, source, wantTarget, len(records))
		return
	}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestChakramRetargetActualCProduction4EB250$")
			cmd.Env = append(os.Environ(), "OPENNOX_CHAKRAM_RETARGET_PRODUCTION_CHILD="+mode)
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err != nil || !bytes.Contains(out, []byte("CHAKRAM_RETARGET_ACTUAL_C_PRODUCTION mode="+mode)) {
				t.Fatal("actual default C production child", mode, err)
			}
		})
	}
}
