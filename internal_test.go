package folder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testDocument = map[string]interface{}{
		"project": "Folder",
		"author": []map[string]interface{}{
			{
				"name": "Lilis Iskandar",
				"details": map[string]interface{}{
					"age":      28,
					"location": "Malaysia",
				},
				"coworkers": []map[string]interface{}{
					{
						"name": "Chae-Young Song",
						"details": map[string]interface{}{
							"age":      26,
							"location": "South Korea",
						},
					},
				},
			},
		},
	}
)

func TestSetField(t *testing.T) {
	document := make(map[string]interface{})

	setField(document, "name", "Lilis Iskandar")
	setField(document, "details.location", "Malaysia")

	expectedResult := map[string]interface{}{
		"name": "Lilis Iskandar",
		"details": map[string]interface{}{
			"location": "Malaysia",
		},
	}

	assert.Equal(t, document, expectedResult)
}

func TestValueFromMapStringInterface(t *testing.T) {
	assert.Equal(t, fieldValuesFromRoot(testDocument, "project"), []string{"Folder"})
	assert.Equal(t, fieldValuesFromRoot(testDocument, "author.name"), []string{"Lilis Iskandar"})
	assert.Equal(t, fieldValuesFromRoot(testDocument, "author.coworkers.name"), []string{"Chae-Young Song"})
	assert.Equal(t, fieldValuesFromRoot(testDocument, "author.details.age"), []string{"28"})
}

func TestInternalAnalyze(t *testing.T) {
	expectedResult := map[string][]string{
		"project":                           {"folder"},
		"author.name":                       {"lilis", "iskandar"},
		"author.details.location":           {"malaysia"},
		"author.coworkers.name":             {"chaeyoung", "song"},
		"author.coworkers.details.location": {"south", "korea"},
	}
	m := make(map[string][]string)
	index := New()
	index.analyze("", testDocument, m)
	assert.Equal(t, m, expectedResult)
}
