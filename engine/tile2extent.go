package engine

import (
	"fmt"
	"math"
)

const (
	tileSize = 256
	half = 20037508.342789244
)



func tileToExtent(z, x, y int) [4]float64 {

	tileWidth := half * 2 / math.Pow(2, float64(z))

	// calculate x lower and high coord
	xMin := float64(x) * tileWidth - half
	xMax := float64(x + 1) * tileWidth - half

	// calculate y lower and high coord
	yMin := half - float64(y + 1) * tileWidth
	yMax := half - float64(y) * tileWidth

	return [4]float64{xMin, yMin, xMax, yMax}
}