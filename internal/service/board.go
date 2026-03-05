package service

type boardResetter interface {
	Reset() error
}

type BoardService struct {
	board boardResetter
}

func NewBoardService(board boardResetter) *BoardService {
	return &BoardService{board: board}
}

func (s *BoardService) Reset() error {
	return s.board.Reset()
}
