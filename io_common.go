package folder

import (
	"fmt"
	"io"
)

func (index *Index) loadShardCountFromReader(r io.Reader) (err error) {
	_, err = fmt.Fscanf(r, "%d", &index.ShardCount)
	if err != nil {
		return
	}

	return
}
