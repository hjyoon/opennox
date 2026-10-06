package opennox

// completeServerSessionStartup is called only after all session resources and
// the game listener have initialized. The Go HTTP/NAT services are the final
// step: a failure must release that session before the main loop retries, while
// its objects still belong to the current allocation classes. The caller then
// restores the selected map/mode, which normal teardown resets for the menu.
func completeServerSessionStartup(startServices func() error, endSession, restoreStartup func()) error {
	if err := startServices(); err != nil {
		endSession()
		restoreStartup()
		return err
	}
	return nil
}
