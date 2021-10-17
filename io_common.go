package folder

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (index *Index) loadShardCountFromReader(r io.Reader) (err error) {
	_, err = fmt.Fscanf(r, "%d", &index.ShardCount)
	if err != nil {
		return
	}

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

func (index *Index) fetchDocumentFromReader(r io.Reader, documentID string) (document map[string]interface{}, err error) {
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
		if id == documentID {
			document = documentFromRecord(headers[1:], record[1:])
			return
		}
	}
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

func (index *Index) loadTermStatsFromReader(r io.Reader) (err error) {
	var record []string

	csvr := csv.NewReader(r)

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
		tfs := strings.Split(record[1], " ")
		if termStat.TermFrequencies == nil {
			termStat.TermFrequencies = make(map[string]int)
		}
		for _, v := range tfs {
			vv := strings.Split(v, ":")
			id := vv[0]
			frequency := vv[1]

			termStat.TermFrequencies[id], err = strconv.Atoi(frequency)
			if err != nil {
				return
			}
		}
		index.TermStats[term] = termStat
	}
	return
}
