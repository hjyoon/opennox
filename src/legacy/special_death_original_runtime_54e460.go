package legacy

import "github.com/opennox/opennox/v1/server"

// Only the independent original-output tests replace these boundaries. The
// complete consumer is compiled and frozen while all three C ABIs are still
// PE32 and the raw native implementation is absent.
var specialDeathOriginalStateOverride54E460 *server.GolemDeathState549E90
var specialDeathOriginalRuntimeOverride54E460 *server.SpecialDeathRuntime54E460
