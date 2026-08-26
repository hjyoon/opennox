package legacy

import (
	"fmt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type mapGroupItemRecord505C30 struct {
	raw0 uint32
	raw4 uint32
}

type mapGroupRecord505C30 struct {
	kind  server.MapGroupKind
	index uint32
	name  string
	items []mapGroupItemRecord505C30
}

func mapWriteGroups505C30(cf *cryptfile.CryptFile, first *server.MapGroup) error {
	var records []mapGroupRecord505C30
	for group := first; group != nil; group = group.Next() {
		record := mapGroupRecord505C30{
			kind:  group.GroupType(),
			index: group.Index(),
			name:  group.ID(),
		}
		for item := group.First(); item != nil; item = item.Next() {
			record.items = append(record.items, mapGroupItemRecord505C30{
				raw0: item.Raw0,
				raw4: item.Raw4,
			})
		}
		records = append(records, record)
	}
	return mapWriteGroupRecords505C30(cf, records)
}

func mapWriteGroupRecords505C30(cf *cryptfile.CryptFile, records []mapGroupRecord505C30) error {
	if err := cf.WriteU16(3); err != nil {
		return err
	}
	if err := cf.WriteU32(uint32(len(records))); err != nil {
		return err
	}
	for _, group := range records {
		if len(group.name)+1 > 0xff {
			return fmt.Errorf("map group name is too long: %d bytes", len(group.name))
		}
		if err := cf.WriteU8(byte(len(group.name) + 1)); err != nil {
			return err
		}
		if _, err := cf.Write([]byte(group.name)); err != nil {
			return err
		}
		if err := cf.WriteU8(0); err != nil {
			return err
		}
		if err := cf.WriteU8(byte(group.kind)); err != nil {
			return err
		}
		if err := cf.WriteU32(group.index); err != nil {
			return err
		}
		if err := cf.WriteU32(uint32(len(group.items))); err != nil {
			return err
		}
		for _, item := range group.items {
			switch group.kind {
			case server.MapGroupObjects, server.MapGroupWaypoints, server.MapGroupGroups:
				if err := cf.WriteU32(item.raw0); err != nil {
					return err
				}
			case server.MapGroupWalls:
				if err := cf.WriteU32(item.raw0); err != nil {
					return err
				}
				if err := cf.WriteU32(item.raw4); err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid map group type: %d", group.kind)
			}
		}
	}
	return nil
}
