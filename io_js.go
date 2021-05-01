// +build js

package folder

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	FieldNamesFileExtension    = "fns"
	DocumentsFileExtension     = "dcs"
	DocumentStatsFileExtension = "dst"
	TermStatsFileExtension     = "tst"
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

func (index *Index) loadShardCount() (err error) {
	var resp *http.Response

	url := fmt.Sprintf("%s/%s/shard_count", index.baseURL, index.Name)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = fmt.Fscanf(resp.Body, "%d", &index.ShardCount)
	if err != nil {
		return
	}

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

func (index *Index) loadFieldNamesFromReader(r io.Reader) (err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		index.FieldNames = append(index.FieldNames, scanner.Text())
	}
	err = scanner.Err()
	return
}

func (index *Index) LoadAllShards(progressCallback ProgressCallback, sleepDuration time.Duration) (err error) {
	for i := 0; i < index.ShardCount; i++ {
		err = index.loadShard(i)
		if err != nil {
			return
		}

		progressCallback(i+1, index.ShardCount)
		time.Sleep(sleepDuration)
	}
	return
}

func (index *Index) loadShard(shardID int) (err error) {
	err = index.loadDocumentsFromShard(shardID)
	if err != nil {
		return
	}

	err = index.loadDocumentStatsFromShard(shardID)
	if err != nil {
		return
	}

	err = index.loadTermStatsFromShard(shardID)
	if err != nil {
		return
	}

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

func (index *Index) loadDocumentsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	record, err = csvr.Read()
	if err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}

	headers := record

	for {
		record, err = csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}

		id := record[0]
		index.Documents[id] = documentFromRecord(headers[1:], record[1:])
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

func (index *Index) loadDocumentStatsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	_, err = csvr.Read()
	if err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}

	for {
		record, err = csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}

		id := record[0]
		tfs := strings.Split(record[1], " ")
		for _, v := range tfs {
			vv := strings.Split(v, ":")
			term := vv[0]
			frequency := vv[1]

			if _, ok := index.DocumentStats[id]; !ok {
				index.DocumentStats[id] = DocumentStat{TermFrequency: make(map[string]int)}
			}

			index.DocumentStats[id].TermFrequency[term], err = strconv.Atoi(frequency)
			if err != nil {
				return
			}
		}
	}
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

func (index *Index) loadTermStatsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	_, err = csvr.Read()
	if err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}

	for {
		record, err = csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}

		term := record[0]
		termStat := index.TermStats[term]
		ids := strings.Split(record[1], " ")
		termStat.DocumentIDs = append(index.TermStats[term].DocumentIDs, ids...)
		index.TermStats[term] = termStat
	}
	return
}
