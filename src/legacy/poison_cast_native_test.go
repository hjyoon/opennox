package legacy

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/poison_cast_original_52c720_v1.json.gz
var poisonCastNativeOriginal []byte

type poisonCastNativeMutation struct {
	Kind, Action string
	Slot, Field  int
	Value        int64
}
type poisonCastNativeInput struct {
	Name                       string
	ID                         int32
	Second, Owner, Caster, Arg int
	Level                      int32
	Objects                    [6][2]uint32
	Arguments                  [2][3]int64
	ActivateResult             uint32 `json:"activate_result"`
	AttributionResult          uint32 `json:"attribution_result"`
	Mutations                  []poisonCastNativeMutation
}
type poisonCastNativeEvent struct {
	Kind          string
	At            int
	Args          []uint32
	Result        uint32
	Before, After string
}
type poisonCastNativeCase struct {
	Input         poisonCastNativeInput
	Before, After [18]uint32
	Events        []poisonCastNativeEvent
	Return        *uint32 `json:"return_u32"`
	Fault         *struct{ Kind string }
}
type poisonCastNativeArena struct {
	raw, expected []byte
	objects       [6]*server.Object
	arguments     [2]*server.SpellAcceptArg
	placement     int
}

var poisonCastNativeClassOffset = unsafe.Offsetof(server.Object{}.ObjClass)
var poisonCastNativeFlagsOffset = unsafe.Offsetof(server.Object{}.ObjFlags)
var poisonCastNativePositionOffset = unsafe.Offsetof(server.SpellAcceptArg{}.Pos)

func poisonCastNativeArenaNew(t *testing.T, placement int) *poisonCastNativeArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 65536)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("native arena must exceed4GiB")
	}
	a := &poisonCastNativeArena{raw: raw, placement: placement}
	for i := range a.objects {
		a.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x800+placement]))
	}
	for i := range a.arguments {
		a.arguments[i] = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x5000+i*0x40+placement]))
	}
	return a
}
func (a *poisonCastNativeArena) object(n int) *server.Object {
	if n == -1 {
		return nil
	}
	return a.objects[n]
}
func (a *poisonCastNativeArena) role(p *server.Object) uint32 {
	if p == nil {
		return math.MaxUint32
	}
	for i, o := range a.objects {
		if p == o {
			return uint32(i)
		}
	}
	panic("unowned native object")
}
func (a *poisonCastNativeArena) offset(p unsafe.Pointer) int {
	n := int(uintptr(p) - uintptr(unsafe.Pointer(&a.raw[0])))
	if n < 0 || n >= len(a.raw) {
		panic("pointer outside native owner")
	}
	return n
}
func poisonCastNativePut(b []byte, off, size int, value uint64) {
	for i := 0; i < size; i++ {
		b[off+i] = byte(value >> uint(8*i))
	}
}
func (a *poisonCastNativeArena) patch(b []byte, m poisonCastNativeMutation) {
	if m.Action == "object" {
		off := poisonCastNativeClassOffset
		if m.Field == 1 {
			off = poisonCastNativeFlagsOffset
		}
		poisonCastNativePut(b, a.offset(unsafe.Pointer(a.objects[m.Slot]))+int(off), 4, uint64(m.Value))
		return
	}
	off := a.offset(unsafe.Pointer(a.arguments[m.Slot]))
	if m.Field == 0 {
		poisonCastNativePut(b, off, int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(a.object(int(m.Value))))))
	} else {
		poisonCastNativePut(b, off+int(poisonCastNativePositionOffset)+4*(m.Field-1), 4, uint64(m.Value))
	}
}
func (a *poisonCastNativeArena) initialize(in poisonCastNativeInput) {
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	for i, row := range in.Objects {
		for j, v := range row {
			a.patch(a.raw, poisonCastNativeMutation{Action: "object", Slot: i, Field: j, Value: int64(v)})
		}
	}
	for i, row := range in.Arguments {
		for j, v := range row {
			a.patch(a.raw, poisonCastNativeMutation{Action: "argument", Slot: i, Field: j, Value: v})
		}
	}
	a.expected = bytes.Clone(a.raw)
}
func (a *poisonCastNativeArena) words() (out [18]uint32) {
	at := 0
	for _, o := range a.objects {
		out[at], out[at+1] = uint32(o.ObjClass), uint32(o.ObjFlags)
		at += 2
	}
	for _, arg := range a.arguments {
		off := a.offset(unsafe.Pointer(arg)) + int(poisonCastNativePositionOffset)
		out[at], out[at+1], out[at+2] = a.role(arg.Obj), binary.LittleEndian.Uint32(a.raw[off:]), binary.LittleEndian.Uint32(a.raw[off+4:])
		at += 3
	}
	return
}
func (a *poisonCastNativeArena) state() string {
	words := a.words()
	var b [72]byte
	for i, v := range words {
		binary.LittleEndian.PutUint32(b[4*i:], v)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b[:]))
}
func (a *poisonCastNativeArena) check(t *testing.T) {
	t.Helper()
	if !bytes.Equal(a.raw, a.expected) {
		t.Fatal("unexpected write in64KiB native owner")
	}
}
func (a *poisonCastNativeArena) mutate(in poisonCastNativeInput, kind string) {
	for _, m := range in.Mutations {
		if m.Kind == kind {
			a.patch(a.raw, m)
			a.patch(a.expected, m)
		}
	}
}
func (a *poisonCastNativeArena) zero(p unsafe.Pointer, n uintptr) {
	off := a.offset(p)
	clear(a.raw[off : off+int(n)])
}
func (a *poisonCastNativeArena) expect(p unsafe.Pointer, n uintptr, v uint64) {
	poisonCastNativePut(a.expected, a.offset(p), int(n), v)
}

func TestPoisonNativeLayout(t *testing.T) {
	var arg server.SpellAcceptArg
	want := [7]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(arg), unsafe.Offsetof(arg.Obj),
		unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.X), unsafe.Offsetof(arg.Pos) + unsafe.Offsetof(arg.Pos.Y),
		poisonCastNativeClassOffset, poisonCastNativeFlagsOffset}
	if got := poisonCastNativeLayoutC(); got != want {
		t.Fatal("independent native C object/acceptance layout", got, want)
	}
}

func TestPoisonNativeOriginal(t *testing.T) {
	var rows []poisonCastNativeCase
	npcVoiceOriginalDecode(t, poisonCastNativeOriginal, "68f94cbd02b6ba3e111d89d653f2e33907513f1c3b0330c7fcefd894a4e86764", "642e42f0173a1ee6f2a4ce3cf5aaa7c3cb635546dce9743204f3306f7faac960", 53, &rows)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			a := poisonCastNativeArenaNew(t, placement)
			for _, c := range rows {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					a.initialize(c.Input)
					if a.words() != c.Before {
						t.Fatal("canonical native input differs")
					}
					cursor := 0
					observe := func(kind string, args []uint32, result uint32) {
						a.check(t)
						before := a.state()
						a.mutate(c.Input, kind)
						event := poisonCastNativeEvent{Kind: kind, Args: args, Result: result, Before: before, After: a.state()}
						if cursor >= len(c.Events) || !reflect.DeepEqual(event, c.Events[cursor]) {
							t.Fatalf("service%d %s differs: %+v", cursor, kind, event)
						}
						cursor++
					}
					previous := poisonCastCall52C720
					defer func() { poisonCastCall52C720 = previous }()
					poisonCastCall52C720 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
						if int32(id) != c.Input.ID || second != a.object(c.Input.Second) || owner != a.object(c.Input.Owner) || caster != a.object(c.Input.Caster) || int32(level) != c.Input.Level {
							t.Fatal("native public argument identity changed")
						}
						return int(new(server.Server).CastPoison52C720(int32(id), second, owner, caster, arg, int32(level), server.PoisonCastRuntime52C720{
							ActivatePoison: func(target *server.Object, increment, maximum int32) int32 {
								observe("activate", []uint32{a.role(target), uint32(increment), uint32(maximum)}, c.Input.ActivateResult)
								return int32(c.Input.ActivateResult)
							},
							RecordPlayerAttribution: func(owner, target *server.Object) {
								// The original void service's declared EAX control is ignored by
								// the parent; it is reference metadata rather than a Go return.
								observe("attribution", []uint32{a.role(owner), a.role(target)}, c.Input.AttributionResult)
							},
						}))
					}
					var arg *server.SpellAcceptArg
					if c.Input.Arg != -1 {
						arg = a.arguments[c.Input.Arg]
					}
					var caught any
					var got int32
					func() {
						defer func() { caught = recover() }()
						got = poisonCastNativeCall(route, c.Input.ID, a.object(c.Input.Second), a.object(c.Input.Owner), a.object(c.Input.Caster), arg, c.Input.Level)
					}()
					if c.Fault != nil {
						err, ok := caught.(runtime.Error)
						if !ok || !strings.Contains(err.Error(), "invalid memory address") || cursor != 0 || c.Return != nil {
							t.Fatalf("original nil-argument first-read fault changed: %v", caught)
						}
					} else if caught != nil || c.Return == nil || uint32(got) != *c.Return {
						t.Fatalf("original return/fault changed: %d/%v", got, caught)
					}
					if cursor != len(c.Events) || a.words() != c.After {
						t.Fatal("complete original observations/final canonical state differ")
					}
					a.check(t)
				})
			}
		}
	}
}

func TestPoisonNativePublicPointers(t *testing.T) {
	a := poisonCastNativeArenaNew(t, 64)
	for i := range a.raw {
		a.raw[i] = byte(i*31 + 19)
	}
	previous := poisonCastCall52C720
	t.Cleanup(func() { poisonCastCall52C720 = previous })
	for route := 0; route < 3; route++ {
		for _, v := range []int32{math.MinInt32, -1, 0, 7, math.MaxInt32} {
			t.Run(fmt.Sprintf("route%d/id%d", route, v), func(t *testing.T) {
				count := 0
				poisonCastCall52C720 = func(id spell.ID, second, owner, caster *server.Object, arg *server.SpellAcceptArg, level int) int {
					count++
					if int32(id) != v || second != a.objects[0] || owner != a.objects[1] || caster != a.objects[2] || arg != a.arguments[1] || level != int(v) {
						t.Fatal("native pointer/signed scalar forwarding changed")
					}
					return -12345
				}
				before := bytes.Clone(a.raw)
				if got := poisonCastNativeCall(route, v, a.objects[0], a.objects[1], a.objects[2], a.arguments[1], v); got != -12345 || count != 1 || !bytes.Equal(before, a.raw) {
					t.Fatalf("public return/calls/owner changed: %d/%d", got, count)
				}
			})
		}
	}
}

func TestPoisonNativeDefaultBindings(t *testing.T) {
	previousServer, previousFrame, previousCast := GetServer, gameFrameHook, poisonCastCall52C720
	t.Cleanup(func() { GetServer, gameFrameHook, poisonCastCall52C720 = previousServer, previousFrame, previousCast })
	poisonCastCall52C720 = castPoisonNative52C720
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			for _, kind := range []string{"nil-target", "immune-monster", "state-increase", "rejected-player-attribution"} {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, kind), func(t *testing.T) {
					a := poisonCastNativeArenaNew(t, placement)
					for i := range a.raw {
						a.raw[i] = byte(i*31 + 19)
					}
					for _, o := range a.objects {
						a.zero(unsafe.Pointer(o), unsafe.Sizeof(*o))
					}
					for _, arg := range a.arguments {
						a.zero(unsafe.Pointer(arg), unsafe.Sizeof(*arg))
					}
					target, arg := a.objects[0], a.arguments[0]
					arg.Obj = target
					s := new(server.Server)
					GetServer = func() Server { return &poisonCastLegacyServer52C720{srv: s} }
					var owner *server.Object
					var health *server.HealthData
					var targetPlayer *server.Player
					level, want := int32(3), int32(1)
					switch kind {
					case "nil-target":
						arg.Obj = nil
						want = 0
					case "immune-monster":
						target.ObjClass = object.ClassMonster
						target.ObjSubClass = object.SubClass(0x200)
						target.Poison540 = 7
					case "state-increase":
						health = (*server.HealthData)(unsafe.Pointer(&a.raw[0x9000+placement]))
						a.zero(unsafe.Pointer(health), unsafe.Sizeof(*health))
						target.HealthData = health
						s.Rand.Logic = prand.New(0)
						s.SetFrame(91)
					case "rejected-player-attribution":
						owner = a.objects[1]
						sourceUpdate := (*server.PlayerUpdateData)(unsafe.Pointer(&a.raw[0xa000+placement]))
						targetUpdate := (*server.PlayerUpdateData)(unsafe.Pointer(&a.raw[0xa400+placement]))
						sourcePlayer := (*server.Player)(unsafe.Pointer(&a.raw[0xb000+placement]))
						targetPlayer = (*server.Player)(unsafe.Pointer(&a.raw[0xd000+placement]))
						a.zero(unsafe.Pointer(sourceUpdate), unsafe.Sizeof(*sourceUpdate))
						a.zero(unsafe.Pointer(targetUpdate), unsafe.Sizeof(*targetUpdate))
						a.zero(unsafe.Pointer(sourcePlayer), unsafe.Sizeof(*sourcePlayer))
						a.zero(unsafe.Pointer(targetPlayer), unsafe.Sizeof(*targetPlayer))
						sourcePlayer.PlayerInd = 0xfe
						targetPlayer.Field3680 = 1
						sourceUpdate.Player = sourcePlayer
						targetUpdate.Player = targetPlayer
						owner.ObjClass, owner.UpdateData = object.ClassPlayer, unsafe.Pointer(sourceUpdate)
						target.ObjClass, target.UpdateData, target.Poison540 = object.ClassPlayer, unsafe.Pointer(targetUpdate), 7
						gameFrameHook = func() uint32 { return 0xfedcba98 }
						level = -3
					}
					a.expected = bytes.Clone(a.raw)
					if kind == "state-increase" {
						a.expect(unsafe.Pointer(&target.Poison540), unsafe.Sizeof(target.Poison540), 3)
						a.expect(unsafe.Pointer(&target.Field542), unsafe.Sizeof(target.Field542), 1000)
						a.expect(unsafe.Pointer(&health.Field16), unsafe.Sizeof(health.Field16), 91)
					} else if targetPlayer != nil {
						expected := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(targetPlayer)), int(unsafe.Sizeof(*targetPlayer))))
						player := (*server.Player)(unsafe.Pointer(&expected[0]))
						player.SetLastAggressorPending(1)
						player.SetLastAggressorPlayerIndex(0xfe)
						player.SetLastAggressorFrame(0xfedcba98)
						copy(a.expected[a.offset(unsafe.Pointer(targetPlayer)):], expected)
					}
					got := poisonCastNativeCall(route, int32(spell.SPELL_POISON), nil, owner, nil, arg, level)
					if got != want {
						t.Fatalf("actual default result%d want%d", got, want)
					}
					if kind == "state-increase" && s.Rand.Logic.Index() != 1 {
						t.Fatal("actual poison RNG order changed")
					}
					if targetPlayer != nil {
						pending, index, frame := targetPlayer.LastAggressorState()
						if pending != 1 || index != 0xfe || frame != 0xfedcba98 {
							t.Fatal("actual native player attribution changed")
						}
					}
					a.check(t)
				})
			}
		}
	}
}
