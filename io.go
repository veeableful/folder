package folder

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	FieldNamesFileExtension    = "fns"
	DocumentsFileExtension     = "dcs"
	DocumentStatsFileExtension = "dst"
	TermStatsFileExtension     = "tst"
)

// Load loads an index from files
func Load(indexName string) (index Index, err error) {
	index.Name = indexName
	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}

	err = index.loadFieldNames()
	if err != nil {
		return
	}

	err = index.loadDocuments()
	if err != nil {
		return
	}

	err = index.loadDocumentStats()
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
func LoadDeferred(indexName string) (index Index, err error) {
	index.Name = indexName
	index.LoadedShards = map[int]struct{}{}

	err = index.loadShardCount()
	if err != nil {
		return
	}

	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}

	err = index.loadFieldNamesDeferred()
	if err != nil {
		return
	}

	return
}

func (index *Index) loadShardCount() (err error) {
	dirPath := fmt.Sprintf(".%s", index.Name)

	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, e error) (err error) {
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

func (index *Index) loadShardCountFS(f fs.FS) (err error) {
	dirPath := fmt.Sprintf(".%s", index.Name)

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

func (index *Index) calculateShardID(s string) (shardID int) {
	for _, r := range s {
		shardID = (shardID + int(r)) % index.ShardCount
	}
	return
}

// LoadFS loads an index from files using FS
func LoadFS(f fs.FS, indexName string) (index Index, err error) {
	index.Name = indexName
	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}
	index.f = f

	err = index.loadFieldNamesFS(f)
	if err != nil {
		return
	}

	err = index.loadDocumentsFS(f)
	if err != nil {
		return
	}

	err = index.loadDocumentStatsFS(f)
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
func LoadDeferredFS(f fs.FS, indexName string) (index Index, err error) {
	index.Name = indexName
	index.LoadedShards = map[int]struct{}{}
	index.f = f

	err = index.loadShardCountFS(f)
	if err != nil {
		return
	}

	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}

	err = index.loadFieldNamesFS(f)
	if err != nil {
		return
	}

	return
}

func (index *Index) loadFieldNames() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", index.Name, FieldNamesFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) loadFieldNamesDeferred() (err error) {
	var file *os.File

	dirPath := fmt.Sprintf(".%s", index.Name)
	err = os.MkdirAll(dirPath, 0700)
	if err != nil {
		return
	}

	filePath := fmt.Sprintf("%s/%s", dirPath, FieldNamesFileExtension)
	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) loadFieldNamesFS(f fs.FS) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", index.Name, FieldNamesFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
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

func (index *Index) loadShard(shardID int) (err error) {
	if _, ok := index.LoadedShards[shardID]; ok {
		return
	}

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

	index.LoadedShards[shardID] = struct{}{}
	return
}

func (index *Index) loadDocuments() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", index.Name, DocumentsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentsFromReader(file)
	return
}

func (index *Index) loadDocumentsFromShard(shardID int) (err error) {
	var file fs.File

	filePath := fmt.Sprintf(".%s/%d/%s", index.Name, shardID, DocumentsFileExtension)
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

	return
}

func (index *Index) loadDocumentsFS(f fs.FS) (err error) {
	var file fs.File

	filePath := fmt.Sprintf(".%s.%s", index.Name, DocumentsFileExtension)
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

func (index *Index) loadDocumentsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	record, err = csvr.Read()
	if err != nil {
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

func (index *Index) loadDocumentStats() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", index.Name, DocumentStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentStatsFromReader(file)
	return
}

func (index *Index) loadDocumentStatsFromShard(shardID int) (err error) {
	var file fs.File

	filePath := fmt.Sprintf(".%s/%d/%s", index.Name, shardID, DocumentStatsFileExtension)
	if index.f == nil {
		file, err = os.Open(filePath)
	} else {
		file, err = index.f.Open(filePath)
	}
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentStatsFromReader(file)
	if err != nil {
		return
	}

	return
}

func (index *Index) loadDocumentStatsFS(f fs.FS) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", index.Name, DocumentStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentStatsFromReader(file)
	return
}

func (index *Index) loadDocumentStatsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	_, err = csvr.Read()
	if err != nil {
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

func (index *Index) loadTermStats() (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", index.Name, TermStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	return
}

func (index *Index) loadTermStatsFromShard(shardID int) (err error) {
	var file fs.File

	filePath := fmt.Sprintf(".%s/%d/%s", index.Name, shardID, TermStatsFileExtension)
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

	return
}

func (index *Index) loadTermStatsFS(f fs.FS) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", index.Name, TermStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	return
}

func (index *Index) loadTermStatsFromReader(r io.Reader) (err error) {
	csvr := csv.NewReader(r)

	var record []string
	_, err = csvr.Read()
	if err != nil {
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

	err = index.saveDocumentStats()
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

	err = index.saveFieldNamesToShards()
	if err != nil {
		return
	}

	err = index.saveDocumentsToShards()
	if err != nil {
		return
	}

	err = index.saveDocumentStatsToShards()
	if err != nil {
		return
	}

	err = index.saveTermStatsToShards()
	if err != nil {
		return
	}

	return
}

func (index *Index) saveFieldNames() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", index.Name, FieldNamesFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

	dirPath := fmt.Sprintf(".%s", index.Name)
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

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", index.Name, DocumentsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
	ws := make(map[int]*csv.Writer)
	headers := []string{"id"}
	headers = append(headers, index.FieldNames...)

	for shardID := 0; shardID < index.ShardCount; shardID++ {
		var file *os.File

		dirPath := fmt.Sprintf(".%s/%d/", index.Name, shardID)
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
		ws[shardID] = w
	}

	for documentID, document := range index.Documents {
		record := recordFromDocument(documentID, headers, document)
		shardID := index.calculateShardID(documentID)
		ws[shardID].Write(record)
	}

	for _, w := range ws {
		w.Flush()
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

func (index *Index) saveDocumentStats() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", index.Name, DocumentStatsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)

	for id, documentStat := range index.DocumentStats {
		record := []string{id}
		pairs := []string{}

		for term, frequency := range documentStat.TermFrequency {
			frequencyStr := strconv.Itoa(frequency)
			pairs = append(pairs, strings.Join([]string{term, frequencyStr}, ":"))
		}

		record = append(record, strings.Join(pairs, " "))
		w.Write(record)
	}
	w.Flush()

	return
}

func (index *Index) saveDocumentStatsToShards() (err error) {
	ws := make(map[int]*csv.Writer)

	for shardID := 0; shardID < index.ShardCount; shardID++ {
		var file *os.File

		dirPath := fmt.Sprintf(".%s/%d/", index.Name, shardID)
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return
		}

		filePath := fmt.Sprintf("%s/%s", dirPath, DocumentStatsFileExtension)
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()

		w := csv.NewWriter(file)
		ws[shardID] = w
	}

	for documentID, documentStat := range index.DocumentStats {
		record := []string{documentID}
		pairs := []string{}

		for term, frequency := range documentStat.TermFrequency {
			frequencyStr := strconv.Itoa(frequency)
			pairs = append(pairs, strings.Join([]string{term, frequencyStr}, ":"))
		}

		record = append(record, strings.Join(pairs, " "))
		shardID := index.calculateShardID(documentID)
		ws[shardID].Write(record)
	}

	for _, w := range ws {
		w.Flush()
	}
	return
}

func (index *Index) saveTermStats() (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", index.Name, TermStatsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)

	for term, stat := range index.TermStats {
		record := []string{term}
		record = append(record, strings.Join(stat.DocumentIDs, " "))
		w.Write(record)
	}
	w.Flush()

	return
}

func (index *Index) saveTermStatsToShards() (err error) {
	ws := make(map[int]*csv.Writer)

	for shardID := 0; shardID < index.ShardCount; shardID++ {
		var file *os.File

		dirPath := fmt.Sprintf(".%s/%d/", index.Name, shardID)
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return
		}

		filePath := fmt.Sprintf("%s/%s", dirPath, TermStatsFileExtension)
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer file.Close()

		w := csv.NewWriter(file)
		ws[shardID] = w
	}

	for term, stat := range index.TermStats {
		record := []string{term}
		record = append(record, strings.Join(stat.DocumentIDs, " "))
		shardID := index.calculateShardID(term)
		ws[shardID].Write(record)
	}

	for _, w := range ws {
		w.Flush()
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
