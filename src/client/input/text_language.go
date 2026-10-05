package input

// Language returns the language code selected by SetLanguage. Entry fields use
// it to distinguish committed IME text from the legacy scan-code text path.
func (h *Handler) Language() int {
	return h.language
}
