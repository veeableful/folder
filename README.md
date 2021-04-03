# Folder

Folder is a tiny search engine for storing and searching small number of documents. It is compatible with WebAssembly and may be compatible with TinyGo in the future as well. This is a work-in-progress.

## Usage

```go
// TODO
```

## File formats

These file formats are not final and may change in the future.

**\[index\].fns**

Newline-separated list of fields. Nested fields are joined together with their parent fields by dots.

**\[index\].dcs**

Contains the documents in CSV format.

**\[index\].dst**

Contains the document stats in CSV format.

**\[index\].tst**

Contains the term stats in CSV format.

## Development

### Structure

+ Main APIs are located in `folder.go`.
+ APIs that deal with I/O are located in `io.go` to separate core operations such as indexing / searching from I/O operations such as saving and loading indexes.
+ Internal code that may change often are located in `internal.go`.
+ Short utility functions are located in `util.go`.
+ Scripts are located inside the `scripts` directory.
+ Command-line tool packages such as `folder` are located inside the `cmd` directory.
+ Assets such as small sample data files are located inside the `assets` directory.

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