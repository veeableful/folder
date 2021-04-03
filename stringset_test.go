package folder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	a = []string{"a", "b", "c", "d", "e"}
	b = []string{"b", "d", "f"}
	c = []string{"d", "e", "f"}
)

func TestUnion(t *testing.T) {
	aa := MakeStringSet(a)
	bb := MakeStringSet(b)
	cc := MakeStringSet(c)
	aa.Union(bb)
	aa.Union(cc)
	assert.Equal(t, aa.m, map[string]struct{}{
		"a": Empty{},
		"b": Empty{},
		"c": Empty{},
		"d": Empty{},
		"e": Empty{},
		"f": Empty{},
	})
}

func TestIntersects(t *testing.T) {
	aa := MakeStringSet(a)
	bb := MakeStringSet(b)
	cc := MakeStringSet(c)
	aa.Intersects(bb)
	aa.Intersects(cc)
	assert.Equal(t, aa.m, map[string]struct{}{"d": Empty{}})
}
