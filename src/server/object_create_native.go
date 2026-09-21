package server

// ObeliskCreateNative54CA10 restores GAME.EXE 0054CA10 without passing the
// live Object through the legacy callback boundary.
func ObeliskCreateNative54CA10(obj *Object) {
	update := (*ObeliskUpdateData)(obj.UpdateData)
	update.Mana = 50
	obj.NeedSync()
}

// AnimCreateNative54CA50 restores GAME.EXE 0054CA50.
func AnimCreateNative54CA50(obj *Object) {
	obj.SetXStatus(2)
}

// TriggerCreateNative54CA60 restores GAME.EXE 0054CA60 against the fixed-width
// TriggerUpdateData record.
func TriggerCreateNative54CA60(obj *Object) {
	update := (*TriggerUpdateData)(obj.UpdateData)
	update.Colors = [6]uint8{90, 90, 90, 10, 10, 10}
}

// RewardMarkerCreateNative54CAC0 restores GAME.EXE 0054CAC0 against the
// pointer-independent RewardMarkerInitData record.
func RewardMarkerCreateNative54CAC0(obj *Object) {
	data := (*RewardMarkerInitData)(obj.InitData)
	data.CategoryMask = 255
	data.ChanceMode = 0
}
