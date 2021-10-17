//go:build js
// +build js

package folder

import (
	"time"
)

const (
	FieldNamesFileExtension = "fns"
	DocumentsFileExtension  = "dcs"
	TermStatsFileExtension  = "tst"
)

type ProgressCallback func(loadedShardsCount, totalShardsCount int)

// LoadDeferred loads an index metadata only the rest of the data is loaded when needed.
func LoadDeferred(indexName, baseURL string) (index *Index, err error) {
	index = New()
	index.Name = indexName
	index.baseURL = baseURL

	err = index.loadShardCount()
	if err != nil {
		return
	}

	err = index.loadFieldNamesDeferred()
	if err != nil {
		return
	}

	return
}

func (index *Index) LoadAllShards(progressCallback ProgressCallback, sleepDuration time.Duration) (err error) {
	for i := 0; i < index.ShardCount; i++ {
		err = index.loadShard(uint32(i))
		if err != nil {
			return
		}

		progressCallback(i+1, index.ShardCount)
		time.Sleep(sleepDuration)
	}
	return
}

func (index *Index) loadShard(shardID uint32) (err error) {
	err = index.loadDocumentsFromShard(shardID)
	if err != nil {
		return
	}

	err = index.loadTermStatsFromShard(shardID)
	if err != nil {
		return
	}

	return
}

func documentFromRecord(headers, record []string) (document map[string]interface{}) {
	document = make(map[string]interface{})

	for i, header := range headers {
		setField(document, header, record[i])
	}

	return
}
