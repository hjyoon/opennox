package legacy

/*
#include "GAME3_2.h"
#include "GAME5_2.h"
*/
import "C"

import "strings"

func loadMapRules4D0550(filename string, readRules func(string) bool) bool {
	if nul := strings.IndexByte(filename, 0); nul >= 0 {
		filename = filename[:nul]
	}
	if len(filename) < 4 {
		return false
	}
	stem := filename[:len(filename)-4]
	userPrefix := stem
	if slash := strings.LastIndexByte(stem, '\\'); slash > 0 {
		userPrefix = stem[:slash+1]
	}
	if !readRules(userPrefix + "user.rul") {
		readRules(stem + ".rul")
	}
	return true
}

func Sub_4D0550(filename string) {
	loadMapRules4D0550(filename, readMapRuleFile4D0670)
}

//export sub_4D0550
func sub_4D0550(name *C.char) C.int {
	if loadMapRules4D0550(GoString(name), readMapRuleFile4D0670) {
		return 1
	}
	return 0
}

//export sub_57A950
func sub_57A950(name *C.char) C.int {
	filename := GoString(name)
	if len(filename) < 4 {
		return 0
	}
	path := "maps\\" + filename[:len(filename)-4] + "\\" + filename
	if loadMapRules4D0550(path, readMapRuleFile4D0670) {
		return 1
	}
	return 0
}
