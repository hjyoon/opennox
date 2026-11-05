package legacy

import "github.com/opennox/opennox/v1/server"

// Independent reference runtimes are installed only by the frozen tests.
// Production leaves both overrides nil and uses the original world services.
var skullRawOriginalStateOverride54F9A0 *server.SkullRawState54F9A0
var skullRawOriginalRuntimeOverride54F9A0 *server.SkullRawRuntime54F9A0
