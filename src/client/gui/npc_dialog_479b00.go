package gui

// NPCDialogRuntime479B00 supplies the original dialog button side effects.
// Repeat must load the current filename when invoked, after the click sound;
// the PE32 file pointer is not a DWORD on native 64-bit hosts.
type NPCDialogRuntime479B00 struct {
	Paused     func() int
	ClickSound func()
	Done       func()
	Repeat     func()
	Answer     func(byte)
}

// NPCDialogProc479B00 preserves GAME.EXE 00479B00..00479BDF. The button ID
// is cached before the pause query, every admitted click plays the shell
// sound (including an unknown button), and all branches return zero.
func NPCDialogProc479B00(event int, button *Window, r NPCDialogRuntime479B00) int {
	if event != 0x4007 {
		return 0
	}
	id := button.ID()
	if r.Paused() != 0 {
		return 0
	}
	r.ClickSound()
	switch id {
	case 3906:
		r.Done()
	case 3907:
		r.Repeat()
	case 3908:
		r.Answer(1)
	case 3909:
		r.Answer(2)
	}
	return 0
}
