package legacy

/*
#include "defs.h"
*/
import "C"

func questNoticeNamePending48EA70() int {
	return int(legacyWords.dword_5d4594_1200832)
}

func questNoticeGuideFixture48EA70(title string) func() {
	old := fieldGuides[1]
	fieldGuides[1] = fieldGuideEntry{thingType: 1, name: internWStr(title)}
	return func() { fieldGuides[1] = old }
}
