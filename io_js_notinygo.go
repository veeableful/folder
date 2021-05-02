// +build js,!tinygo

package folder

import (
	"fmt"
	"net/http"
)

func (index *Index) loadShardCount() (err error) {
	var resp *http.Response

	url := fmt.Sprintf("%s/%s/shard_count", index.baseURL, index.Name)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = index.loadShardCountFromReader(resp.Body)
	return
}

func (index *Index) loadFieldNamesDeferred() (err error) {
	var resp *http.Response

	dirPath := fmt.Sprintf("%s/%s", index.baseURL, index.Name)
	url := fmt.Sprintf("%s/%s", dirPath, FieldNamesFileExtension)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = index.loadFieldNamesFromReader(resp.Body)
	return
}

func (index *Index) loadDocumentsFromShard(shardID int) (err error) {
	var resp *http.Response

	if _, ok := index.LoadedDocumentsShards[shardID]; ok {
		return
	}

	debug("Loading documents shard", shardID)

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, DocumentsFileExtension)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = index.loadDocumentsFromReader(resp.Body)
	if err != nil {
		return
	}

	index.LoadedDocumentsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadDocumentStatsFromShard(shardID int) (err error) {
	var resp *http.Response

	if _, ok := index.LoadedDocumentStatsShards[shardID]; ok {
		return
	}

	debug("Loading document stats shard", shardID)

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, DocumentStatsFileExtension)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = index.loadDocumentStatsFromReader(resp.Body)
	if err != nil {
		return
	}

	index.LoadedDocumentStatsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadTermStatsFromShard(shardID int) (err error) {
	var resp *http.Response

	if _, ok := index.LoadedTermStatsShards[shardID]; ok {
		return
	}

	debug("Loading term stats shard", shardID)

	url := fmt.Sprintf("%s/%s/%d/%s", index.baseURL, index.Name, shardID, TermStatsFileExtension)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = index.loadTermStatsFromReader(resp.Body)
	if err != nil {
		return
	}

	index.LoadedTermStatsShards[shardID] = struct{}{}

	return
}
