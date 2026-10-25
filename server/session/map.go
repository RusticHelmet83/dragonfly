package session

import (
	"image/color"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SendMapData sends the pixels of the map with the ID passed to the client. pixels must hold exactly 128*128
// colours, row by row, and every colour should have its alpha set to 255.
func (s *Session) SendMapData(id int64, pixels []color.RGBA) {
	s.writePacket(&packet.ClientBoundMapItemData{
		MapID:          id,
		Scale:          protocol.Option[byte](0),
		MapsIncludedIn: protocol.Option([]int64{id}),
		Width:          protocol.Option[int32](128),
		Height:         protocol.Option[int32](128),
		XOffset:        protocol.Option[int32](0),
		YOffset:        protocol.Option[int32](0),
		Pixels:         protocol.Option(pixels),
	})
}
