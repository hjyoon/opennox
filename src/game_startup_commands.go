package opennox

import "context"

// Both the graphical session and RunServerArgs must run these commands only
// after their initial map has loaded. Keep the existing rule/cycle/CLI order
// and let the ordinary console parser enforce the original command gates.
func execGameStartupCommands(ctx context.Context) {
	execConsoleCmdAuthed(ctx, "execrul autoexec.rul")
	if isServer {
		execServerCmd("set cycle on")
		for _, cmd := range serverExec {
			if len(cmd) == 0 {
				continue
			}
			execServerCmd(cmd)
		}
	}
}
