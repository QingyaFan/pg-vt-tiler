package generator

import (
	"log"
	"testing"
)

func TestGenerateTile(t *testing.T) {
	// generateTile(8, 208, 100, `water`, `geom`) // 0.5s
	// generateTile(7, 106, 46, `railways_3857`, `geom`) // 1.8s
	// generateTile(6, 52, 24, `water`, `geom`)  // 58s
	// generateTile(4, 4, 1, `railways_3857`, `geom`) // s
}

func TestBoxToArray(t *testing.T) {
	// boxToArray(`BOX(8155154.57602443 1865495.57173284,15038985.6866807 7087842.63996618)`)
}

func TestGetDataExtent(t *testing.T) {

}

func TestMain(t *testing.T) {
	err := Generate(`railways_3857`, `geom`)
	if err != nil {
		log.Fatal(`生成出错`)
	}
}
