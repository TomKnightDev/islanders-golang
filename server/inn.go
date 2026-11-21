package main

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/tomknightdev/islanders-golang/resources"
)

// maxRumours is how many rumours an inn tells at once.
const maxRumours = 3

// rumours are directions from a village to the nearest places the player
// hasn't found yet.
func (w *World) rumours(p *player, village string) []string {
	from := w.place(village)
	if from == nil {
		return nil
	}
	var unknown []resources.Place
	for _, pl := range w.wm.Places {
		if !p.found[pl.Name] {
			unknown = append(unknown, pl)
		}
	}
	dist := func(pl resources.Place) float64 { return math.Hypot(float64(pl.X-from.X), float64(pl.Y-from.Y)) }
	slices.SortFunc(unknown, func(a, b resources.Place) int { return int(dist(a) - dist(b)) })
	var out []string
	for _, pl := range unknown[:min(maxRumours, len(unknown))] {
		out = append(out, rumour(*from, pl))
	}
	return out
}

// rumour is one line of directions: "Travellers speak of a mine called
// Greyrock, about 40 tiles north-east of here, where copper and iron are
// mined."
func rumour(from, to resources.Place) string {
	dx, dy := float64(to.X-from.X), float64(to.Y-from.Y)
	tiles := int(math.Round(math.Hypot(dx, dy)/5)) * 5
	what := map[resources.PlaceKind]string{
		resources.PlaceVillage: "a village",
		resources.PlaceMine:    "a mine",
		resources.PlaceForest:  "a forest",
		resources.PlaceFarm:    "a farm",
		resources.PlaceCamp:    "a camp",
	}[to.Kind]
	line := fmt.Sprintf("Travellers speak of %s called %s, about %d tiles %s of here", what, to.Name, max(tiles, 5), compass(dx, dy))
	if len(to.Materials) > 0 {
		var names []string
		for _, m := range to.Materials {
			names = append(names, materialName(to.Kind, m))
		}
		switch to.Kind {
		case resources.PlaceMine:
			line += ", where " + list(names) + " ore is mined"
		case resources.PlaceCamp:
			line += ", where " + list(names) + " gather"
		case resources.PlaceForest:
			line += ", full of " + list(names) + " trees"
		case resources.PlaceFarm:
			line += ", growing " + list(names)
		}
	}
	return line + "."
}

// compass is the direction of (dx, dy) on the map, where y grows south.
func compass(dx, dy float64) string {
	dirs := []string{"east", "south-east", "south", "south-west", "west", "north-west", "north", "north-east"}
	a := math.Atan2(dy, dx) // 0 is east, positive is south
	i := int(math.Round(a/(math.Pi/4))+8) % 8
	return dirs[i]
}

// materialName is how a place's material reads in a rumour.
func materialName(kind resources.PlaceKind, id string) string {
	if kind == resources.PlaceCamp {
		if e, ok := resources.FindEnemy(id); ok {
			return strings.ToLower(e.Plural)
		}
	}
	return id
}

// list joins names as "a", "a and b" or "a, b and c".
func list(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}
