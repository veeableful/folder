// +build tinygo

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"syscall"

	"github.com/veeableful/folder"
)

func doIndex(args []string) (err error) {
	var index folder.Index

	if len(args) < 1 {
		err = errors.New("please specify the document file path")
		return
	}

	filePath := args[0]

	index, err = folder.Load("index")
	if err != nil && err != os.ErrNotExist {
		return
	}

	err = index.IndexFilePath(filePath, "text")
	if err != nil {
		return err
	}

	err = index.Save("index")
	if err != nil {
		return
	}

	return
}

func doSearch(args []string) error {
	index, err := folder.Load("index")
	if err != nil {
		return err
	}

	s := args[0]
	result := index.Search(s)
	fmt.Printf("%+v\n", result)
	return nil
}

func main() {

}
