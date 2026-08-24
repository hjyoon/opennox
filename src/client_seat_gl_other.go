//go:build !darwin && !server

package opennox

func prepareSeatOpenGL() error {
	return nil
}
