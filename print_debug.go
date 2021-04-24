// +build debug

package folder

import (
	"fmt"
)

func debug(a ...interface{}) {
	fmt.Println(a...)
}
