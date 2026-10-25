package item

// FilledMap is a map item. The client looks up the pixels of a map by its ID; the server supplies them with a
// ClientBoundMapItemData packet.
type FilledMap struct {
	// ID is the unique ID of the map, stored in the item's NBT as map_uuid.
	ID int64
}

// EncodeItem ...
func (FilledMap) EncodeItem() (name string, meta int16) {
	return "minecraft:filled_map", 0
}

// EncodeNBT ...
func (m FilledMap) EncodeNBT() map[string]any {
	return map[string]any{"map_uuid": m.ID}
}

// DecodeNBT ...
func (m FilledMap) DecodeNBT(data map[string]any) any {
	if v, ok := data["map_uuid"].(int64); ok {
		m.ID = v
	}
	return m
}
