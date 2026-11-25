package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestWOLMessageQueueKeepsNativePointerAndWordBounds(t *testing.T) {
	const index = 7
	const offset = 58132 + 16*index
	slot := memmap.PtrPtr(0x587000, offset)
	count := memmap.PtrUint32(0x587000, offset+4)
	previous, previousCount := *slot, *count
	previousChange := Sub_41E300
	t.Cleanup(func() {
		*slot, *count = previous, previousCount
		Sub_41E300 = previousChange
	})

	buffer, free := alloc.NewHigh([wolQueueSize + 1]uint16{})
	defer free()
	buffer[wolQueueSize] = 0xabcd
	*slot, *count = unsafe.Pointer(buffer), wolQueueSize-1
	fallbacks := 0
	Sub_41E300 = func(state int) int {
		if state != 11 {
			t.Fatalf("queue fallback state = %d", state)
		}
		fallbacks++
		return 0
	}

	if result := sub_41DA70(index, -2); result != 16*index || *count != wolQueueSize || buffer[wolQueueSize-1] != 0xfffe {
		t.Fatalf("last queue append = %d, count %d, value %#x", result, *count, buffer[wolQueueSize-1])
	}
	if result := sub_41DA70(index, 42); result != 0 || fallbacks != 1 || *count != wolQueueSize {
		t.Fatalf("full queue = %d, count %d, fallbacks %d", result, *count, fallbacks)
	}
	if result := sub_41DA10(index); result != 0 || *count != 0 || buffer[wolQueueSize] != 0xabcd || buffer[wolQueueSize-1] != 0 {
		t.Fatalf("queue reset = %d, count %d, queue %v", result, *count, buffer)
	}
	*slot, *count = nil, 12
	if reset, appendResult := sub_41DA10(index), sub_41DA70(index, 9); reset != 0 || appendResult != 0 || *count != 12 || fallbacks != 2 {
		t.Fatalf("missing queue = reset %d, append %d, count %d, fallbacks %d", reset, appendResult, *count, fallbacks)
	}
}

func TestWOLOfficialServerRecordNative(t *testing.T) {
	selected := cfixture.GlobalRecordNativeSelectedC(false, 0)
	owner := memmap.PtrPtr(0x5D4594, 534812)
	oldRecord := *owner
	official := noxflags.HasGame(noxflags.GameFlag25)
	scratch := memmap.Slice(0x85B3FC, 10307)[:0x6e]
	source := memmap.Slice(0x5D4594, 531660)[:24]
	login := memmap.Slice(0x5D4594, 534756)[:10]
	oldScratch := bytes.Clone(scratch)
	oldSource := bytes.Clone(source)
	oldLogin := bytes.Clone(login)
	t.Cleanup(func() {
		cfixture.GlobalRecordNativeSelectedC(true, selected)
		*owner = oldRecord
		copy(scratch, oldScratch)
		copy(source, oldSource)
		copy(login, oldLogin)
		if official {
			noxflags.SetGame(noxflags.GameFlag25)
		} else {
			noxflags.UnsetGame(noxflags.GameFlag25)
		}
	})
	for i := range scratch {
		scratch[i] = 0xab
	}
	cfixture.GlobalRecordNativeSelectedC(true, ^uint32(0))
	*owner = nil
	if result := nox_xxx_officialStringCmp_41FDE0(); result != 0 || scratch[0] != 0xab || scratch[0x6d] != 0xab || !bytes.Equal(scratch[1:0x6d], make([]byte, 0x6c)) {
		t.Fatalf("missing selection = %d, scratch %x", result, scratch)
	}
	cfixture.GlobalRecordNativeSelectedC(true, 0)
	if result := nox_xxx_officialStringCmp_41FDE0(); result != 1 {
		t.Fatalf("selection without record = %d", result)
	}

	record, free := alloc.NewHigh([256]byte{})
	defer free()
	*owner = unsafe.Pointer(record)
	copy(source, "testname\x00")
	copy(source[10:], "key\x00")
	copy(record[24:], "Service:Official ladder\x00")
	if result := nox_xxx_officialStringCmp_41FDE0(); result != 1 || !noxflags.HasGame(noxflags.GameFlag25) ||
		string(record[228:237]) != "testname\x00" || string(record[238:242]) != "key\x00" || string(login[:9]) != "testname\x00" {
		t.Fatalf("official record = %d, flag %t, name %q, key %q, login %q", result, noxflags.HasGame(noxflags.GameFlag25), record[228:237], record[238:242], login[:9])
	}
	copy(record[24:], "Service:official ladder\x00")
	if result := nox_xxx_officialStringCmp_41FDE0(); result != 1 || noxflags.HasGame(noxflags.GameFlag25) {
		t.Fatalf("unofficial record = %d, flag %t", result, noxflags.HasGame(noxflags.GameFlag25))
	}
}
