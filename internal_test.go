package folder

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	document := map[string]interface{}{
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

	assert.Equal(t, fieldValuesFromRoot(document, "project"), []string{"Folder"})
	assert.Equal(t, fieldValuesFromRoot(document, "author.name"), []string{"Lilis Iskandar"})
	assert.Equal(t, fieldValuesFromRoot(document, "author.coworkers.name"), []string{"Chae-Young Song"})
}
