package conway

type GeneticData struct {
	Board      [][]bool
	Generation int
}

func (g *GeneticData) NextGeneration(boardSize int) {
	newBoard := make([][]bool, boardSize)
	for i := range newBoard {
		newBoard[i] = make([]bool, boardSize)
	}

	for i, gene := range g.Board {
		for j, alive := range gene {

			neighbors := g.GetNumberNeighbors(i, j, boardSize)

			if alive {
				if neighbors < 2 || neighbors > 3 {
					newBoard[i][j] = false
				} else {
					newBoard[i][j] = true
				}
			} else {
				if neighbors == 3 {
					newBoard[i][j] = true
				} else {
					newBoard[i][j] = false
				}
			}
		}
	}

	for i := range g.Board {
		copy(g.Board[i], newBoard[i])
	}
}

func (g *GeneticData) GetNumberNeighbors(x, y, boardSize int) int {
	count := 0

	neighbors := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, offset := range neighbors {
		nX, nY := x+offset[0], y+offset[1]
		if nX >= 0 && nX < boardSize && nY >= 0 && nY < boardSize && g.GetGene(nX, nY, boardSize) {
			count++
		}
	}

	return count
}

func (g *GeneticData) GetGene(x, y, boardSize int) bool {
	if x < 0 || x >= boardSize || y < 0 || y >= boardSize {
		return false
	}

	return g.Board[x][y]
}
