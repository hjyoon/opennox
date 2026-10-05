package opennox

import (
	"fmt"
	"image"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/client/seat"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type e2eNPCDialogPlayback479B00 struct {
	opened, queued, processed uint64
}

type e2eNPCDialogRepeatFixture479B00 struct {
	root    *gui.Window
	filePtr unsafe.Pointer
	file    string
	text    string
	choice  byte
	before  e2eNPCDialogPlayback479B00
}

func e2eNPCDialogText479B00(root *gui.Window) (string, error) {
	title, list := root.ChildByID(3910), root.ChildByID(3901)
	if title == nil || list == nil || list.WidgetData == nil {
		return "", fmt.Errorf("stock dialog title/list is missing")
	}
	d := (*gui.ScrollListBoxData)(list.WidgetData)
	if d.Items == nil || d.Field_11_0 == 0 || d.Field_11_0 > d.Count {
		return "", fmt.Errorf("stock dialog list has invalid population %d/%d", d.Field_11_0, d.Count)
	}
	parts := []string{title.DrawData().Text()}
	for _, item := range unsafe.Slice(d.Items, int(d.Field_11_0)) {
		parts = append(parts, alloc.GoString16(&item.Text[0]))
	}
	return strings.Join(parts, "\n"), nil
}

func (f *e2eNPCDialogRepeatFixture479B00) snapshot() error {
	f.root = legacy.Get_dword_5d4594_1123524()
	if f.root == nil || f.root.GetFlags().IsHidden() {
		return fmt.Errorf("stock dialog is not visible")
	}
	button := f.root.ChildByID(3907)
	if button == nil || button.GetFlags().IsHidden() || !button.GetFlags().IsEnabled() {
		return fmt.Errorf("stock Repeat button is unavailable")
	}
	f.filePtr = *memmap.PtrPtr(0x5D4594, 1115312)
	f.file = alloc.GoString((*byte)(f.filePtr))
	if f.file == "" || legacy.Dialogs.State() != 3 || legacy.Dialogs.CurrentPlayingFile() != f.file || legacy.Dialogs.FileToRead() != f.file || legacy.Dialogs.GetStream() == 0 {
		return fmt.Errorf("stock opening voice is not playing the dialog filename %q", f.file)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(f.filePtr) <= math.MaxUint32 || uintptr(f.root.C()) <= math.MaxUint32 || uintptr(button.C()) <= math.MaxUint32) {
		return fmt.Errorf("stock filename/window/button did not exceed 4 GiB: %p/%p/%p", f.filePtr, f.root, button)
	}
	var err error
	if f.text, err = e2eNPCDialogText479B00(f.root); err != nil {
		return err
	}
	f.choice = *memmap.PtrUint8(0x5D4594, 1123516)
	if f.before, err = e2eNPCDialogAudioStats479B00(); err != nil {
		return err
	}
	e2eLog.Printf("NPC REPEAT STOCK: root=%p button=%p filename=%p file=%q choice=%d text-bytes=%d playback=%+v", f.root, button, f.filePtr, f.file, f.choice, len(f.text), f.before)
	return nil
}

func (f *e2eNPCDialogRepeatFixture479B00) surface() error {
	root := legacy.Get_dword_5d4594_1123524()
	if root != f.root || root == nil || root.GetFlags().IsHidden() {
		return fmt.Errorf("Repeat changed/closed the NPC dialog")
	}
	button := root.ChildByID(3907)
	if button == nil || button.GetFlags().IsHidden() || !button.GetFlags().IsEnabled() {
		return fmt.Errorf("Repeat control no longer available")
	}
	text, err := e2eNPCDialogText479B00(root)
	if err != nil {
		return err
	}
	if text != f.text || *memmap.PtrUint8(0x5D4594, 1123516) != f.choice || *memmap.PtrPtr(0x5D4594, 1115312) != f.filePtr {
		return fmt.Errorf("Repeat changed the dialog text, choice, or native filename")
	}
	return nil
}

func (f *e2eNPCDialogRepeatFixture479B00) ended() bool {
	return legacy.Dialogs.State() == 0 && legacy.Dialogs.GetStream() == 0 && legacy.Dialogs.FileToRead() == ""
}

func (f *e2eNPCDialogRepeatFixture479B00) finish(replay int) {
	if err := f.surface(); err != nil {
		e2eError(err)
		return
	}
	stats, err := e2eNPCDialogAudioStats479B00()
	if err != nil {
		e2eError(err)
		return
	}
	wantOpened := f.before.opened
	if replay != 0 {
		wantOpened++
	}
	if stats.opened != wantOpened || stats.processed <= f.before.processed || stats.queued < stats.processed {
		e2eError(fmt.Errorf("dialog %d did not naturally consume its voice buffers: before=%+v after=%+v", replay, f.before, stats))
		return
	}
	f.before = stats
	e2eLog.Printf("NPC REPEAT NATURAL END: replay=%d file=%q state=%d stream=%#x playback=%+v text-unchanged=true choice-unchanged=true", replay, f.file, legacy.Dialogs.State(), legacy.Dialogs.GetStream(), stats)
}

func (f *e2eNPCDialogRepeatFixture479B00) mouse() {
	if err := f.surface(); err != nil {
		e2eError(err)
		return
	}
	button := f.root.ChildByID(3907)
	size := button.Size()
	pos := button.GlobalPos().Add(image.Pt(size.X/2, size.Y/2))
	e2eLog.Printf("NPC REPEAT MOUSE: point=%v button=%p", pos, button)
	e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
}

func (f *e2eNPCDialogRepeatFixture479B00) playing() bool {
	stats, err := e2eNPCDialogAudioStats479B00()
	return err == nil && stats.opened == f.before.opened+1 && stats.queued > f.before.queued && legacy.Dialogs.State() == 3 && legacy.Dialogs.GetStream() != 0 && legacy.Dialogs.GetStream().Status() == 4 && legacy.Dialogs.FileToRead() == f.file && legacy.Dialogs.CurrentPlayingFile() == f.file
}

func (f *e2eNPCDialogRepeatFixture479B00) samePlaying(replay int, busy bool) {
	if err := f.surface(); err != nil {
		e2eError(err)
		return
	}
	if !f.playing() {
		e2eError(fmt.Errorf("mouse Repeat %d did not play the same stock file (busy=%t)", replay, busy))
		return
	}
	stats, _ := e2eNPCDialogAudioStats479B00()
	e2eLog.Printf("NPC REPEAT PLAYING: replay=%d busy=%t file=%q state=%d stream=%#x playback=%+v", replay, busy, f.file, legacy.Dialogs.State(), legacy.Dialogs.GetStream(), stats)
}

// CheckNPCDialogRepeat uses only stock campaign navigation and real mouse
// events. No playback request, stream, timer, counter, dialog text, or reply
// result is supplied by this observer. The OpenAL backend may use its null
// sink for headless testing; this is not proof of audible hardware output.
func (sc *e2eScenario) CheckNPCDialogRepeat(count int, name string) {
	if count < 1 || count > 2 {
		panic("NPC Repeat count must be 1 or 2")
	}
	var f e2eNPCDialogRepeatFixture479B00
	sc.add(0, name+" require native audio", func() {
		if _, err := e2eNPCDialogAudioStats479B00(); err != nil {
			e2eError(err)
		}
	})
	sc.addWhen(0, name+" stock opening voice", 6000, func() bool {
		root := legacy.Get_dword_5d4594_1123524()
		return root != nil && !root.GetFlags().IsHidden() && legacy.Dialogs != nil && legacy.Dialogs.State() == 3 && legacy.Dialogs.GetStream() != 0 && legacy.Dialogs.FileToRead() != ""
	}, func() {
		if err := f.snapshot(); err != nil {
			e2eError(err)
		}
	})
	sc.CaptureMagicFrame(name + " original dialog")
	sc.addWhen(0, name+" original natural end", 30000, f.ended, func() { f.finish(0) })
	for replay := 1; replay <= count; replay++ {
		label := fmt.Sprintf("%s replay %d", name, replay)
		sc.add(0, label+" mouse move", f.mouse)
		sc.Input(1, label+" mouse press", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		sc.Input(1, label+" mouse release", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.addWhen(1, label+" native stream", 1500, f.playing, func() { f.samePlaying(replay, false) })
		sc.CaptureMagicFrame(label + " actual repeated voice")
		// A second click during this same stream must preserve the original
		// PlayFile same-filename behavior, rather than inventing a rewind.
		sc.add(0, label+" busy mouse move", f.mouse)
		sc.Input(1, label+" busy mouse press", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		sc.Input(1, label+" busy mouse release", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.add(20, label+" busy same-file check", func() { f.samePlaying(replay, true) })
		sc.addWhen(0, label+" natural end", 30000, f.ended, func() { f.finish(replay) })
	}
	sc.ClickNPCDialogDone(name + " actual Done")
	sc.addWhen(0, name+" actual close", 1200, func() bool {
		root := legacy.Get_dword_5d4594_1123524()
		return root == f.root && root.GetFlags().IsHidden()
	}, func() { e2eLog.Printf("NPC REPEAT COMPLETE: replays=%d file=%q Done-closed=true", count, f.file) })
}
