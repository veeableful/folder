package main

import (
	"fmt"
	"log"

	"github.com/veeableful/folder"
)

func main() {
	index := folder.New()

	// Prepare the documents
	firstDocument := map[string]interface{}{
		"title": "Folder is a tiny little static search engine",
		"author": map[string]interface{}{
			"name": "Chae-Young Song",
			"hobbies": []string{"drawing", "gaming", "swimming"},
		},
	}
	secondDocument := map[string]interface{}{
		"title": "Folder v0.1.0 has been released!",
		"author": map[string]interface{}{
			"name": "Lilis Iskandar",
			"hobbies": []string{"cooking", "gardening", "hiking"},
		},
	}

	// Index the documents
	index.Index(firstDocument)
	index.Index(secondDocument)

	// Save the index to disk
	err := index.SaveToShards("index", 5)
	if err != nil {
		log.Fatal(err)
	}

	// Load the index from disk
	index, err = folder.LoadDeferred("index")
	if err != nil {
		log.Fatal(err)
	}

	// Search the index
	searchResult, err := index.Search("chaeyoung drawing")
	if err != nil {
		log.Fatal(err)
	}

	// Print the search results
	for _, hit := range searchResult.Hits {
		fmt.Println(hit.Source)
	}
}

