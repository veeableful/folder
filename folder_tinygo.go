// +build tinygo

package folder

import "errors"

var (
	ErrUnsupportedDocumentDataType = errors.New("unsupported document data type")
)

// IndexData indexes an array of bytes and assumes a certain data type such as text, JSON, or JSONL.
func (index *Index) IndexData(data []byte, dataType string) (err error) {
	var m map[string]interface{}

	if dataType == "text" {
		m = make(map[string]interface{})
		m["text"] = string(data)
		index.Index(m)
	} else {
		err = ErrUnsupportedDocumentDataType
	}

	return
}
