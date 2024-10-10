package cube

import (
	"errors"
)

var (
	ErrFaceLengthsDiffer    = errors.New("face lengths differ")
	ErrFaceNotPerfectSquare = errors.New("face length not perfect square")
	ErrSliceParam           = errors.New("slice parameter is too big for cube")
)

const (
	axisX axis = iota
	axisY
	axisZ
)

type axis int

type piece struct {
	x, y, z *tile
}

type tile struct {
	coordinate      int
	coordinateStart int
	color           rune
}

type turn struct {
	axis      axis
	layerMin  int
	layerMax  int
	rotations int
	clockwise bool
}

type Cube struct {
	max    int
	pieces []*piece
}

type CubeFaces struct {
	Up    []rune
	Left  []rune
	Front []rune
	Right []rune
	Back  []rune
	Down  []rune
}

func (cf *CubeFaces) All() []*[]rune {
	return []*[]rune{&cf.Up, &cf.Left, &cf.Front, &cf.Right, &cf.Back, &cf.Down}
}

func (cf *CubeFaces) Validate() error {
	if !isEqualLength(cf.All()) {
		return ErrFaceLengthsDiffer
	}

	if !isPerfectSquare(len(cf.Left)) {
		return ErrFaceNotPerfectSquare
	}

	return nil
}

func (c *Cube) Dimension() int {
	return c.max + 1
}

func (c *Cube) Faces() *CubeFaces {
	numSide := c.Dimension()
	numTiles := numSide * numSide

	up := make([]rune, numTiles)
	down := make([]rune, numTiles)
	left := make([]rune, numTiles)
	right := make([]rune, numTiles)
	front := make([]rune, numTiles)
	back := make([]rune, numTiles)

	for _, p := range c.pieces {
		// Up face (y == max)
		if p.y.coordinate == c.max {
			index := p.x.coordinate + p.z.coordinate*numSide
			up[index] = p.y.color
		}
		// Down face (y == 0)
		if p.y.coordinate == 0 {
			index := (c.max-p.z.coordinate)*numSide + p.x.coordinate
			down[index] = p.y.color
		}
		// Left face (x == 0)
		if p.x.coordinate == 0 {
			index := (c.max-p.y.coordinate)*numSide + p.z.coordinate
			left[index] = p.x.color
		}
		// Right face (x == max)
		if p.x.coordinate == c.max {
			index := (c.max-p.y.coordinate)*numSide + (c.max - p.z.coordinate)
			right[index] = p.x.color
		}

		// Back face (z == 0)
		if p.z.coordinate == 0 {
			index := (c.max - p.x.coordinate) + (c.max-p.y.coordinate)*numSide
			back[index] = p.z.color
		}
		// Front face (z == max)
		if p.z.coordinate == c.max {
			index := p.x.coordinate + (c.max-p.y.coordinate)*numSide
			front[index] = p.z.color
		}
	}

	return &CubeFaces{
		Up:    up,
		Down:  down,
		Left:  left,
		Right: right,
		Front: front,
		Back:  back,
	}
}

func NewCube(size int) *Cube {
	n := pow(size, 3) - pow(size-2, 3)

	pieces := make([]*piece, 0, n)

	max := size - 1
	isInner := func(pos ...int) bool {
		for _, p := range pos {
			if p == 0 || p == max {
				return false
			}
		}
		return true
	}

	for x := range size {
		for y := range size {
			for z := range size {
				if isInner(x, y, z) {
					continue
				}
				pieces = append(pieces, newPiece(max, x, y, z))
			}
		}
	}

	return &Cube{
		max:    max,
		pieces: pieces,
	}
}

func newPiece(max int, x, y, z int) *piece {
	colorize := func(coord int, c0, cMax rune) rune {
		if coord == 0 {
			return c0
		} else if coord == max {
			return cMax
		}
		return 0
	}

	return &piece{
		x: newTile(x, colorize(x, 'o', 'r')),
		y: newTile(y, colorize(y, 'y', 'w')),
		z: newTile(z, colorize(z, 'b', 'g')),
	}
}

func newTile(coord int, color rune) *tile {
	return &tile{coordinate: coord, coordinateStart: coord, color: color}
}

func (c *Cube) turnAxis(t1, t2 *tile, direction bool, rotations int) {
	for range rotations {
		if direction {
			t1.coordinate, t2.coordinate = c.max-t2.coordinate, t1.coordinate
		} else {
			t2.coordinate, t1.coordinate = c.max-t1.coordinate, t2.coordinate
		}
		t1.color, t2.color = t2.color, t1.color
	}
}

func (c *Cube) turn(turn turn) *Cube {
	inRange := func(n int) bool {
		return turn.layerMin <= n && n <= turn.layerMax
	}

	for _, p := range c.pieces {
		switch turn.axis {
		case axisX:
			if inRange(p.x.coordinate) {
				c.turnAxis(p.y, p.z, turn.clockwise, turn.rotations)
			}
		case axisY:
			if inRange(p.y.coordinate) {
				c.turnAxis(p.x, p.z, turn.clockwise, turn.rotations)
			}
		case axisZ:
			if inRange(p.z.coordinate) {
				c.turnAxis(p.x, p.y, turn.clockwise, turn.rotations)
			}
		}
	}
	return c
}

func (c *Cube) ExecuteMove(move Move) error {
	layerMin := c.max
	layerMax := c.max

	if move.Wide {
		layerMin -= move.Slices - 1
	}

	if layerMin < 0 {
		return ErrSliceParam
	}

	if move.isAny('L', 'B', 'D', 'E') {
		layerMin, layerMax = c.max-layerMax, c.max-layerMin
	} else if move.isAny('x', 'y', 'z') {
		layerMin = 0
		layerMax = c.max
	}

	clockwise := !move.Inverted
	if move.isAny('R', 'F', 'D', 'E', 'S', 'x', 'z') {
		clockwise = !clockwise
	}

	axis := axisX
	if move.isAny('F', 'B', 'S', 'z') {
		axis = axisZ
	} else if move.isAny('U', 'D', 'E', 'y') {
		axis = axisY
	}

	if move.isAny('M', 'E', 'S') {
		if c.max%2 == 1 {
			return nil
		}

		layerMin = c.max / 2
		layerMax = layerMin
	}

	c.turn(
		turn{
			axis:      axis,
			layerMin:  layerMin,
			layerMax:  layerMax,
			rotations: move.Rotations,
			clockwise: clockwise,
		})
	return nil
}

func (c *Cube) ExecuteMoves(moves ...Move) *Cube {
	for _, m := range moves {
		c.ExecuteMove(m)
	}
	return c
}
