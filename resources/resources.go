package resources

type WorldMap struct {
	Compressionlevel int     `json:"compressionlevel"`
	Height           int     `json:"height"`
	Infinite         bool    `json:"infinite"`
	Layers           []Layer `json:"layers"`
	Nextlayerid      int     `json:"nextlayerid"`
	Nextobjectid     int     `json:"nextobjectid"`
	Orientation      string  `json:"orientation"`
	Renderorder      string  `json:"renderorder"`
	Tiledversion     string  `json:"tiledversion"`
	Tileheight       int     `json:"tileheight"`
	Tilesets         []struct {
		Firstgid int    `json:"firstgid"`
		Source   string `json:"source"`
	} `json:"tilesets"`
	Tilewidth int    `json:"tilewidth"`
	Type      string `json:"type"`
	Version   string `json:"version"`
	Width     int    `json:"width"`
}

type Layer struct {
	Data    []int  `json:"data"`
	Height  int    `json:"height"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Opacity int    `json:"opacity"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
	Width   int    `json:"width"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

func WorldMapWebSocketMessageConvert(message map[string]interface{}) *WorldMap {
	return &WorldMap{
		Height: int(message["height"].(float64)),
		Width:  int(message["width"].(float64)),
		Layers: LayersWebSocketMessageConvert(message["layers"].([]interface{})),
	}
}

func LayersWebSocketMessageConvert(message []interface{}) []Layer {
	var layers = []Layer{}
	for _, m := range message {
		layer := m.(map[string]interface{})

		layers = append(layers, Layer{
			Data: convertToInt(layer["data"].([]interface{})),
		})
	}

	return layers
}

func convertToInt(floats []interface{}) []int {
	ints := []int{}

	for _, f := range floats {
		ints = append(ints, int(f.(float64)))
	}

	return ints
}
