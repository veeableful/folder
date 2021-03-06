package folder

import (
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// IDScores is a structure used for sorting document IDs using their respective scores.
type IDScores struct {
	IDs    sort.StringSlice
	Scores sort.Float64Slice
}

// Len returns the number of document IDs.
func (ids IDScores) Len() int {
	return len(ids.IDs)
}

// Less compares the scores between two documents.
func (ids IDScores) Less(i, j int) bool {
	if ids.Scores[i] == ids.Scores[j] {
		return !ids.IDs.Less(i, j) // Put document IDs with "lower" values higher in the list
	}
	return ids.Scores.Less(i, j)
}

// Swap swaps two documents and their respective scores in the arrays.
func (ids IDScores) Swap(i, j int) {
	ids.Scores[i], ids.Scores[j] = ids.Scores[j], ids.Scores[i]
	ids.IDs[i], ids.IDs[j] = ids.IDs[j], ids.IDs[i]
}

// setField sets a value in a document at a field path
// e.g. setField(doc, "author.name", "Lilis Iskandar")
func setField(document map[string]interface{}, fieldPath, record string) {
	var it interface{} = document
	var ok bool

	fields := strings.Split(fieldPath, ".")
	level := len(fields)

	for _, field := range fields {
		level -= 1

		switch t := it.(type) {
		case map[string]interface{}:
			it, ok = t[field]
			if ok {
				continue
			}

			if level > 0 {
				t[field] = make(map[string]interface{})
			} else {
				t[field] = record
			}
			it = t[field]
		}
	}
}

// analyze analyzes an arbitrary value and returns the tokens for each field
func (index *Index) analyze(parentField string, v interface{}, m map[string][]string) {
	if m == nil {
		return
	}

	switch value := v.(type) {
	case map[string]interface{}:
		if len(parentField) > 0 {
			debug("  Analyze field " + parentField + ": map[string]interface{}")
		}
		for field, value := range value {
			if parentField != "" {
				field = parentField + "." + field
			}
			index.analyze(field, value, m)
		}
	case []map[string]interface{}:
		debug("  Analyze field " + parentField + ": []map[string]interface{}")
		for _, v := range value {
			index.analyze(parentField, v, m)
		}
	case []interface{}:
		debug("  Analyze field " + parentField + ": []interface{}")
		for _, v := range value {
			index.analyze(parentField, v, m)
		}
	case []string:
		debug("  Analyze field " + parentField + ": []string")
		tokens := []string{}
		for _, v := range value {
			tokens = append(tokens, index.Analyze(v)...)
		}
		m[parentField] = append(m[parentField], tokens...)
	case *string:
		debug("  Analyze field " + parentField + ": *string")
		if value != nil {
			m[parentField] = append(m[parentField], index.Analyze(*value)...)
		}
	case string:
		debug("  Analyze field " + parentField + ": string")
		m[parentField] = append(m[parentField], index.Analyze(value)...)
	case int:
		// TODO
	}
}

func (index *Index) index(documentID string, document map[string]interface{}) (err error) {
	debug("Index", documentID)
	m := make(map[string][]string)
	index.analyze("", document, m)
	for field, tokens := range m {
		debug("  Index field", field)
		index.indexTokens(documentID, field, tokens)
	}
	return
}

func (index *Index) indexTokens(documentID string, field string, tokens []string) (err error) {
	err = index.updateTermStat(documentID, tokens)
	if err != nil {
		return
	}
	if contains(index.FieldNames, field) {
		return
	}

	debug("  Add new field name", field)
	index.FieldNames = append(index.FieldNames, field)
	return
}

func (index *Index) updateTermStat(documentID string, tokens []string) (err error) {
	var termStat TermStat
	var ok bool

	debug("  Update term stat in", documentID, "for tokens", tokens)

	if index.TermStats == nil {
		index.TermStats = make(map[string]TermStat)
	}

	for _, token := range tokens {
		termStat, ok, err = index.fetchTermStat(token)
		if err != nil {
			return
		}
		if ok {
			termStat.TermFrequencies[documentID] += 1
		} else {
			if termStat.TermFrequencies == nil {
				termStat.TermFrequencies = map[string]int{documentID: 1}
			} else {
				termStat.TermFrequencies[documentID] += 1
			}
		}
		index.TermStats[token] = termStat
	}

	return
}

func (index *Index) removeDocumentFromTermStats(documentID string, tokens []string) (err error) {
	var termStat TermStat
	var ok bool

	debug("  Remove document ID", documentID, "from term stats with tokens", tokens)

	for _, token := range tokens {
		termStat, ok, err = index.fetchTermStat(token)
		if err != nil {
			return
		}
		if !ok {
			continue
		}

		delete(termStat.TermFrequencies, documentID)
		index.TermStats[token] = termStat
	}

	return
}

func (index *Index) nextDocumentID() (id string) {
	for {
		id = generateRandomID(8)
		if _, ok := index.Documents[id]; !ok {
			break
		}
	}
	return
}

// findDocuments finds document IDs which contain the tokens. The more tokens provided, the fewer number of documents would be found as they are narrowed down.
func (index *Index) findDocuments(tokens []string) (documentIDs []string, elapsedTime time.Duration, err error) {
	var documentIDsSet StringSet
	var termStat TermStat
	var ok bool

	startTime := time.Now()
	debug("  Find document IDs with tokens", tokens)

	for _, token := range tokens {
		ids := MakeStringSet([]string{})

		termStat, ok, err = index.fetchTermStat(token)
		if !ok {
			continue
		}
		if err != nil {
			return
		}

		for id, _ := range termStat.TermFrequencies {
			ids.Add(id)
		}

		if documentIDsSet.Len() == 0 {
			documentIDsSet = ids
		} else if documentIDsSet.Len() == 1 {
			break
		} else {
			documentIDsSet.Intersects(ids)
		}
	}

	documentIDs = documentIDsSet.List()
	elapsedTime = time.Since(startTime)
	return
}

func (index *Index) fetchTermStat(token string) (termStat TermStat, ok bool, err error) {
	termStat, ok = index.TermStats[token]
	if ok || index.ShardCount == 0 {
		return
	}

	shardID := index.CalculateShardID(token)
	err = index.loadTermStatsFromShard(shardID)
	if err != nil {
		return
	}

	termStat, ok = index.TermStats[token]
	return
}

func (index *Index) sortDocuments(documentIDs []string, tokens []string) (sortedDocumentIDs []string, sortedScores []float64, elapsedTime time.Duration, err error) {
	startTime := time.Now()

	debug("  Sort", len(documentIDs), "documents with tokens", tokens)

	scores := make([]float64, len(documentIDs))
	for i, id := range documentIDs {
		scores[i], err = index.CalculateScore(id, tokens)
		if err != nil {
			return
		}
	}
	idScores := IDScores{IDs: documentIDs, Scores: scores}
	sort.Sort(sort.Reverse(idScores))

	sortedDocumentIDs = idScores.IDs
	sortedScores = idScores.Scores

	elapsedTime = time.Since(startTime)
	return
}

func (index *Index) CalculateScore(documentID string, tokens []string) (score float64, err error) {
	var tf int

	for _, token := range tokens {
		tf, err = index.termFrequency(documentID, token)
		if err != nil {
			return
		}

		score += float64(tf) * float64(index.inverseDocumentFrequency(token))
	}

	return
}

func (index *Index) fetchHits(documentIDs []string, scores []float64, size, from int) (hits []Hit, err error) {
	var document map[string]interface{}

	if from < 0 {
		from = 0
	}

	n := len(documentIDs)
	if size == 0 || from >= n {
		return
	}
	if n > size {
		n = size
	}
	debug("  Fetch", n, "documents")

	hits = make([]Hit, 0)
	for i, documentID := range documentIDs[from : from+n] {
		document, _, err = index.fetchDocument(documentID)
		if err != nil {
			return
		}

		hits = append(hits, Hit{
			ID:     documentID,
			Score:  scores[from+i],
			Source: document,
		})
	}
	return
}

func (index *Index) fetchDocument(documentID string) (document map[string]interface{}, ok bool, err error) {
	document, ok = index.Documents[documentID]
	if ok || index.ShardCount == 0 {
		return
	}

	shardID := index.CalculateShardID(documentID)
	err = index.loadDocumentsFromShard(shardID)
	if err != nil {
		return
	}

	document, ok = index.Documents[documentID]
	return
}

// termFrequency returns the number of times a token appears in a certain document
func (index *Index) termFrequency(documentID, token string) (frequency int, err error) {
	var termStat TermStat
	var ok bool

	termStat, _, err = index.fetchTermStat(token)
	if err != nil {
		return
	}

	frequency, ok = termStat.TermFrequencies[documentID]
	if !ok {
		return
	}

	return
}

// inverseDocumentFrequency calculates how rare a token is across all documents
func (index *Index) inverseDocumentFrequency(token string) (frequency float64) {
	frequency = math.Log10(float64(len(index.Documents)) / float64(index.documentFrequency(token)))
	return
}

// documentFrequency returns the number of documents a token is available in
func (index *Index) documentFrequency(token string) (frequency int) {
	frequency = len(index.TermStats[token].TermFrequencies)
	return
}

func fieldValuesFromRoot(document map[string]interface{}, fieldPath string) (values []string) {
	fields := strings.Split(fieldPath, ".")
	values = fieldValuesFromMapStringInterface(document, fields, 0)
	return
}

func fieldValuesFromMapStringInterface(document map[string]interface{}, fields []string, depth int) (values []string) {
	it, ok := document[fields[depth]]
	if !ok {
		return
	}
	switch t := it.(type) {
	case []interface{}:
		values = append(values, fieldValuesFromArrayInterface(t, fields, depth)...)
	case string:
		values = []string{t}
	case *string:
		if t == nil {
			return
		}
		values = []string{*t}
	case []string:
		values = append(values, t...)
	case []map[string]interface{}:
		for _, node := range t {
			values = append(values, fieldValuesFromMapStringInterface(node, fields, depth+1)...)
		}
	case map[string]interface{}:
		values = append(values, fieldValuesFromMapStringInterface(t, fields, depth+1)...)
	case float64:
		values = append(values, strconv.FormatFloat(t, 'g', 'g', 64))
	case int:
		values = append(values, strconv.FormatInt(int64(t), 10))
	default:
		debug("fieldValuesFromMapStringInterface(): Unimplemented for", reflect.TypeOf(t))
	}
	return
}

func fieldValuesFromArrayInterface(node []interface{}, fields []string, depth int) (values []string) {
	for _, v := range node {
		switch value := v.(type) {
		case string:
			values = append(values, value)
		case *string:
			values = append(values, *value)
		case []string:
			values = append(values, value...)
		case []interface{}:
			values = append(values, fieldValuesFromArrayInterface(value, fields, depth)...)
		case map[string]interface{}:
			values = append(values, fieldValuesFromMapStringInterface(value, fields, depth+1)...)
		default:
			debug("fieldValuesFromArrayInterface(): Unimplemented for", reflect.TypeOf(value))
		}
	}
	return
}
