package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestContinualReplenishment4E02C0NativeChargePacketAndHolderPointersV3(t *testing.T) {
	// The actual C report queues an important packet, rather than invoking the
	// Go Server.NetSendPacketXxx test hook. Own and restore that native queue.
	handles.Init()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	pool := alloc.NewClass("continual-replenishment-native-report", importantPacketSizeC(), 4)
	setImportantAllocClassC(pool.UPtr())
	setImportantRecipientMaskC(0xffffffff)
	gameFrameHook = func() uint32 { return 240 }
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeUpdate()
	player, freePlayer := alloc.New(server.Player{})
	defer freePlayer()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	use, freeUse := alloc.New(server.WandUseData{})
	defer freeUse()
	effect, freeEffect := alloc.New(server.ModifierEff{})
	defer freeEffect()
	owner.ObjClass = object.ClassPlayer
	owner.UpdateData = unsafe.Pointer(update)
	update.Player = player
	item.ObjClass = object.ClassWand
	item.InvHolder = owner
	item.UseData.Ptr = unsafe.Pointer(use)
	item.NetCode = 0x1234
	use.Charge = 254
	use.MaxCharge = 255
	use.Progress = 0xa5a5a5a5
	effect.Update100 = server.ModifierEffFnc{Fnc: continualReplenishmentEffectPointer4E02C0(), Val: 240}
	srv := new(server.Server)
	srv.SetFrame(240)
	old := GetServer
	t.Cleanup(func() { GetServer = old })
	GetServer = func() Server { return &continualReplenishmentServer218{srv: srv} }
	var previous unsafe.Pointer
	for _, ind := range []uint8{0, 31, 128, 255} {
		player.PlayerInd = ind
		use.Charge = 254
		effect.CallUpdateNil(item)
		first, last := importantNativeListC()
		if first == nil || last == nil || first == previous || importantPacketRelatedObjectC(first) != nil {
			t.Fatal("missing native charge report")
		}
		packet := (*importantPacketLegacy)(first)
		if packet.CreatedFrame != 240 || uint8(packet.Recipient) != ind || packet.PayloadSize != 5 || !bytes.Equal(packet.Payload[:5], []byte{100, 0x34, 0x12, 255, 255}) || packet.RemoveIfDisconnected != 0 || packet.SequenceEnabled != 1 || use.Charge != 255 || use.Progress != 0xa5a5a5a5 {
			t.Fatalf("native queue frame=%d recipient=%d payload=%x remove=%d sequence=%d charge=%d progress=%x", packet.CreatedFrame, uint8(packet.Recipient), packet.Payload[:5], packet.RemoveIfDisconnected, packet.SequenceEnabled, use.Charge, use.Progress)
		}
		previous = first
	}
	t.Log("actual callback -> native owner/player -> C charge packet -> important queue: indices0/31/128/255")
}
