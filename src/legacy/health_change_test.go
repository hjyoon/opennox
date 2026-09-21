package legacy

import (
	"encoding/binary"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"golang.org/x/image/font"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

type healthChangeRender struct {
	Render2
	data      *noxrender.RenderData
	fonts     noxrender.RenderFonts
	strings   []string
	positions []image.Point
}

func (r *healthChangeRender) Data() *noxrender.RenderData {
	return r.data
}

func (r *healthChangeRender) GetFonts() *noxrender.RenderFonts {
	return &r.fonts
}

func (r *healthChangeRender) GetStringSizeWrapped(font.Face, string, int) image.Point {
	return image.Pt(10, 7)
}

func (r *healthChangeRender) DrawString(_ font.Face, str string, pos image.Point) int {
	r.strings = append(r.strings, str)
	r.positions = append(r.positions, pos)
	return 1
}

type healthChangeClient struct {
	Client
	seq    uint
	render *healthChangeRender
}

func (c *healthChangeClient) GetInputSeq() uint {
	return c.seq
}

func (c *healthChangeClient) R2() Render2 {
	return c.render
}

func newHealthChangeTestClient(t *testing.T) *healthChangeClient {
	t.Helper()
	handles.Init()
	data, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	render := &healthChangeRender{data: data}
	client := &healthChangeClient{render: render}
	oldGetClient := GetClient
	GetClient = func() Client { return client }
	t.Cleanup(func() { GetClient = oldGetClient })
	oldIsConnected := Nox_client_isConnected
	Nox_client_isConnected = func() bool { return false }
	t.Cleanup(func() { Nox_client_isConnected = oldIsConnected })
	ok := healthChangeTestBegin()
	t.Cleanup(healthChangeTestEnd)
	if !ok {
		t.Fatal("cannot allocate native HealthChange test pool")
	}
	return client
}

func TestHealthChangeNativeLayout(t *testing.T) {
	size, nextOffset, prevOffset := healthChangeNativeLayout()
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if size != 20 || nextOffset != 12 || prevOffset != 16 {
			t.Fatalf("PE32 layout = size %d, links %d/%d; want 20, 12/16", size, nextOffset, prevOffset)
		}
		return
	}
	if size != 32 || nextOffset != 16 || prevOffset != 24 {
		t.Fatalf("native 64-bit layout = size %d, links %d/%d; want 32, 16/24", size, nextOffset, prevOffset)
	}
}

func TestHealthChangePointerSlotsPreserveNativeWidth(t *testing.T) {
	want := uintptr(0x12345)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		highBit := uint64(1) << 40
		want += uintptr(highBit)
	}
	for slot, name := range []string{"list head", "numbers font", "next link", "previous link"} {
		if got := healthChangePointerRoundTrip(slot, want); got != want {
			t.Errorf("%s pointer round trip = %#x, want %#x", name, got, want)
		}
	}
}

func TestHealthChangePacketRendersAndExpiresDamageNumber(t *testing.T) {
	client := newHealthChangeTestClient(t)
	const drawableID = uint16(0x1234)
	damage := int16(-17)

	client.seq = 100
	packet := []byte{66, 0, 0, 0, 0}
	binary.LittleEndian.PutUint16(packet[1:3], drawableID)
	binary.LittleEndian.PutUint16(packet[3:5], uint16(damage))
	if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, netmsg.Op(66), packet); got != len(packet) {
		t.Fatalf("MSG_REPORT_HEALTH_DELTA consumed %d bytes, want %d", got, len(packet))
	}

	head := healthChangeHeadSnapshot()
	if head.address == 0 {
		t.Fatal("MSG_REPORT_HEALTH_DELTA did not create a damage-number event")
	}
	if head.drawableID != uint32(drawableID) || head.delta != damage || head.frame != 100 || head.next != 0 || head.prev != 0 {
		t.Fatalf("damage event = %+v, want id %#x, delta %d, frame 100, no links", head, drawableID, damage)
	}
	if got, ok := HealthChangeForDrawable(uint32(drawableID)); !ok || got != damage {
		t.Fatalf("HealthChangeForDrawable = %d, %t; want %d, true", got, ok, damage)
	}
	if got, ok := HealthChangeForDrawable(0xffff); ok || got != 0 {
		t.Fatalf("missing HealthChangeForDrawable = %d, %t; want 0, false", got, ok)
	}

	client.seq = 105
	healthChangeDraw(uint32(drawableID), 100, 200, 20, 30, 50, 80, 3, 4.75)
	if want := []string{"17", "17", "17", "17", "17"}; !reflect.DeepEqual(client.render.strings, want) {
		t.Fatalf("drawn strings = %#v, want %#v", client.render.strings, want)
	}
	if want := []image.Point{
		image.Pt(124, 232),
		image.Pt(124, 234),
		image.Pt(126, 232),
		image.Pt(126, 234),
		image.Pt(125, 233),
	}; !reflect.DeepEqual(client.render.positions, want) {
		t.Fatalf("draw positions = %#v, want %#v", client.render.positions, want)
	}

	client.seq = 131
	healthChangeDraw(uint32(drawableID), 100, 200, 20, 30, 50, 80, 3, 4.75)
	if len(client.render.strings) != 5 {
		t.Fatalf("expired damage number drew again: got %d total draws, want 5", len(client.render.strings))
	}
	if !healthChangeHeadEmpty() {
		t.Fatal("damage-number event was not removed after its 30-frame lifetime")
	}
}

func TestHealthChangeRendersFromNativeGoDrawable(t *testing.T) {
	cli := newHealthChangeTestClient(t)
	const drawableID = uint16(0x1234)
	damage := int16(-17)

	cli.seq = 100
	packet := []byte{66, 0, 0, 0, 0}
	binary.LittleEndian.PutUint16(packet[1:3], drawableID)
	binary.LittleEndian.PutUint16(packet[3:5], uint16(damage))
	if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, netmsg.Op(66), packet); got != len(packet) {
		t.Fatalf("MSG_REPORT_HEALTH_DELTA consumed %d bytes, want %d", got, len(packet))
	}

	cli.seq = 105
	vp := &noxrender.Viewport{
		Screen: image.Rect(100, 200, 740, 680),
		World:  image.Rect(20, 30, 660, 510),
	}
	dr := &client.Drawable{
		PosVec:    image.Pt(50, 80),
		ZVal:      3,
		ZSizeMax:  4.75,
		NetCode32: uint32(drawableID),
	}
	Sub_49A6A0(vp, dr)

	if want := []string{"17", "17", "17", "17", "17"}; !reflect.DeepEqual(cli.render.strings, want) {
		t.Fatalf("drawn strings = %#v, want %#v", cli.render.strings, want)
	}
	if want := []image.Point{
		image.Pt(124, 232),
		image.Pt(124, 234),
		image.Pt(126, 232),
		image.Pt(126, 234),
		image.Pt(125, 233),
	}; !reflect.DeepEqual(cli.render.positions, want) {
		t.Fatalf("draw positions = %#v, want %#v", cli.render.positions, want)
	}
}

func TestHealthChangeListLinksMultipleEventsAtNativeWidth(t *testing.T) {
	client := newHealthChangeTestClient(t)
	client.seq = 77

	for id, damage := range []struct {
		id     uint16
		damage int16
	}{{0x1111, -3}, {0x2222, -9}} {
		packet := []byte{66, 0, 0, 0, 0}
		binary.LittleEndian.PutUint16(packet[1:3], damage.id)
		binary.LittleEndian.PutUint16(packet[3:5], uint16(damage.damage))
		if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(ntype.PlayerInd(id), netmsg.Op(66), packet); got != len(packet) {
			t.Fatalf("event %d consumed %d bytes, want %d", id, got, len(packet))
		}
	}

	head := healthChangeHeadSnapshot()
	if head.drawableID != 0x2222 || head.delta != -9 || head.next == 0 || head.prev != 0 {
		t.Fatalf("newest event = %+v, want linked id %#x delta -9", head, 0x2222)
	}
	older := healthChangeSnapshotAt(head.next)
	if older.drawableID != 0x1111 || older.delta != -3 || older.next != 0 || older.prev != head.address {
		t.Fatalf("older event = %+v, want id %#x delta -3, prev %#x", older, 0x1111, head.address)
	}
	client.seq = 108
	healthChangeDraw(0x2222, 0, 0, 0, 0, 0, 0, 0, 0)
	if !healthChangeHeadEmpty() {
		t.Fatal("expired linked damage-number events were not both removed")
	}
}
