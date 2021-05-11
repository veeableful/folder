// +build js

package main

import (
	"syscall/js"

	"github.com/veeableful/folder"
)

func main() {
	js.Global().Set("folder", js.ValueOf(
		map[string]interface{}{
			"load": jsLoad(),
		},
	))
	select {}
}

func load(indexName, baseURL string) (index *folder.Index, err error) {
	index, err = folder.LoadDeferred(indexName, baseURL)
	return
}

func jsLoad() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		indexName := args[0].String()
		baseURL := ""
		if len(args) >= 2 {
			baseURL = args[1].String()
		} else {
			baseURL = js.Global().Get("location").Get("origin").String()
		}

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]
			go func() {
				index, err := load(indexName, baseURL)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				jsIndex := js.ValueOf(map[string]interface{}{
					"search":     jsSearch(index),
					"fetch":      jsFetch(index),
					"shardCount": js.ValueOf(index.ShardCount),
				})
				resolve.Invoke(jsIndex)
			}()
			return nil
		})

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func jsSearch(index *folder.Index) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		query := args[0].String()
		opts := folder.DefaultSearchOptions
		opts.UseCache = false // Prevents folder from populating the memory overtime by default.
		if len(args) >= 2 {
			opts.UseCache = args[1].Get("useCache").Bool()
		}

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]
			go func() {
				result, err := index.SearchWithOptions(query, opts)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}
				resolve.Invoke(jsSearchResultValue(result))
			}()
			return nil
		})

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func jsFetch(index *folder.Index) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		documentID := args[0].String()

		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]
			go func() {
				document, err := index.Fetch(documentID)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}
				resolve.Invoke(js.ValueOf(document))
			}()
			return nil
		})

		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func jsHitsValue(hits []folder.Hit) js.Value {
	vs := []interface{}{}
	for _, hit := range hits {
		vs = append(vs, jsHitValue(hit))
	}
	return js.ValueOf(vs)
}

func jsHitValue(hit folder.Hit) js.Value {
	return js.ValueOf(map[string]interface{}{
		"_id":     js.ValueOf(hit.ID),
		"_score":  js.ValueOf(hit.Score),
		"_source": js.ValueOf(hit.Source),
	})
}

func jsSearchResultValue(result folder.SearchResult) js.Value {
	return js.ValueOf(map[string]interface{}{
		"hits": js.ValueOf(jsHitsValue(result.Hits)),
		"time": js.ValueOf(map[string]interface{}{
			"match": js.ValueOf(float64(result.Time.Match)),
			"sort":  js.ValueOf(float64(result.Time.Sort)),
			"total": js.ValueOf(float64(result.Time.Total)),
		}),
		"count": js.ValueOf(result.Count),
	})
}
