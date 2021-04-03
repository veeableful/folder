package folder

type StringSet struct {
	m map[string]struct{}
}

func MakeStringSet(a []string) StringSet {
	m := make(map[string]struct{})
	for _, v := range a {
		m[v] = struct{}{}
	}
	return StringSet{m}
}

func (s *StringSet) Add(v string) {
	s.m[v] = struct{}{}
}

func (s *StringSet) Remove(v string) {
	delete(s.m, v)
}

func (s *StringSet) Contains(v string) (ok bool) {
	_, ok = s.m[v]
	return
}

func (s *StringSet) List() (a []string) {
	for v := range s.m {
		a = append(a, v)
	}
	return
}

func (s *StringSet) Union(other StringSet) {
	for v := range other.m {
		s.Add(v)
	}
}

func (s *StringSet) Intersects(other StringSet) {
	m := make(map[string]struct{})

	for v := range other.m {
		if s.Contains(v) {
			m[v] = struct{}{}
		}
	}

	s.m = m
}

func (s *StringSet) Len() int {
	return len(s.m)
}
