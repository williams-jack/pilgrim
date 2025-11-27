package shared

type Set[T comparable] struct {
	elements map[T]bool
}

func NewSet[T comparable](args ...T) *Set[T] {
	s := &Set[T]{elements: make(map[T]bool)}
	for _, arg := range args {
		s.Add(arg)
	}
	return s
}

func (s *Set[T]) Add(element T) {
	if s.elements == nil {
		s.elements = make(map[T]bool)
	}
	s.elements[element] = true
}

func (s *Set[T]) Remove(element T) {
	if s.elements != nil {
		delete(s.elements, element)
	}
}

func (s *Set[T]) Contains(element T) bool {
	if s.elements == nil {
		return false
	}
	_, exists := s.elements[element]
	return exists
}

func (s *Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s.elements))
	for element := range s.elements {
		slice = append(slice, element)
	}
	return slice
}

func (s *Set[T]) Size() int {
	return len(s.elements)
}

