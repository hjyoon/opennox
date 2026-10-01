package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

func TestOptionsAuditListFirst(t *testing.T) {
	items := []gui.ScrollListBoxItem{{Field_0: 18}, {Field_0: 36}, {Field_0: 54}}
	data := gui.ScrollListBoxData{Items: &items[0], Field_11_0: uint16(len(items))}
	win := gui.Window{WidgetData: unsafe.Pointer(&data)}
	for _, tc := range []struct {
		offset uint16
		want   int
	}{{0, 0}, {17, 0}, {18, 1}, {19, 1}, {35, 1}, {36, 2}, {53, 2}, {54, 0}} {
		data.Field_13_1 = tc.offset
		if got := optionsAuditListFirst(&win); got != tc.want {
			t.Errorf("offset=%d: first=%d, want %d", tc.offset, got, tc.want)
		}
	}
	for _, win := range []*gui.Window{nil, {}, {WidgetData: unsafe.Pointer(&gui.ScrollListBoxData{})}} {
		if got := optionsAuditListFirst(win); got != 0 {
			t.Errorf("empty list first=%d, want 0", got)
		}
	}
}

func TestOptionsAuditListTextBounds(t *testing.T) {
	items := []gui.ScrollListBoxItem{{Text: [256]uint16{'F', '1', '0'}}}
	data := gui.ScrollListBoxData{Items: &items[0], Field_11_0: 1}
	win := gui.Window{WidgetData: unsafe.Pointer(&data)}
	for _, tc := range []struct {
		index int
		want  string
	}{{-1, ""}, {0, "F10"}, {1, ""}, {2, ""}} {
		if got := optionsAuditListText(&win, tc.index); got != tc.want {
			t.Errorf("index=%d: text=%q, want %q", tc.index, got, tc.want)
		}
	}
	if got := optionsAuditListText(nil, 0); got != "" {
		t.Errorf("nil list text=%q, want empty", got)
	}
}

func TestAuditClientOptionsRejectsInvalidMode(t *testing.T) {
	for _, mode := range []int{-1, 2, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("mode=%d did not panic", mode)
				}
			}()
			new(e2eScenario).AuditClientOptions(mode, "invalid test mode")
		}()
	}
}

func TestOptionsAuditInputExitControl(t *testing.T) {
	for _, tc := range []struct {
		mode int
		id   uint
		name string
	}{{0, 152, "Back"}, {1, 932, "Apply"}} {
		id, name := optionsAuditInputExitControl(tc.mode)
		if id != tc.id || name != tc.name {
			t.Errorf("mode=%d: exit=(%d, %q), want (%d, %q)", tc.mode, id, name, tc.id, tc.name)
		}
	}
	for _, mode := range []int{-1, 2, 99} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("mode=%d did not panic", mode)
				}
			}()
			optionsAuditInputExitControl(mode)
		}()
	}
}
