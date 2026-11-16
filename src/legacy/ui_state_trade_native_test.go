package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/ui_state_trade_original_v1.json.gz
var uiStateTradeOriginal []byte

type uiStateTradeOriginalCase struct {
	Entry, Name string
	Input       struct {
		Word       uint32
		OwnerIndex int `json:"owner"`
		FirstIndex int `json:"first_owner"`
		Gold       [2]uint32
		Costs      [8]uint32
		Chains     [2][]int
		Fault      string
	}
	Return *uint32 `json:"return_u32"`
	Fault  []struct {
		Address, Size, PC uint32
	}
}

func uiStateTradeCases(t *testing.T) []uiStateTradeOriginalCase {
	t.Helper()
	var cases []uiStateTradeOriginalCase
	npcVoiceOriginalDecode(t, uiStateTradeOriginal, "a0aacd9a9784b31760d8d8e50f20a3229167688ba1caecac54267c58b2068454", "96ba57fbfa0842f8cce4b36ed82b21649a231b2bda459cbef0b2f29ea22aa0e5", 1630, &cases)
	return cases
}

func TestUIStateTradeNativeLayout(t *testing.T) {
	var session server.TradeSession
	var item server.TradeItem
	var obj server.Object
	want := [14]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(session), unsafe.Offsetof(session.Field8), unsafe.Offsetof(session.Field32), unsafe.Offsetof(session.Field36), unsafe.Offsetof(session.Field48), unsafe.Offsetof(session.Field52), unsafe.Sizeof(item), unsafe.Offsetof(item.Cost4), unsafe.Offsetof(item.Field8), unsafe.Sizeof(item.Cost4), unsafe.Sizeof(item.Field8), unsafe.Offsetof(obj.InitData), unsafe.Sizeof(obj.InitData)}
	if got := uiStateTradeLayoutC(); got != want {
		t.Fatal("native session/list/object field layouts", got, want)
	}
}

func TestUIStateTradeNativeWords(t *testing.T) {
	names := []string{"sub_41E2F0", "nox_client_mousePriKey_430AF0", "nox_xxx_cursor_430B00", "sub_43AF30", "sub_43AF40", "sub_43AF80", "sub_4611A0", "sub_4674A0", "sub_4675B0", "sub_46A4A0", "nox_xxx_playerGet_470A90", "sub_478030", "sub_479590", "sub_47A260"}
	var saved [14]uint32
	for i := range saved {
		saved[i] = uiStateWordGet(i)
	}
	t.Cleanup(func() {
		for i, value := range saved {
			uiStateWordSet(i, value)
		}
	})
	cases := uiStateTradeCases(t)
	for getter, name := range names {
		for route := 0; route < 4; route++ {
			for _, c := range cases {
				if c.Entry != name {
					continue
				}
				t.Run(fmt.Sprintf("%s/route%d/%s", name, route, c.Name), func(t *testing.T) {
					if c.Return == nil || len(c.Fault) != 0 || *c.Return != c.Input.Word {
						t.Fatal("invalid original scalar observation")
					}
					var before [14]uint32
					for i := range before {
						before[i] = 0xcccccccc + uint32(i)
						uiStateWordSet(i, before[i])
					}
					before[getter] = c.Input.Word
					uiStateWordSet(getter, c.Input.Word)
					if route < 3 {
						if got := uiStateWordCall(route, getter); uint32(got) != *c.Return {
							t.Fatal("original C return word", got, *c.Return)
						}
					} else {
						want := int64(int32(*c.Return))
						if getter == 3 || getter == 7 {
							want = int64(*c.Return)
						} else if getter == 0 || getter == 10 {
							// These pre-existing wrappers read uint32 directly into host int.
							want = int64(int(*c.Return))
						}
						if got := uiStateWordPublic(getter); got != want {
							t.Fatal("existing public Go conversion", got, want)
						}
					}
					for i, want := range before {
						if got := uiStateWordGet(i); got != want {
							t.Fatal("shared scalar changed", i, got, want)
						}
					}
				})
			}
		}
	}
}

func TestUIStateTradeNativeValue(t *testing.T) {
	cases := uiStateTradeCases(t)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign session/list/gold owner must exceed4GiB")
			}
			session := (*server.TradeSession)(unsafe.Pointer(&raw[0x1000+placement]))
			var owners [3]*server.Object
			var golds [2]*server.Object
			var data [2]*uint32
			var nodes [8]*server.TradeItem
			for i := range owners {
				owners[i] = (*server.Object)(unsafe.Pointer(&raw[0x2000+i*0x1000+placement]))
			}
			for i := range golds {
				golds[i] = (*server.Object)(unsafe.Pointer(&raw[0x5000+i*0x1000+placement]))
				data[i] = (*uint32)(unsafe.Pointer(&raw[0x7000+i*0x100+placement]))
			}
			for i := range nodes {
				nodes[i] = (*server.TradeItem)(unsafe.Pointer(&raw[0x8000+i*0x100+placement]))
			}
			expected := make([]byte, len(raw))
			for _, c := range cases {
				if c.Entry != "sub_50FD20" || c.Input.Fault != "" {
					// Eight original invalid-read prefixes are CPU/model evidence only.
					continue
				}
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Name), func(t *testing.T) {
					if c.Return == nil || len(c.Fault) != 0 {
						t.Fatal("invalid original trade observation")
					}
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					owner := func(i int) *server.Object {
						if i == -1 {
							return nil
						}
						if i < 0 || i >= len(owners) {
							t.Fatal("unowned canonical identity", i)
						}
						return owners[i]
					}
					session.Field8 = owner(c.Input.FirstIndex)
					session.Field48, session.Field52 = golds[0], golds[1]
					session.Field32, session.Field36 = nil, nil
					for i, node := range nodes {
						node.Cost4, node.Field8 = c.Input.Costs[i], nil
					}
					for side, chain := range c.Input.Chains {
						golds[side].InitData = unsafe.Pointer(data[side])
						*data[side] = c.Input.Gold[side]
						for at, i := range chain {
							if i < 0 || i >= len(nodes) {
								t.Fatal("unowned canonical node", i)
							}
							if at+1 < len(chain) {
								nodes[i].Field8 = nodes[chain[at+1]]
							}
						}
						if len(chain) > 0 {
							if side == 0 {
								session.Field32 = nodes[chain[0]]
							} else {
								session.Field36 = nodes[chain[0]]
							}
						}
					}
					copy(expected, raw)
					if got := uiStateTradeCall(route, session, owner(c.Input.OwnerIndex)); uint32(got) != *c.Return {
						t.Fatal("original wrapping32 trade value", got, *c.Return)
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("unexpected write to whole64KiB foreign session/list/gold owner")
					}
				})
			}
		}
	}
}
