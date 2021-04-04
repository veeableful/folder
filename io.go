package folder

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
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
	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}

	err = index.loadFieldNames(indexName)
	if err != nil {
		return
	}

	err = index.loadDocuments(indexName)
	if err != nil {
		return
	}

	err = index.loadDocumentStats(indexName)
	if err != nil {
		return
	}

	err = index.loadTermStats(indexName)
	if err != nil {
		return
	}

	return
}

// LoadFS loads an index from files using FS
func LoadFS(f fs.FS, indexName string) (index Index, err error) {
	index.FieldNames = make([]string, 0)
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = map[string]DocumentStat{}
	index.TermStats = map[string]TermStat{}

	err = index.loadFieldNamesFS(f, indexName)
	if err != nil {
		return
	}

	err = index.loadDocumentsFS(f, indexName)
	if err != nil {
		return
	}

	err = index.loadDocumentStatsFS(f, indexName)
	if err != nil {
		return
	}

	err = index.loadTermStatsFS(f, indexName)
	if err != nil {
		return
	}

	return
}

func (index *Index) loadFieldNames(indexName string) (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", indexName, FieldNamesFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadFieldNamesFromReader(file)
	return
}

func (index *Index) loadFieldNamesFS(f fs.FS, indexName string) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", indexName, FieldNamesFileExtension))
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

func (index *Index) loadDocuments(indexName string) (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", indexName, DocumentsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentsFromReader(file)
	return
}

func (index *Index) loadDocumentsFS(f fs.FS, indexName string) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", indexName, DocumentsFileExtension))
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

func (index *Index) loadDocumentStats(indexName string) (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", indexName, DocumentStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadDocumentStatsFromReader(file)
	return
}

func (index *Index) loadDocumentStatsFS(f fs.FS, indexName string) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", indexName, DocumentStatsFileExtension))
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

func (index *Index) loadTermStats(indexName string) (err error) {
	var file *os.File

	file, err = os.Open(fmt.Sprintf(".%s.%s", indexName, TermStatsFileExtension))
	if err != nil {
		return
	}
	defer file.Close()

	err = index.loadTermStatsFromReader(file)
	return
}

func (index *Index) loadTermStatsFS(f fs.FS, indexName string) (err error) {
	var file fs.File

	file, err = f.Open(fmt.Sprintf(".%s.%s", indexName, TermStatsFileExtension))
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
	err = index.saveFieldNames(indexName)
	if err != nil {
		return
	}

	err = index.saveDocuments(indexName)
	if err != nil {
		return
	}

	err = index.saveDocumentStats(indexName)
	if err != nil {
		return
	}

	err = index.saveTermStats(indexName)
	if err != nil {
		return
	}

	return
}

func (index *Index) saveFieldNames(indexName string) (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", indexName, FieldNamesFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

func (index *Index) saveDocuments(indexName string) (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", indexName, DocumentsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

func (index *Index) saveDocumentStats(indexName string) (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", indexName, DocumentStatsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

func (index *Index) saveTermStats(indexName string) (err error) {
	var file *os.File

	file, err = os.OpenFile(fmt.Sprintf(".%s.%s", indexName, TermStatsFileExtension), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
