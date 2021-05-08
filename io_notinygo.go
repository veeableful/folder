// +build !tinygo

package folder

import (
	"bufio"
	"bytes"
	"encoding/json"
)

// IndexData indexes an array of bytes and assumes a certain data type such as text, JSON, or JSONL.
func (index *Index) IndexData(data []byte, dataType string) (err error) {
	var m map[string]interface{}

	if dataType == "text" {
		m = make(map[string]interface{})
		m["text"] = string(data)
		index.Index(m)
	} else if dataType == "json" {
		m = make(map[string]interface{})
		err = json.Unmarshal(data, &m)
		if err != nil {
			return err
		}

		index.Index(m)
	} else if dataType == "jsonl" {
		r := bytes.NewReader(data)
		scanner := bufio.NewScanner(r)

		i := 0

		for scanner.Scan() {
			i += 1
			m = make(map[string]interface{})
			err = json.Unmarshal(scanner.Bytes(), &m)
			if err != nil {
				return err
			}
			index.Index(m)
		}

		err = scanner.Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// IndexDataWithIDField performs similar operation as IndexData but user can provide which document
// field value to be used as the document ID. This is only compatible with "json" and "jsonl" data
// types.
func (index *Index) IndexDataWithIDField(data []byte, dataType, idField string) (err error) {
	var m map[string]interface{}

	if dataType == "text" {
		m = make(map[string]interface{})
		m["text"] = string(data)
		index.Index(m)
	} else if dataType == "json" {
		m = make(map[string]interface{})
		err = json.Unmarshal(data, &m)
		if err != nil {
			return err
		}

		values := fieldValuesFromRoot(m, idField)
		if len(values) == 0 {
			err = ErrDocumentMissingIDField
			return
		}

		documentID := values[0]
		_, err = index.IndexWithID(m, documentID)
		if err != nil {
			return
		}
	} else if dataType == "jsonl" {
		r := bytes.NewReader(data)
		scanner := bufio.NewScanner(r)

		i := 0

		for scanner.Scan() {
			i += 1
			m = make(map[string]interface{})
			err = json.Unmarshal(scanner.Bytes(), &m)
			if err != nil {
				return err
			}

			values := fieldValuesFromRoot(m, idField)
			if len(values) == 0 {
				err = ErrDocumentMissingIDField
				return
			}

			documentID := values[0]
			_, err = index.IndexWithID(m, documentID)
			if err != nil {
				return
			}
		}

		err = scanner.Err()
		if err != nil {
			return err
		}
	}

	return nil
}
