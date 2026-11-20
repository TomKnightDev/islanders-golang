package entities

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	shared_resources "github.com/tomknightdev/islanders-golang/resources"
)

type World struct {
	imageTile *ebiten.Image
	worldMap  shared_resources.WorldMap
	mapImage  *ebiten.Image
	Cam       *camera.Camera
}

func NewWorld(tilesImage *ebiten.Image, worldMap shared_resources.WorldMap) *World {
	w := &World{
		imageTile: tilesImage,
		worldMap:  worldMap,
	}

	return w
}

func (w *World) Update() error {
	return nil
}

// tileImage returns the 8x8 sprite for a Tiled-style gid (row*16 + col + 1).
func (w *World) tileImage(gid int) *ebiten.Image {
	const size = shared_resources.TileSize
	cols := w.imageTile.Bounds().Dx() / size
	x := (gid - 1) % cols * size
	y := (gid - 1) / cols * size
	return w.imageTile.SubImage(image.Rect(x, y, x+size, y+size)).(*ebiten.Image)
}

// renderMap draws every layer once into a single image. The map is static, so
// each frame only has to draw this one image instead of every tile.
func (w *World) renderMap() {
	const size = shared_resources.TileSize
	wm := w.worldMap
	w.mapImage = ebiten.NewImage(wm.Width*size, wm.Height*size)

	for _, layer := range wm.Layers {
		for i, gid := range layer.Data {
			if gid == shared_resources.TileNone {
				continue
			}
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(i%wm.Width*size), float64(i/wm.Width*size))
			w.mapImage.DrawImage(w.tileImage(gid), opts)
		}
	}
}

// SetTile changes one tile, such as a rock depleting, and redraws it.
func (w *World) SetTile(x, y, layer, gid int) {
	wm := w.worldMap
	if x < 0 || y < 0 || x >= wm.Width || y >= wm.Height || layer >= len(wm.Layers) {
		return
	}
	i := y*wm.Width + x
	wm.Layers[layer].Data[i] = gid
	if w.mapImage == nil {
		return // drawn with the rest of the map on first Draw
	}
	// Ground tiles are opaque, so redrawing every layer at this spot
	// replaces whatever was there.
	const size = shared_resources.TileSize
	for _, l := range wm.Layers {
		if l.Data[i] == shared_resources.TileNone {
			continue
		}
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(x*size), float64(y*size))
		w.mapImage.DrawImage(w.tileImage(l.Data[i]), opts)
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	if w.mapImage == nil {
		w.renderMap()
	}

	w.Cam.Surface.Clear()
	w.Cam.Surface.DrawImage(w.mapImage, w.Cam.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0))
}

// DrawLabel names the villages and mines on the map.
func (w *World) DrawLabel(screen *ebiten.Image) {
	for _, p := range w.worldMap.Places {
		drawPlaceLabel(screen, w.Cam, p)
	}
}
