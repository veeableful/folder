// +build !tinygo

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/veeableful/folder"
)

func doIndex(c *cli.Context) (err error) {
	var index *folder.Index

	if c.NArg() <= 0 {
		err = errors.New("please specify the document file or directory path")
		return
	}

	filePath := c.Args().Get(0)
	dataType := c.String("type")
	pluginName := c.String("plugin")
	indexName := c.String("index")
	idField := c.String("id-field")

	var info os.FileInfo
	info, err = os.Stat(filePath)
	if err != nil {
		return
	}

	index, err = folder.Load(indexName)
	if err != nil {
		if err2 := err.(*fs.PathError); err2.Err == syscall.ENOENT {
			// do nothing
		} else {
			return
		}
	}

	if pluginName != "" {
		var p *plugin.Plugin
		var sym plugin.Symbol
		var data []byte

		p, err = plugin.Open(path.Join("plugins", pluginName, pluginName+".so"))
		if err != nil {
			return
		}

		sym, err = p.Lookup("Data")
		if err != nil {
			return
		}

		data, err = sym.(func(args ...interface{}) ([]byte, error))(c.Args().First())
		if err != nil {
			return
		}

		if idField != "" {
			err = index.IndexDataWithIDField(data, dataType, idField)
			if err != nil {
				return
			}
		} else {
			err = index.IndexData(data, dataType)
			if err != nil {
				return
			}
		}
	} else {
		if info.IsDir() {
			err = filepath.WalkDir(filePath, func(path string, d os.DirEntry, err error) error {
				if idField != "" {
					return index.IndexFilePathWithIDField(path, dataType, idField)
				} else {
					return index.IndexFilePath(path, dataType)
				}
			})
		} else {
			if idField != "" {
				err = index.IndexFilePathWithIDField(filePath, dataType, idField)
			} else {
				err = index.IndexFilePath(filePath, dataType)
			}
		}
		if err != nil {
			return err
		}
	}

	err = index.SaveToShards(indexName, c.Int("shards"))
	if err != nil {
		return
	}

	return
}

func doSearch(c *cli.Context) error {
	indexName := c.String("index")

	index, err := folder.LoadDeferred(indexName)
	if err != nil {
		return err
	}

	s := c.Args().First()
	format := c.String("format")
	result, err := index.Search(s) // Load related shards to memory first
	if err != nil {
		log.Fatal(err)
	}

	result, err = index.Search(s)
	if err != nil {
		log.Fatal(err)
	}

	if format == "go" {
		fmt.Printf("%+v\n", result)
	} else if format == "json" {
		data, _ := json.Marshal(result)
		fmt.Printf("%s\n", string(data))
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:  "folder",
		Usage: "Folder is a utility program for testing indexing and searching documents",
		Commands: []*cli.Command{
			{
				Name:    "index",
				Aliases: []string{"i"},
				Usage:   "Index a document",
				Action:  doIndex,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Usage: "Document file data type [text, json, jsonl]",
						Value: "text",
					},
					&cli.StringFlag{
						Name:  "plugin",
						Usage: "Plugin name that provides documents to index",
					},
					&cli.StringFlag{
						Name:  "index",
						Usage: "Name of the index",
						Value: "index",
					},
					&cli.IntFlag{
						Name:  "shards",
						Usage: "Number of shards",
						Value: 1000,
					},
					&cli.StringFlag{
						Name:  "id-field",
						Usage: "Field to be used for document ID",
						Value: "",
					},
				},
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "Search documents containing specified terms",
				Action:  doSearch,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Usage: "Format of the search result output [go, json]",
						Value: "go",
					},
					&cli.StringFlag{
						Name:  "index",
						Usage: "Name of the index",
						Value: "index",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
