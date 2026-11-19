package legacy

import "github.com/opennox/libs/ifs"

func saveAccessList4E43F0(path string) bool {
	file, err := ifs.Create(path)
	if err != nil {
		return false
	}
	defer file.Close()

	file.WriteString("[Banned Users]\n")
	for p := sub416900Native(); p != nil; p = sub416910Native(p) {
		entry := (*accessBanned416720)(p)
		if entry.Expires != 0 {
			continue
		}
		file.WriteString(accessListNameLine4E43F0(&entry.Name))
		serial := GoStringS(entry.Serial[:])
		if serial == "" {
			serial = "0"
		}
		file.WriteString(serial + "\n")
	}
	file.WriteString("\n[Allowed Users]\n")
	for p := sub4168E0Native(); p != nil; p = sub4168F0Native(p) {
		entry := (*accessAllowed416720)(p)
		file.WriteString(accessListNameLine4E43F0(&entry.Name))
	}
	// The original result reflects whether opening for writing succeeded.
	return true
}

// GAME.EXE's default CRT locale writes one byte per name code unit and stops
// at the first NUL or unrepresentable unit. The line separator is still written.
func accessListNameLine4E43F0(name *[26]uint16) string {
	var line [27]byte
	n := 0
	for _, unit := range name {
		if unit == 0 || unit > 0xff {
			break
		}
		line[n] = byte(unit)
		n++
	}
	line[n] = '\n'
	return string(line[:n+1])
}
