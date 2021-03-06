//go:build !js
// +build !js

package folder

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	FieldNamesFileExtension = "fns"
	DocumentsFileExtension  = "dcs"
	TermStatsFileExtension  = "tst"
	ShardCountFileName      = "shard_count"
)

type ProgressCallback func(loadedShardsCount, totalShardsCount int)

// Load loads an index from files
func Load(indexName string) (index *Index, err error) {
	index = New()
	index.Name = indexName

	err = index.loadFieldNames()
	if err != nil {
		return
	}

	err = index.loadDocuments()
	if err != nil {
		return
	}

	err = index.loadTermStats()
	if err != nil {
		return
	}

	return
}

// LoadDeferred loads an index metadata only the rest of the data is loaded when needed.
func LoadDeferred(indexName string) (index *Index, err error) {
	index = New()
	index.Name = indexName

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
	var file *os.File

	url := fmt.Sprintf("%s/shard_count", index.Name)
	file, err = os.Open(url)
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadShardCountFromReader(file)
	return
}

func (index *Index) loadShardCountFS(f fs.FS) (err error) {
	dirPath := index.Name

	err = fs.WalkDir(f, dirPath, func(path string, d fs.DirEntry, e error) (err error) {
		if e != nil {
			err = e
			return
		}

		if !d.IsDir() {
			return
		}

		shardID := -1
		fmt.Sscanf(filepath.Base(path), "%d", &shardID)
		if shardID < 0 {
			return
		}

		index.ShardCount += 1
		return
	})
	if err != nil {
		return
	}
	return
}

// LoadFS loads an index from files using FS
func LoadFS(f fs.FS, indexName string) (index *Index, err error) {
	index = New()
	index.Name = indexName
	index.f = f

	err = index.loadFieldNamesFS(f)
	if err != nil {
		return
	}

	err = index.loadDocumentsFS(f)
	if err != nil {
		return
	}

	err = index.loadTermStatsFS(f)
	if err != nil {
		return
	}

	return
}

// LoadDeferredFS loads an index metadata only the rest of the data is loaded when needed.
func LoadDeferredFS(f fs.FS, indexName string) (index *Index, err error) {
	index = New()
	index.Name = indexName
	index.f = f

	err = index.loadShardCountFS(f)
	if err != nil {
		return
	}

	err = index.loadFieldNamesDeferred()
	if err != nil {
		return
	}

	return
}

// LoadWithProgressFS loads the entire index however user can monitor progress by passing a progress callback and also specify sleep duration between each shard.
func LoadWithProgressFS(f fs.FS, indexName string, progressCallback ProgressCallback, sleepDuration time.Duration) (index *Index, err error) {
	index = New()
	index.Name = indexName
	index.f = f

	err = index.loadShardCountFS(f)
	if err != nil {
		return
	}

	err = index.loadFieldNamesDeferred()
	if err != nil {
		return
	}

	err = index.LoadAllShards(progressCallback, sleepDuration)
	if err != nil {
		return
	}

	return
}

func (index *Index) loadFieldNames() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf("%s.%s", index.Name, FieldNamesFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) loadFieldNamesDeferred() (err error) {
	var file fs.File

	dirPath := index.Name
	filePath := fmt.Sprintf("%s/%s", dirPath, FieldNamesFileExtension)
	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) loadFieldNamesFS(f fs.FS) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf("%s.%s", index.Name, FieldNamesFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) LoadAllShards(progressCallback ProgressCallback, sleepDuration time.Duration) (err error) {
	for i := 0; i < index.ShardCount; i++ {
		err = index.loadDocumentsFromShard(uint32(i))
		if err != nil {
			return
		}

		err = index.loadTermStatsFromShard(uint32(i))
		if err != nil {
			return
		}

		progressCallback(i+1, index.ShardCount)
		time.Sleep(sleepDuration)
	}
	return
}

func (index *Index) loadDocuments() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf("%s.%s", index.Name, DocumentsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentsFromReader(file)
	return
}

func (index *Index) fetchDocumentFromShard(shardID uint32, documentID string) (document map[string]interface{}, err error) {
	var file fs.File

	if _, ok := index.LoadedDocumentsShards[shardID]; ok {
		return
	}

	filePath := fmt.Sprintf("%s/%d/%s", index.Name, shardID, DocumentsFileExtension)
	debug("  Fetching document from shard: ", filePath)

	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	document, err = index.fetchDocumentFromReader(file, documentID)
	if err != nil {
		return
	}

	return
}

func (index *Index) loadDocumentsFromShard(shardID uint32) (err error) {
	var file fs.File

	if _, ok := index.LoadedDocumentsShards[shardID]; ok {
		return
	}

	filePath := fmt.Sprintf("%s/%d/%s", index.Name, shardID, DocumentsFileExtension)
	debug("  Loading documents shard: ", filePath)

	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentsFromReader(file)
	if err != nil {
		return
	}

	index.LoadedDocumentsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadDocumentsFS(f fs.FS) (err error) {
	var file fs.File

	filePath := fmt.Sprintf("%s.%s", index.Name, DocumentsFileExtension)
	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentsFromReader(file)
	return
}

func documentFromRecord(headers, record []string) (document map[string]interface{}) {
	document = make(map[string]interface{})

	for i, header := range headers {
		setField(document, header, record[i])
	}

	return
}

func (index *Index) loadTermStats() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf("%s.%s", index.Name, TermStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	return
}

func (index *Index) loadTermStatsFromShard(shardID uint32) (err error) {
	var file fs.File

	if _, ok := index.LoadedTermStatsShards[shardID]; ok {
		return
	}

	filePath := fmt.Sprintf("%s/%d/%s", index.Name, shardID, TermStatsFileExtension)
	debug("  Loading term stats shard: ", filePath)

	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	if err != nil {
		return
	}

	index.LoadedTermStatsShards[shardID] = struct{}{}

	return
}

func (index *Index) loadTermStatsFS(f fs.FS) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf("%s.%s", index.Name, TermStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	return
}

func (index *Index) Save(indexName string) (err error) {
	index.Name = indexName

	err = index.saveFieldNames()
	if err != nil {
		return
	}

	err = index.saveDocuments()
	if err != nil {
		return
	}

	err = index.saveTermStats()
	if err != nil {
		return
	}

	return
}

func (index *Index) SaveToShards(indexName string, shardCount int) (err error) {
	index.Name = indexName
	index.ShardCount = shardCount

	err = index.saveShardCount()
	if err != nil {
		return
	}

	err = index.saveFieldNamesToShards()
	if err != nil {
		return
	}

	err = index.saveDocumentsToShards()
	if err != nil {
		return
	}

	err = index.saveTermStatsToShards()
	if err != nil {
		return
	}

	return
}

func (index *Index) saveShardCount() (err error) {
	var file *os.File

	dirPath := index.Name
	err = os.MkdirAll(dirPath, 0700)
	if err != nil {
		return
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, ShardCountFileName)
	file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprint(index.ShardCount))
	if err != nil {
		return
	}

	return
}

func (index *Index) saveFieldNames() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf("%s.%s", index.Name, FieldNamesFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	for _, field := range index.FieldNames {
		_, err = io.WriteString(file, fmt.Sprintf("%s\n", field))
		if err != nil {
			return
		}
	}

	return
}

func (index *Index) saveFieldNamesToShards() (err error) {
	var file *os.File

	dirPath := index.Name
	err = os.MkdirAll(dirPath, 0700)
	if err != nil {
		return
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, FieldNamesFileExtension)
	file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	for _, field := range index.FieldNames {
		_, err = io.WriteString(file, fmt.Sprintf("%s\n", field))
		if err != nil {
			return
		}
	}

	return
}

func (index *Index) saveDocuments() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf("%s.%s", index.Name, DocumentsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)

	headers := []string{"id"}
	headers = append(headers, index.FieldNames...)
	w.Write(headers)

	for id, document := range index.Documents {
		record := recordFromDocument(id, headers, document)
		w.Write(record)
	}
	w.Flush()

	return
}

func (index *Index) saveDocumentsToShards() (err error) {
	shardDocumentIDsMap := make(map[int][]string)
	headers := []string{"id"}
	headers = append(headers, index.FieldNames...)

	for documentID := range index.Documents {
		shardID := index.CalculateShardID(documentID)
		shardDocumentIDsMap[int(shardID)] = append(shardDocumentIDsMap[int(shardID)], documentID)
	}

	for shardID, documentIDs := range shardDocumentIDsMap {
		var file *os.File

		dirPath := fmt.Sprintf("%s/%d/", index.Name, shardID)
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return
		}

		filePath := fmt.Sprintf("%s/%s", dirPath, DocumentsFileExtension)
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()

		w := csv.NewWriter(file)
		w.Write(headers)

		for _, documentID := range documentIDs {
			document := index.Documents[documentID]
			record := recordFromDocument(documentID, headers, document)
			w.Write(record)
		}

		w.Flush()
		file.Close()
	}

	return
}

func recordFromDocument(id string, headers []string, document map[string]interface{}) (record []string) {
	for _, header := range headers {
		if header == "id" {
			record = append(record, id)
			continue
		}

		values := fieldValuesFromRoot(document, header)
		value := strings.Join(values, ",")
		record = append(record, value)
	}
	return
}

func (index *Index) saveTermStats() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf("%s.%s", index.Name, TermStatsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)

	for term, stat := range index.TermStats {
		record := []string{term}
		pairs := []string{}

		for documentID, frequency := range stat.TermFrequencies {
			frequencyStr := strconv.Itoa(frequency)
			pairs = append(pairs, strings.Join([]string{documentID, frequencyStr}, ":"))
		}

		record = append(record, strings.Join(pairs, " "))
		w.Write(record)
	}
	w.Flush()

	return
}

func (index *Index) saveTermStatsToShards() (err error) {
	shardTermsMap := make(map[int][]string)

	for term := range index.TermStats {
		shardID := index.CalculateShardID(term)
		shardTermsMap[int(shardID)] = append(shardTermsMap[int(shardID)], term)
	}

	for shardID, terms := range shardTermsMap {
		var file *os.File

		dirPath := fmt.Sprintf("%s/%d/", index.Name, shardID)
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return
		}

		filePath := fmt.Sprintf("%s/%s", dirPath, TermStatsFileExtension)
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		w := csv.NewWriter(file)

		for _, term := range terms {
			stat := index.TermStats[term]
			record := []string{term}
			pairs := []string{}

			for documentID, frequency := range stat.TermFrequencies {
				frequencyStr := strconv.Itoa(frequency)
				pairs = append(pairs, strings.Join([]string{documentID, frequencyStr}, ":"))
			}

			record = append(record, strings.Join(pairs, " "))
			w.Write(record)
		}

		w.Flush()
		file.Close()
	}

	return
}

// IndexFilePath indexes a file or directory containing files and assumes a certain data type such
// as text, JSON, or JSONL.
func (index *Index) IndexFilePath(filePath, dataType string) (err error) {
	var file *os.File

	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	err = index.IndexReader(file, dataType)
	return
}

// IndexFilePathWithIDField performs similar operation as IndexFilePath but user can provide which
// document field value to be used as the document ID. This is only compatible with "json" and "jsonl"
// data types.
func (index *Index) IndexFilePathWithIDField(filePath, dataType, idField string) (err error) {
	var file *os.File

	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	err = index.IndexReaderWithIDField(file, dataType, idField)
	return
}

// IndexReader indexes a reader and assumes a certain data type such as text, JSON, or JSONL.
func (index *Index) IndexReader(r io.Reader, dataType string) (err error) {
	var data []byte

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	err = index.IndexData(data, dataType)
	return
}

// IndexReaderWithIDField performs similar operation as IndexReader but user can provide which
// document field value to be used as the document ID. This is only compatible with "json" and "jsonl"
// data types.
func (index *Index) IndexReaderWithIDField(r io.Reader, dataType, idField string) (err error) {
	var data []byte

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	err = index.IndexDataWithIDField(data, dataType, idField)
	return
}
