package generator

import "testing"

func TestGenerateTile(t *testing.T) {
	getDataExtent(`water`, `geom`)
}

func TestBoxToArray(t *testing.T) {
	boxToArray(`BOX(8155154.57602443 1865495.57173284,15038985.6866807 7087842.63996618)`)
}

func TestGetDataExtent(t *testing.T) {

}
