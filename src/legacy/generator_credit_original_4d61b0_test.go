package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type generatorCreditFault4D61B0 struct{ Kind, Site string }
type generatorCreditWrite4D61B0 struct {
	PC, Address string
	Value       uint32
}
type generatorCreditOutput4D61B0 struct {
	Name           string
	Fault          *generatorCreditFault4D61B0
	OriginalWrites []generatorCreditWrite4D61B0 `json:"original_writes"`
	HeapSHA256     string                       `json:"heap_sha256"`
}

type generatorCreditRunner4D61B0 interface {
	RecordMonsterGeneratorCreditRaw4D61B0()
}

// Keep the absent-engine branch dynamic so the entire consumer is compiled
// before the native method exists.
//
//go:noinline
func generatorCreditRunnerOf4D61B0(value any) (generatorCreditRunner4D61B0, bool) {
	runner, ok := value.(generatorCreditRunner4D61B0)
	return runner, ok
}

func (m *generatorCreditMemory4D61B0) finalHeap(writes []generatorCreditWrite4D61B0) string {
	pe, native := bytes.Clone(m.pe), bytes.Clone(m.before)
	for _, slot := range m.slots {
		binary.LittleEndian.PutUint32(pe[slot.pe:], m.toPE(m.pointerWord(slot.native)))
	}
	for _, write := range writes {
		var address uint32
		if _, err := fmt.Sscanf(write.Address, "0x%x", &address); err != nil {
			m.t.Fatal("original write address", err)
		}
		var at int
		switch write.PC {
		case "0x4d61d6":
			if address != m.chosenPE+4668 {
				m.t.Fatal("original counter destination")
			}
			at = m.chosenNative + int(m.off[17])
		case "0x4d61e2":
			delta := int64(address) - int64(m.chosenPE) - 4692
			if delta < 0 || delta > 32 {
				m.t.Fatal("undeclared original player-reload displacement", delta)
			}
			at = m.chosenNative + int(m.off[18]) + int(delta)
		default:
			m.t.Fatal("undeclared original write PC", write.PC)
		}
		copy(native[at:at+4], m.buf[at:at+4])
		value := binary.LittleEndian.Uint32(m.buf[at:])
		for _, slot := range m.slots {
			if at == slot.native {
				value = m.toPE(m.pointerWord(at))
			}
		}
		if value != write.Value {
			m.t.Fatalf("original write destination %s value=%#x want=%#x", write.Address, value, write.Value)
		}
		binary.LittleEndian.PutUint32(pe[int(address-0x21000000):], value)
	}
	// Every native byte outside the original observable write destinations
	// must remain exact. Normalized pointer slots are not permissions to write.
	if !bytes.Equal(native, m.buf) {
		m.t.Fatal("native memory changed outside the exact original write destinations")
	}
	return fmt.Sprintf("%x", sha256.Sum256(pe))
}

func generatorCreditInvoke4D61B0(run func()) (fault *generatorCreditFault4D61B0) {
	old := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(old)
	defer func() {
		if value := recover(); value != nil {
			if _, ok := value.(interface{ RuntimeError() }); !ok {
				panic(value)
			}
			stack := string(debug.Stack())
			for _, site := range []struct{ helper, name string }{
				{"generatorCreditFlags4D61B0", "unit-flags"},
				{"generatorCreditUpdate4D61B0", "unit-update"},
				{"generatorCreditPlayer4D61B0", "update-player"},
				{"generatorCreditIncrement4D61B0", "player-counter"},
				{"generatorCreditSetFlags4D61B0", "player-flags"},
			} {
				if strings.Contains(stack, "."+site.helper+"(") {
					fault = &generatorCreditFault4D61B0{Kind: "memory", Site: site.name}
					return
				}
			}
			panic(fmt.Sprintf("unclassified native memory failure: %v\n%s", value, stack))
		}
	}()
	run()
	return
}

func TestGeneratorCreditOriginal4D61B0(t *testing.T) {
	inputs, initial := generatorCreditInputs4D61B0(t)
	for _, profile := range []string{"tozero", "nearest"} {
		t.Run(profile, func(t *testing.T) {
			var expected []generatorCreditOutput4D61B0
			trollDeathFixture54A270(t, "generator_credit_original_4d61b0_pc53_"+profile+".json.gz", "0eb974de529c5318fb01df4a4319a88893d55b3dfc379e3d1ba10c7f91bca2b5", "f0c52f84e982bcf12e461aa1f86fa0712631434a17baceae2c96251ff5064759", &expected)
			if len(expected) != len(inputs) {
				t.Fatal("complete immutable original outputs")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					_, goPresent := generatorCreditRunnerOf4D61B0((*server.Object)(nil))
					if (mode == "Go" && !goPresent) || (mode == "C" && specialDeathOriginalABI54E460(3) != 64) {
						t.Log("complete consumer compiled before native implementation; no native comparison credit")
						return
					}
					for i, input := range inputs {
						t.Run(input.Name, func(t *testing.T) {
							if expected[i].Name != input.Name || initial[i].Name != input.Name {
								t.Fatal("immutable independent input/output identity")
							}
							m := newGeneratorCreditMemory4D61B0(t, input, initial[i])
							fault := generatorCreditInvoke4D61B0(func() {
								if mode == "Go" {
									method, _ := generatorCreditRunnerOf4D61B0(m.argument)
									method.RecordMonsterGeneratorCreditRaw4D61B0()
								} else {
									specialDeathOriginalC54E460(3, m.argument)
								}
							})
							want := expected[i]
							if (fault == nil) != (want.Fault == nil) || (fault != nil && *fault != *want.Fault) {
								t.Fatal("actual native fault and immutable original site", fault, want.Fault)
							}
							if got := m.finalHeap(want.OriginalWrites); got != want.HeapSHA256 {
								t.Fatal("whole original heap output and complete native memory guards", got, want.HeapSHA256)
							}
						})
					}
				})
			}
		})
	}
}
