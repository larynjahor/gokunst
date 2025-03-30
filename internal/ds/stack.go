package ds

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		buf: make([]T, 0, 256),
	}
}

type Stack[T any] struct {
	buf []T
}

func (s *Stack[T]) Push(val T) {
	s.buf = append(s.buf, val)
}

func (s *Stack[T]) Top() T {
	return s.buf[len(s.buf)-1]
}

func (s *Stack[T]) Pop() T {
	top := s.Top()

	s.buf = s.buf[:len(s.buf)-1]

	return top
}

func (s *Stack[T]) Empty() bool {
	return len(s.buf) == 0
}
