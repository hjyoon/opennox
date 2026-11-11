package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestChakramRetargetNativeLayout4EB250(t *testing.T) {
	u, d := server.Object{}, server.ChakramUpdateData{}
	prefix := unsafe.Offsetof(u.ScriptPickup) + unsafe.Sizeof(u.ScriptPickup)
	want := [11]uintptr{unsafe.Offsetof(u.ObjClass), unsafe.Offsetof(u.ObjFlags), unsafe.Offsetof(u.PosVec), unsafe.Offsetof(u.VelVec), unsafe.Offsetof(u.ObjOwner), unsafe.Offsetof(u.SpeedCur), unsafe.Offsetof(u.UpdateData), prefix, unsafe.Offsetof(d.LastHit), unsafe.Offsetof(d.ReturnState), unsafe.Sizeof(d)}
	if got := chakramRetargetOriginalLayout4EB250(); got != want {
		t.Fatal("independent complete C/native Object and ChakramUpdateData layouts", got, want)
	}
}

func TestChakramRetargetActualCBefore4EB250(t *testing.T) {
	if os.Getenv("OPENNOX_CHAKRAM_RETARGET_C_BEFORE") != "1" {
		return
	}
	if mode := os.Getenv("OPENNOX_CHAKRAM_RETARGET_C_CHILD"); mode != "" {
		if mode != "driver" && mode != "callback" {
			t.Fatal("unknown before mode", mode)
		}
		TestChakramRetargetNativeLayout4EB250(t)
		s, source, _, owner := monsterMainBlockLegacyFixture547210(t)
		update, freeUpdate := alloc.NewHigh(server.ChakramUpdateData{})
		last, freeLast := alloc.NewHigh(server.Object{})
		target, freeTarget := alloc.NewHigh(server.Object{})
		t.Cleanup(freeUpdate)
		t.Cleanup(freeLast)
		t.Cleanup(freeTarget)
		source.ObjClass, source.ObjOwner, source.UpdateData = object.ClassMissile, owner, unsafe.Pointer(update)
		source.SpeedCur = 11
		update.LastHit = last
		*target = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagActive | object.FlagEnabled, PosVec: types.Ptf(320, 330), NewPos: types.Ptf(320, 330)}
		target.Shape.Kind = server.ShapeKindCenter
		s.Map.AddObjectToIndex(target)
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
		found := false
		s.Map.EachObjInRect(types.Rectf{Min: types.Ptf(0, 0), Max: types.Ptf(700, 700)}, func(it *server.Object) bool {
			if it == target {
				found = true
			}
			return true
		})
		if !found || target.HasEnchant(0) || !s.MapTraceVision(target, source) || !chakramRetargetOriginalMap4EB250(target, source) {
			t.Fatal("real native rectangle/buff/vision and typed C vision controls")
		}
		if got := chakramRetargetOriginalControls4EB250(source); got != [3]uintptr{uintptr(unsafe.Pointer(owner)), uintptr(unsafe.Pointer(update)), uintptr(unsafe.Pointer(last))} {
			t.Fatal("typed C source/owner/update/LastHit chain loses native identity", got)
		}
		records := []struct {
			ptr  unsafe.Pointer
			size uintptr
		}{{unsafe.Pointer(source), unsafe.Sizeof(*source)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(last), unsafe.Sizeof(*last)}, {unsafe.Pointer(owner), unsafe.Sizeof(*owner)}, {unsafe.Pointer(target), unsafe.Sizeof(*target)}}
		before := make([][]byte, len(records))
		for i, r := range records {
			if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("all actual before records must be above4GiB")
			}
			before[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		entry := 0
		if mode == "callback" {
			entry = 1
		}
		if chakramRetargetOriginalABI4EB250(entry) != 32 {
			t.Fatal("unchanged old public C is required")
		}
		t.Logf("CHAKRAM_RETARGET_REAL_NATIVE_CONTROLS mode=%s source=%p update=%p target=%p rectangle-buff-Go-and-C-vision=true full-records=true", mode, source, update, target)
		if mode == "driver" {
			chakramRetargetOriginalC4EB250(0, source, nil)
			t.Fatal("unported high-pointer C driver unexpectedly returned")
		}
		// The old C callback reads class bits as a float and converts their
		// numeric value to uint32. ClassMonster becomes zero and prevents it
		// from reaching the narrowed point. Record that semantic failure;
		// it would be false to claim this valid callback input SIGSEGVs.
		chakramRetargetOriginalSeed4EB250(source)
		chakramRetargetOriginalC4EB250(1, target, &source.PosVec)
		if chakramRetargetOriginalSelected4EB250() != 0 {
			t.Fatal("unchanged old callback must reproduce lost valid class bits")
		}
		for i, r := range records {
			if !bytes.Equal(before[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				t.Fatal("whole valid native record modified by old rejected callback", i)
			}
		}
		t.Log("CHAKRAM_RETARGET_OLD_CALLBACK_CLASS_BITS_LOST valid-target-not-selected=true original-integer-bit-test=true whole-records=true")
		return
	}
	for _, mode := range []string{"driver", "callback"} {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestChakramRetargetActualCBefore4EB250$")
			cmd.Env = append(os.Environ(), "OPENNOX_CHAKRAM_RETARGET_C_CHILD="+mode)
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if !bytes.Contains(out, []byte("CHAKRAM_RETARGET_REAL_NATIVE_CONTROLS")) {
				t.Fatal("independent controls did not complete", err)
			}
			if mode == "driver" {
				if err == nil || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_chakram_retarget_original_public_4eb250")) {
					t.Fatal("actual old driver must fault only after valid controls", err)
				}
			} else if err != nil || !bytes.Contains(out, []byte("CHAKRAM_RETARGET_OLD_CALLBACK_CLASS_BITS_LOST")) {
				t.Fatal("actual old callback semantic mismatch must be observed", err)
			}
		})
	}
}
