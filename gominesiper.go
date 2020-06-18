package gominesweeper

type cell struct {
}
type State int
type Minesweeper struct {
	Height int
	Width  int
	cells  [][]cell
	sendCh <-chan <-chan State
}

func New(h, w int) *Minesweeper {
	m := Minesweeper{Height: h, Width: w}
	cells := make([][]cell, m.Height)
	for i := 0; i < h; i++ {
		cells[i] = make([]cell, m.Width)
	}
	m.cells = cells
	return &m
}
