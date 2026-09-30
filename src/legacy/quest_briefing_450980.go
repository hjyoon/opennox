package legacy

const questStageBriefingSource450980 = `C:\NoxPost\src\client\Gui\GUIBrief.c`

type questStageBriefingHooks450980[P, I, T any] struct {
	questStartBriefingHooks450A30[P, I, T]
	flags func(P) uint8
}

// questStageBriefing450980 implements the complete sealed 00450980..00450A28
// body. Stage and flags remain live reads after the text and stage setters.
// The empty text address and final lock flags differ from the start briefing.
func questStageBriefing450980[P, I, T any](packet P, show int32, h questStageBriefingHooks450980[P, I, T]) int32 {
	h.storeState(0)
	h.resetParticles()
	h.hideBook(1)
	h.prepare()
	h.setImage(h.loadImage(h.offset(packet, 5)))
	key := h.offset(packet, 37)
	if h.stringLength(key) != 0 {
		h.setText(h.loadText(key, questStageBriefingSource450980, 1714))
	} else {
		h.setText(h.emptyText())
	}
	h.setStage(uint32(h.stage(packet)))
	if h.flags(packet)&2 != 0 {
		h.storeState(1)
	}
	if show != 0 {
		return h.lock(254, 1, 2)
	}
	return show
}
