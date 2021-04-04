package folder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:generate sh scripts/setup.sh

var index Index

func init() {
	index, _ = Load("index_test")
}

func TestAnalyze(t *testing.T) {
	expectedResult := []string{"my", "name", "lilis", "iskandar"}
	index := New()
	res := index.Analyze("My name is Lilis Iskandar")
	assert.Equal(t, res, expectedResult)
}
func TestIndexAndSearch(t *testing.T) {
	index := New()

	firstDocument := map[string]interface{}{
		"title": "Folder is a tiny little static search engine",
		"author": map[string]interface{}{
			"name": "Chae-Young Song",
		},
	}
	secondDocument := map[string]interface{}{
		"title": "Folder v0.1.0 has been released!",
		"author": map[string]interface{}{
			"name": "Lilis Iskandar",
		},
	}

	index.Index(firstDocument)
	index.Index(secondDocument)

	searchResult := index.Search("chaeyoung search")
	assert.Equal(t, len(searchResult.Hits), 1)
	assert.Equal(t, searchResult.Hits[0].Source, firstDocument)
}

func TestIndexAndSearchJSONLines(t *testing.T) {
	index := New()

	err := index.IndexFilePath("assets/users_test.jsonl", "jsonl")
	if err != nil {
		t.Fatal(err)
	}

	expectedResult := map[string]interface{}{
		"first_name": "Lilis",
		"last_name":  "Iskandar",
		"details": map[string]interface{}{
			"age":     28.0,
			"country": "Malaysia",
			"hobbies": []interface{}{"cooking", "gardening", "hiking"},
		},
	}

	res := index.Search("cooking")
	assert.Equal(t, len(res.Hits), 1)
	assert.Equal(t, res.Hits[0].Source, expectedResult)
}

func BenchmarkSearch(b *testing.B) {
	for n := 0; n < b.N; n++ {
		index.Search("ashtanga yoga los angeles")
	}
}

func TestUpdate(t *testing.T) {
	index := New()

	originalDocument := map[string]interface{}{
		"title": "Folder is a tiny little static search engine",
		"author": map[string]interface{}{
			"name": "Chae-Young Song",
		},
	}
	updatedDocument := map[string]interface{}{
		"title": "Folder v0.1.0 has been released!",
		"author": map[string]interface{}{
			"name": "Lilis Iskandar",
		},
	}

	documentID, err := index.Index(originalDocument)
	if err != nil {
		t.Fatal(err)
	}

	// We should be able to find the original document and not find the updated document
	searchResult := index.Search("chaeyoung search")
	assert.Equal(t, len(searchResult.Hits), 1)
	fmt.Printf("%+v\n", searchResult)

	searchResult = index.Search("lilis released")
	assert.Equal(t, len(searchResult.Hits), 0)

	// After update, we should be able to find the updated document and not find the original document
	index.Update(documentID, updatedDocument)

	searchResult = index.Search("chaeyoung search")
	assert.Equal(t, len(searchResult.Hits), 0)

	searchResult = index.Search("lilis released")
	assert.Equal(t, len(searchResult.Hits), 1)
	fmt.Printf("%+v\n", searchResult)
}
