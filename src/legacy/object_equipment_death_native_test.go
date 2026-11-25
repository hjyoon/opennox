package legacy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestEquipmentDeathSendsLocalizedMessageAndDeletes(t *testing.T) {
	strings := strman.New()
	path := filepath.Join(t.TempDir(), "strings.json")
	const localized = `{"entries":[
		{"id":"die.c:WeaponDieMetal","vals":[{"str":"Metal %s broke"}]},
		{"id":"die.c:ArmorDieWood","vals":[{"str":"Wood %s broke"}]}
	]}`
	if err := os.WriteFile(path, []byte(localized), 0600); err != nil {
		t.Fatal(err)
	}
	if err := strings.ReadJSON(path); err != nil {
		t.Fatal(err)
	}
	world := server.NewHigh(nil, nil, strings)
	t.Cleanup(world.Close)
	world.NetList.Init()
	t.Cleanup(world.NetList.Free)
	outer := &questLoseArmorLegacyServer54CD30{native: world}
	oldServer, oldNames := GetServer, objectItemNameRuntime
	t.Cleanup(func() {
		GetServer, objectItemNameRuntime = oldServer, oldNames
	})
	GetServer = func() Server { return outer }
	name := itemNameWidePointer4E77E0("item")
	objectItemNameRuntime = objectItemNameRuntime4E77E0{
		buffer:        func() *uint16 { return name },
		noDescription: func() *uint16 { return name },
		copy:          func(*uint16) {},
	}

	const index = 3
	player := world.Players.ByIndRaw(index)
	player.PlayerInd = index
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	update.Player = player
	holder, freeHolder := alloc.NewHigh(server.Object{})
	t.Cleanup(freeHolder)
	*holder = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	item, freeItem := alloc.NewHigh(server.Object{})
	t.Cleanup(freeItem)
	*item = server.Object{InvHolder: holder, TypeInd: 0x8001}

	for _, tc := range []struct {
		handler  string
		material object.Material
		message  string
	}{
		{"WeaponDie", object.MaterialMetal | object.MaterialWood, "Metal item broke"},
		{"ArmorDie", object.MaterialWood | object.MaterialCloth, "Wood item broke"},
	} {
		t.Run(tc.handler, func(t *testing.T) {
			item.Material = uint16(tc.material)
			callback, size, ok := server.ObjectDeathHandler(tc.handler)
			if !ok || callback == nil || size != 0 {
				t.Fatalf("registered callback = %p/%d/%t", callback, size, ok)
			}
			outer.deletes, outer.deleted = 0, nil
			server.CallObjectDeath(callback, item)
			if outer.deletes != 1 || outer.deleted != item {
				t.Fatalf("delayed deletion = %d/%p, want 1/%p", outer.deletes, outer.deleted, item)
			}
			want := append([]byte{0xa8, 0, 0, 2, 0, 0, 0, 0, byte(len(tc.message) + 1), 0, 0}, tc.message...)
			want = append(want, 0)
			if got := world.NetList.CopyPacketsA(index, netlist.Kind1); !bytes.Equal(got, want) {
				t.Fatalf("localized packet = %x, want %x", got, want)
			}
		})
	}
}
