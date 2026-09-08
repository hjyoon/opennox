package legacy

import "testing"

func TestQuestJournalQualifyExport5009B0WritesNativeScratch(t *testing.T) {
	_, scratch := prepareQuestJournalQualifyNative5009B0(t, "War01a", 0x6d)
	if got := questJournalQualifyExportCall5009B0("Count"); got != 0 {
		t.Fatalf("export return = %d, want zero", got)
	}
	requireQuestJournalScratch5009B0(t, scratch, "War01a:Count", 0x6d)
}

func TestQuestJournalQualifyExport5009B0PreservesQualifiedReturn(t *testing.T) {
	_, scratch := prepareQuestJournalQualifyNative5009B0(t, "must-not-be-read", 0xd6)
	name := "War02a:Open"
	if got, want := questJournalQualifyExportCall5009B0(name), uint32(len(name)+1); got != want {
		t.Fatalf("export return = %d, want %d", got, want)
	}
	requireQuestJournalScratch5009B0(t, scratch, name, 0xd6)
}
