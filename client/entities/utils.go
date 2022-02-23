package entities

func RemoveEntity(game *main.Game, i int) {
	if i >= len(game.Entities) {
		return
	}

	game.Entities[i] = game.Entities[len(game.Entities)-1]
	game.Entities = game.Entities[:len(game.Entities)-1]

	// return append(scenes[:s], scenes[s+1:]...)
}
