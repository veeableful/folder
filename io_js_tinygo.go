// +build js,tinygo

package folder

import (
	"fmt"
	"io"
	"strings"
	"syscall/js"
)

func textDataFromURL(url string) (text string, err error) {
	c := make(chan string, 1)
	jsURL := js.ValueOf(url)
	jsTextCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c <- args[0].String()
		return nil
	})
	jsFetchCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsResponse := args[0]
		jsTextPromise := jsResponse.Call("text")
		jsTextPromise.Call("then", jsTextCallback)
		return nil
	})
	promise := js.Global().Call("fetch", jsURL)
	promise.Call("then", jsFetchCallback)
	data := <-c
	text = string(data)
	return
}

func textReaderFromURL(url string) (r io.Reader, err error) {
	var text string

	text, err = textDataFromURL(url)
	if err != nil {
		return
	}

	r = strings.NewReader(text)
	return
}

func (index *Index) loadShardCount() (err error) {
	var r io.Reader

	url := fmt.Sprintf("%s/%s/shard_count", index.baseURL, index.Name)
	r, err = textReaderFromURL(url)
	if err != nil {
		return
	}

	err = index.loadShardCountFromReader(r)
	return
}

func (index *Index) loadFieldNamesDeferred() (err error) {
	var r io.Reader

	dirPath := fmt.Sprintf("%s/%s", index.baseURL, index.Name)
	url := fmt.Sprintf("%s/%s", dirPath, FieldNamesFileExtension)
	r, err = textReaderFromURL(url)
	if err != nil {
		return
	}

	err = index.loadFieldNamesFromReader(r)
	return
}

func (index *Index) loadDocumentsFromShard(shardID int) (err error) {
	var r io.Reader

	if _, ok := index.LoadedDocumentsShards[shardID]; ok {
		return
	}

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, DocumentsFileExtension)
	debug("  Loading documents shard:", url)

	r, err = textReaderFromURL(url)
	if err != nil {
		return
	}

	err = index.loadDocumentsFromReader(r)
	if err != nil {
		return
	}

	index.LoadedDocumentsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadDocumentStatsFromShard(shardID int) (err error) {
	var r io.Reader

	if _, ok := index.LoadedDocumentStatsShards[shardID]; ok {
		return
	}

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, DocumentStatsFileExtension)
	debug("  Loading document stats shard:", url)

	r, err = textReaderFromURL(url)
	if err != nil {
		return
	}

	err = index.loadDocumentStatsFromReader(r)
	if err != nil {
		return
	}

	index.LoadedDocumentStatsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadTermStatsFromShard(shardID int) (err error) {
	var r io.Reader

	if _, ok := index.LoadedTermStatsShards[shardID]; ok {
		return
	}

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, TermStatsFileExtension)
	debug("  Loading term stats shard:", url)

	r, err = textReaderFromURL(url)
	if err != nil {
		return
	}

	err = index.loadTermStatsFromReader(r)
	if err != nil {
		return
	}

	index.LoadedTermStatsShards[shardID] = struct{}{}

	return
}
