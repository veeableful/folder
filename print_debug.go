// +build debug

package folder

import (
	"fmt"
	"os"
)

func debug(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}
