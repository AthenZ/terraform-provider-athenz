package athenz

type stringSet map[string]struct{}

func (set stringSet) add(s string) {
	set[s] = struct{}{}
}

func (set stringSet) contains(s string) bool {
	_, ok := set[s]
	return ok
}
