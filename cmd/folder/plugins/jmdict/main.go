package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/FooSoft/jmdict"
)

func Data(args ...interface{}) ([]byte, error) {
	file, err := os.Open(args[0].(string))
	if err != nil {
		return nil, err
	}

	dict, _, err := jmdict.LoadJmdict(file)
	if err != nil {
		return nil, err
	}

	lines := make([][]byte, 0)

	for _, entry := range dict.Entries {
		data, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}

		lines = append(lines, data)
	}

	return bytes.Join(lines, []byte{'\n'}), nil
}
