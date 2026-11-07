//go:build darwin && arm64

package legacy

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

// Expected values come exclusively from the immutable original call frames
// and completed original stores. The helper under test never produces an
// expected value, never executes a parent algorithm, and receives no row.
func TestRawDefaultDamageNativeFloatOriginalFrames4E0B30(t *testing.T) {
	bundle, err := rddOpenBundle4E0B30()
	if err != nil {
		t.Fatal(err)
	}
	defer bundle.close()
	type observedCounts struct{ Rows, Converter, DirectSpill, ProtectionProduct, VampProduct int }
	counts := map[uint16]*observedCounts{0x0e3f: {}, 0x023f: {}}
	for _, runName := range []string{"baseline1034", "additive4"} {
		run, ok := bundle.Index.Runs[runName]
		if !ok {
			t.Fatal("immutable original run missing")
		}
		err = bundle.rows(run, func(raw json.RawMessage) error {
			var row rddRow4E0B30
			if err := json.Unmarshal(raw, &row); err != nil {
				return err
			}
			cw := uint16(row.Control)
			count := counts[cw]
			if count == nil {
				return fmt.Errorf("unsealed original control word %#x", row.Control)
			}
			count.Rows++
			var input rddInput4E0B30
			if err := json.Unmarshal(row.Input, &input); err != nil {
				return err
			}
			damage := input.Args[3]
			var loaded rawDefaultDamageFloat4E0B30
			loadedValid := false
			var storedProtection, product uint32
			protectionValid, productValid := false, false
			fail := func(format string, args ...any) error {
				return fmt.Errorf("%s/%s: %s", row.Mode, row.Name, fmt.Sprintf(format, args...))
			}
			for _, rawEvent := range row.Events {
				var event rddEvent4E0B30
				if err := json.Unmarshal(rawEvent, &event); err != nil {
					return err
				}
				if event.Kind == "return-completed" {
					var frame rddFrame4E0B30
					if err := json.Unmarshal(event.Frame, &frame); err != nil {
						return err
					}
					if mustRDDNumber4E0B30(frame.Entry) == 0x419a70 {
						if len(frame.Args) != 1 || !frame.Completed {
							return fail("converter exact completed one-word frame")
						}
						want, exists := frame.After["eax"]
						if !exists {
							return fail("converter actual original EAX missing")
						}
						if got := uint32(rawDefaultDamageFloatConvert32Bits4E0B30(frame.Args[0], cw)); got != want {
							return fail("converter call=%s input=%08x got=%08x original=%08x", frame.CallSite, frame.Args[0], got, want)
						}
						count.Converter++
					}
					if len(frame.X87) != 0 && string(frame.X87) != "null" {
						var packet struct{ Format, Hex string }
						if err := json.Unmarshal(frame.X87, &packet); err != nil {
							return err
						}
						payload, err := hex.DecodeString(packet.Hex)
						width, known := map[string]int{"f32": 4, "f64": 8, "ext80": 10}[packet.Format]
						if err != nil || !known || len(payload) != width {
							return fail("exact raw original x87 reply format/width")
						}
						reply := rawDefaultDamageNativeReply4E0B30{FloatFormat: packet.Format}
						copy(reply.FloatPayload[:], payload)
						loaded = rawDefaultDamageFloatFromReply4E0B30(reply)
						loadedValid = true
					}
				}
				if event.Kind == "original-instruction" {
					switch mustRDDNumber4E0B30(event.PC) {
					case 0x4e0dd7, 0x4e0e40:
						if !protectionValid {
							return fail("product without original stored protection")
						}
						product = rawDefaultDamageFloatProtectionProduct4E0B30(storedProtection, damage, cw)
						productValid = true
					case 0x4e10b5:
						if !loadedValid {
							return fail("Vamp multiply without original unspilled reply")
						}
						product = rawDefaultDamageFloatStore32Bits4E0B30(rawDefaultDamageFloatMultiplyDamage4E0B30(loaded, int32(damage), cw), cw)
						productValid = true
					}
				}
				if event.Kind != "write" && event.Kind != "host-write" {
					continue
				}
				after, err := hex.DecodeString(event.After)
				if err != nil {
					return err
				}
				if event.Width == 4 && len(after) == 4 && mustRDDNumber4E0B30(event.Address) == rddCaller4E0B30+16 {
					// Use the observed current original cell, including declared
					// in/out service changes; never simulate damage branches.
					damage = binary.LittleEndian.Uint32(after)
				}
				if event.Kind != "write" || !event.Completed {
					continue
				}
				pc := mustRDDNumber4E0B30(event.PC)
				if pc != 0x4e0cf1 && pc != 0x4e0da2 && pc != 0x4e0e0b && pc != 0x4e0ddc && pc != 0x4e0e45 && pc != 0x4e10bc {
					continue
				}
				if event.Width != 4 || len(after) != 4 {
					return fail("original FST32 width")
				}
				want := binary.LittleEndian.Uint32(after)
				got := product
				switch pc {
				case 0x4e0cf1, 0x4e0da2, 0x4e0e0b:
					if !loadedValid {
						return fail("direct spill without original reply")
					}
					got = rawDefaultDamageFloatStore32Bits4E0B30(loaded, cw)
					count.DirectSpill++
					if pc != 0x4e0cf1 {
						storedProtection, protectionValid = want, true
					}
				case 0x4e0ddc, 0x4e0e45, 0x4e10bc:
					if !productValid {
						return fail("FST32 without observed original multiply boundary")
					}
					productValid = false
					if pc == 0x4e10bc {
						count.VampProduct++
					} else {
						count.ProtectionProduct++
					}
				}
				if got != want {
					return fail("FST32 PC=%#x got=%08x original=%08x", pc, got, want)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for cw, count := range counts {
		if count.Rows != 1038 || count.Converter == 0 || count.DirectSpill == 0 || count.ProtectionProduct == 0 || count.VampProduct == 0 {
			t.Fatalf("original numerical boundary coverage CW=%#x counts=%+v", cw, count)
		}
		t.Logf("RAW_DEFAULT_DAMAGE_FLOAT_ORIGINAL_FRAMES CW=%04x rows=%d converters=%d direct_spills=%d protection_products=%d vamp_products=%d parent_engine_leaves=0", cw, count.Rows, count.Converter, count.DirectSpill, count.ProtectionProduct, count.VampProduct)
	}
}
