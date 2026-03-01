/*
Package decoration provides procedural room decoration and environmental storytelling
for dungeon generation in the Violence game.

The system assigns semantic room types (Armory, Library, Shrine, etc.) to BSP-generated
rooms based on size, position, and genre. Each room type receives genre-appropriate
decorations including:

  - Landmarks: Central focal points (altars, terminals, treasure chests)
  - Furniture: Wall-adjacent elements (crates, shelves, workbenches)
  - Obstacles: Blocking environmental hazards (rubble, columns, machinery)
  - Details: Non-blocking visual elements (debris, plants, cables)

Decoration density and variety are genre-aware. Fantasy dungeons emphasize shrines
and libraries with stone furniture. SciFi stations favor laboratories and storage
with metallic props. Horror settings use sparse, decayed decorations.

Integration:

The decoration system integrates with BSP level generation. After rooms are carved,
the system determines room types and places decorations:

	decorSys := decoration.NewSystem()
	decorSys.SetGenre("fantasy")

	rooms := bsp.GetRooms(bspTree)
	for i, room := range rooms {
		roomType := decorSys.DetermineRoomType(room.W, room.H, i, len(rooms), rng)
		decor := decorSys.DecorateRoom(roomType, room.X, room.Y, room.W, room.H, tiles, rng)
		// Store decoration data for rendering
	}

Decorations use seeded sprite generation to ensure deterministic visuals across
multiple playthroughs of the same seed.
*/
package decoration
