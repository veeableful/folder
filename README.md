# Folder

Folder is a search engine that can be embedded into Go programs or web apps via WebAssembly.

## Example

```go
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
```

## File formats

These file formats are not final and may change in the future.

**shard_count**

Contains the number of shards in the index.

**fns**

Newline-separated list of fields. Nested fields are joined together with their parent fields by dots.

**dcs**

Contains the documents in CSV format.

**tst**

Contains the term stats in CSV format.

## Development

### Structure

+ Main APIs are located in `folder.go`.
+ APIs that deal with I/O are located in `io.go` to separate core operations such as indexing / searching from I/O operations such as saving and loading indexes.
+ Internal code that may change often are located in `internal.go`.
+ Short utility functions are located in `util.go`.
+ Scripts are located inside the `scripts` directory.
+ Command-line tool packages such as `folder` are located inside the `cmd` directory.
+ Assets such as sample data files are located inside the `assets` directory.

### Setting up

For the development of this library, first we run `go generate` which runs a script that install the command-line tool and fetch additional, large files for indexing and testing.

To speed up testing of new code, we may also change where `github.com/veeableful/folder` points to by putting the following line in the `go.mod` files:
```
replace github.com/veeableful/folder => [your Folder directory path]
```

The project is not currently structured for development in Windows as the scripts are shell scripts only.

## Contributions

Contributions in any form are most welcome as I'm not familiar with search engine implementations myself and I don't have a CS degree. Nothing in this project is final so any kind of feedback or advice would be appreciated!

## License

Folder is BSD 3-clause licensed.