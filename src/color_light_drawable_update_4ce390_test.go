package opennox

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestUpdateColorLightDrawable4CE390NativeWidth(t *testing.T) {
	vp := &noxrender.Viewport{World: image.Rect(0, 0, 100, 100), Size: image.Pt(200, 200)}
	dr := &client.Drawable{PosVec: image.Pt(50, 50)}
	target := &client.Drawable{PosVec: image.Pt(50, 150)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(target)) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}
	dr.UnionEffect().Field_108 = 2 | 2<<8 | 2<<16
	data := dr.ColorLightData()
	data[0] = 1 | 4 | 0x10 | 0x40
	copy(data[2:8], []byte{10, 20, 30, 110, 120, 130})
	data[50], data[51] = 2, 6
	data[66], data[67] = 90, 180
	binary.LittleEndian.PutUint16(data[82:84], 10)
	binary.LittleEndian.PutUint16(data[84:86], 10)
	binary.LittleEndian.PutUint16(data[86:88], 10)
	binary.LittleEndian.PutUint32(data[88:92], 0x12345678)

	frameCalls := 0
	lookupCalls := 0
	hooks := colorLightDrawableHooks4CE390{
		gameFlag22: func() bool { return false },
		frame: func() uint32 {
			frameCalls++
			return 5
		},
		fps: func() uint32 { return 30 },
		delete: func(*client.Drawable) {
			t.Fatal("active color light was deleted")
		},
		byStaticCode: func(code int) *client.Drawable {
			lookupCalls++
			if code != 0x12345678 {
				t.Fatalf("target code = %#x", code)
			}
			return target
		},
	}
	if got := updateColorLightDrawable4CE390(vp, dr, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if frameCalls != 3 || lookupCalls != 1 {
		t.Fatalf("frame calls=%d lookup calls=%d", frameCalls, lookupCalls)
	}
	if dr.LightFlags != 2 || dr.LightColor.R != 60 || dr.LightColor.G != 70 || dr.LightColor.B != 80 {
		t.Fatalf("light color = flags %d, rgb %+v", dr.LightFlags, dr.LightColor)
	}
	if dr.LightIntensity != 4 {
		t.Fatalf("light intensity = %v, want 4", dr.LightIntensity)
	}
	if got, want := dr.LightPenumbra, uint16(setColorLightSize4CE760(&client.Drawable{}, 135)); got != want {
		t.Fatalf("light penumbra = %d, want %d", got, want)
	}
	// The legacy path rounds acos to float32, then truncates the resulting
	// 89.999... degree value before converting it to the fixed angle.
	if got, want := dr.LightDir, uint16(setColorLightDirection4CE8C0(&client.Drawable{}, 89)); got != want {
		t.Fatalf("light direction = %d, want %d", got, want)
	}
	if dr.Field_42 != 0 {
		t.Fatalf("field 42 = %d, want 0", dr.Field_42)
	}
}

func TestUpdateColorLightDrawable4CE390Guards(t *testing.T) {
	if got := updateColorLightDrawable4CE390(nil, nil, colorLightDrawableHooks4CE390{}); got != 1 {
		t.Fatalf("nil update = %d", got)
	}

	t.Run("empty animation deletes before culling", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(-1000, -1000)}
		deleted := false
		flagCalls := 0
		got := updateColorLightDrawable4CE390(&noxrender.Viewport{}, dr, colorLightDrawableHooks4CE390{
			delete: func(got *client.Drawable) {
				deleted = got == dr
			},
			gameFlag22: func() bool {
				flagCalls++
				return false
			},
		})
		if got != 1 || !deleted || flagCalls != 0 {
			t.Fatalf("update=%d deleted=%t flag calls=%d", got, deleted, flagCalls)
		}
	})

	t.Run("offscreen short circuits game flag", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(201, 0)}
		dr.UnionEffect().Field_108 = 2
		flagCalls := 0
		got := updateColorLightDrawable4CE390(&noxrender.Viewport{World: image.Rect(0, 0, 50, 50), Size: image.Pt(100, 100)}, dr, colorLightDrawableHooks4CE390{
			gameFlag22: func() bool {
				flagCalls++
				return false
			},
		})
		if got != 1 || flagCalls != 0 {
			t.Fatalf("update=%d flag calls=%d", got, flagCalls)
		}
	})

	t.Run("editor flag suppresses animation", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(10, 10)}
		dr.UnionEffect().Field_108 = 2
		frameCalls := 0
		got := updateColorLightDrawable4CE390(&noxrender.Viewport{World: image.Rect(0, 0, 50, 50), Size: image.Pt(100, 100)}, dr, colorLightDrawableHooks4CE390{
			gameFlag22: func() bool { return true },
			frame: func() uint32 {
				frameCalls++
				return 0
			},
		})
		if got != 1 || frameCalls != 0 {
			t.Fatalf("update=%d frame calls=%d", got, frameCalls)
		}
	})

	t.Run("single key does not read frame", func(t *testing.T) {
		dr := &client.Drawable{PosVec: image.Pt(10, 10)}
		dr.UnionEffect().Field_108 = 1 | 1<<8 | 1<<16
		got := updateColorLightDrawable4CE390(&noxrender.Viewport{Size: image.Pt(100, 100)}, dr, colorLightDrawableHooks4CE390{
			gameFlag22: func() bool { return false },
		})
		if got != 1 {
			t.Fatalf("update = %d, want 1", got)
		}
	})
}

func TestRotateColorLight4CE960(t *testing.T) {
	dr := &client.Drawable{Field_42: 1}
	data := dr.ColorLightData()
	binary.LittleEndian.PutUint16(data[0:2], 0x180)
	binary.LittleEndian.PutUint16(data[92:94], 30)
	binary.LittleEndian.PutUint16(data[94:96], 10)
	binary.LittleEndian.PutUint16(data[96:98], 90)
	frameCalls := 0
	rotateColorLight4CE960(dr, data, colorLightDrawableHooks4CE390{
		fps: func() uint32 { return 30 },
		frame: func() uint32 {
			frameCalls++
			return 45
		},
	})
	if frameCalls != 2 {
		t.Fatalf("frame calls = %d, want 2", frameCalls)
	}
	if got, want := dr.LightDir, uint16(setColorLightDirection4CE8C0(&client.Drawable{}, 45)); got != want {
		t.Fatalf("light direction = %d, want %d", got, want)
	}
	if dr.Field_42 != 0 {
		t.Fatalf("field 42 = %d, want 0", dr.Field_42)
	}
}

func TestColorLightKeyFrames4CE440PingPong(t *testing.T) {
	from, to, fraction := colorLightKeyFrames4CE440(1.25, 2, 10, false)
	if from != 1 || to != 0 || fraction != 0.5 {
		t.Fatalf("key frames = %d -> %d at %v", from, to, fraction)
	}
}
