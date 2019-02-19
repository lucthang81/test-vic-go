package components

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type DominoesDeck struct {
	tiles map[string]bool
}

func (deck *DominoesDeck) Tiles() map[string]bool {
	return deck.tiles
}

func NewDominoesDeck() *DominoesDeck {
	deck := &DominoesDeck{
		tiles: make(map[string]bool),
	}

	for i := 0; i <= 6; i++ {
		for j := 0; j <= i; j++ {
			deck.tiles[TileString(fmt.Sprintf("%d", j), fmt.Sprintf("%d", i))] = true
		}
	}
	return deck
}

func TileString(value1 string, value2 string) string {
	return fmt.Sprintf("%s %s", value1, value2)
}

func (deck *DominoesDeck) DrawTile(tile string) bool {
	if deck.tiles[tile] {
		delete(deck.tiles, tile)
		return true
	}
	return false
}

func (deck *DominoesDeck) DrawRandomTile() string {
	if deck.NumberOfTilesLeft() == 0 {
		return ""
	}

	tilesLeft := make([]string, 0)
	for tileString, stillThere := range deck.tiles {
		if stillThere {
			tilesLeft = append(tilesLeft, tileString)
		}
	}
	sort.Sort(ByRandom(tilesLeft))
	tile := tilesLeft[0]
	deck.DrawTile(tile)
	return tile
}

func (deck *DominoesDeck) DrawRandomTiles(quantity int) []string {

	tilesLeft := make([]string, 0)
	for tileString, stillThere := range deck.tiles {
		if stillThere {
			tilesLeft = append(tilesLeft, tileString)
		}
	}
	sort.Sort(ByRandom(tilesLeft))
	tiles := make([]string, 0)
	for index, tile := range tilesLeft {
		if index < quantity {
			tiles = append(tiles, tile)
		}
	}

	deck.DrawSpecificTiles(tiles)
	return tiles
}

func (deck *DominoesDeck) DrawSpecificTiles(tiles []string) {
	for _, removedTile := range tiles {
		deck.DrawTile(removedTile)
	}
}

func (deck *DominoesDeck) PutTilesBack(tiles []string) {
	for _, removedTile := range tiles {
		deck.tiles[removedTile] = true
	}
}

func (deck *DominoesDeck) NumberOfTilesLeft() int {
	return len(deck.tiles)
}

func (deck *DominoesDeck) Contain(tileString string) bool {
	return deck.tiles[tileString]
}

func (deck *DominoesDeck) ContainTiles(tiles []string) bool {
	for _, tile := range tiles {
		if !deck.Contain(tile) {
			return false
		}
	}
	return true
}

func (deck *DominoesDeck) SerializedData() (data map[string]interface{}) {
	data = make(map[string]interface{})
	for tileString, boolValue := range deck.tiles {
		data[tileString] = boolValue
	}
	return data
}

func GetPoint(tileString string) int {
	tokens := strings.Split(tileString, " ")
	if len(tokens) == 2 {
		value1, _ := strconv.ParseInt(tokens[0], 10, 64)
		value2, _ := strconv.ParseInt(tokens[1], 10, 64)
		return int(value1 + value2)
	}
	return -1
}

func IsTwinTile(tileString string) bool {
	tokens := strings.Split(tileString, " ")
	if len(tokens) == 2 {
		value1, _ := strconv.ParseInt(tokens[0], 10, 64)
		value2, _ := strconv.ParseInt(tokens[1], 10, 64)

		return value1 == value2
	}
	return false
}

func ContainTiles(tileString string, tiles []string) bool {
	for _, tileInTiles := range tiles {
		if tileInTiles == tileString {
			return true
		}
	}
	return false
}
