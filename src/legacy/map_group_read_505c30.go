package legacy

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const mapGroupNameCapacity505C30 = 76

type mapGroupReadHooks505C30 struct {
	currentMap string
	skip       bool
	addGroup   func(name string, index uint32, kind server.MapGroupKind)
	addItem    func(index uint32, kind server.MapGroupKind, item mapGroupItemRecord505C30)
}

func mapReadGroups505C30(cf *cryptfile.CryptFile, hooks mapGroupReadHooks505C30) error {
	if cf == nil {
		return fmt.Errorf("map group read: nil crypt file")
	}
	if !cf.ReadOnly() {
		return fmt.Errorf("map group read: crypt file is not read-only")
	}

	rawVersion, err := cf.ReadU16()
	if err != nil {
		return fmt.Errorf("map group version: %w", err)
	}
	version := int16(rawVersion)
	if version > 3 {
		return fmt.Errorf("unsupported map group version: %d", version)
	}
	rawCount, err := cf.ReadU32()
	if err != nil {
		return fmt.Errorf("map group count: %w", err)
	}
	count := int32(rawCount)
	if count <= 0 {
		return nil
	}

	for groupIndex := int32(0); groupIndex < count; groupIndex++ {
		name, err := mapReadGroupName505C30(cf)
		if err != nil {
			return fmt.Errorf("map group %d name: %w", groupIndex, err)
		}
		name, err = mapQualifyGroupName505C30(version, hooks.currentMap, name)
		if err != nil {
			return fmt.Errorf("map group %d name: %w", groupIndex, err)
		}
		rawKind, err := cf.ReadU8()
		if err != nil {
			return fmt.Errorf("map group %d type: %w", groupIndex, err)
		}
		kind := server.MapGroupKind(rawKind)
		index, err := cf.ReadU32()
		if err != nil {
			return fmt.Errorf("map group %d index: %w", groupIndex, err)
		}
		if !hooks.skip && hooks.addGroup != nil {
			hooks.addGroup(name, index, kind)
		}

		rawItemCount, err := cf.ReadU32()
		if err != nil {
			return fmt.Errorf("map group %d item count: %w", groupIndex, err)
		}
		itemCount := int32(rawItemCount)
		for itemIndex := int32(0); itemIndex < itemCount; itemIndex++ {
			item, err := mapReadGroupItem505C30(cf, kind)
			if err != nil {
				return fmt.Errorf("map group %d item %d: %w", groupIndex, itemIndex, err)
			}
			if !hooks.skip && hooks.addItem != nil {
				hooks.addItem(index, kind, item)
			}
		}
	}
	return nil
}

func mapReadGroupName505C30(cf *cryptfile.CryptFile) (string, error) {
	sz, err := cf.ReadU8()
	if err != nil {
		return "", err
	}
	if sz > mapGroupNameCapacity505C30 {
		return "", fmt.Errorf("name record is too long: %d bytes", sz)
	}
	buf := make([]byte, int(sz))
	if _, err := io.ReadFull(cf, buf); err != nil {
		return "", err
	}
	if i := bytes.IndexByte(buf, 0); i >= 0 {
		buf = buf[:i]
	} else if sz == mapGroupNameCapacity505C30 {
		return "", fmt.Errorf("name record fills native buffer without a terminator")
	}
	return string(buf), nil
}

func mapQualifyGroupName505C30(version int16, currentMap, name string) (string, error) {
	switch {
	case version < 2:
		if len(currentMap)+1+len(name) >= 0x35 {
			return "", fmt.Errorf("qualified name is too long")
		}
		return currentMap + ".map:" + name, nil
	case version == 2:
		name = mapGroupSecondToken505C30(name)
		if name == "" {
			return "", fmt.Errorf("version 2 name has no second component")
		}
		if len(currentMap)+1+len(name) >= 0x35 {
			return "", fmt.Errorf("qualified name is too long")
		}
		return currentMap + ":" + name, nil
	default:
		return name, nil
	}
}

func mapGroupSecondToken505C30(s string) string {
	var token int
	for len(s) != 0 {
		s = strings.TrimLeft(s, ":")
		if len(s) == 0 {
			return ""
		}
		end := strings.IndexByte(s, ':')
		if end < 0 {
			end = len(s)
		}
		token++
		if token == 2 {
			return s[:end]
		}
		s = s[end:]
	}
	return ""
}

func mapReadGroupItem505C30(cf *cryptfile.CryptFile, kind server.MapGroupKind) (mapGroupItemRecord505C30, error) {
	var item mapGroupItemRecord505C30
	switch kind {
	case server.MapGroupObjects, server.MapGroupWaypoints, server.MapGroupGroups:
		v, err := cf.ReadU32()
		if err != nil {
			return item, err
		}
		item.raw0 = v
	case server.MapGroupWalls:
		v0, err := cf.ReadU32()
		if err != nil {
			return item, err
		}
		v1, err := cf.ReadU32()
		if err != nil {
			return item, err
		}
		item.raw0 = v0
		item.raw4 = v1
	default:
		return item, fmt.Errorf("invalid map group type: %d", kind)
	}
	return item, nil
}
