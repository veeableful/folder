package folder

import (
	"math"
	"sort"
	"strings"
	"time"
)

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
		for field, value := range value {
			if parentField != "" {
				field = parentField + "." + field
			}
			index.analyze(field, value, m)
		}
	case []map[string]interface{}:
		for _, v := range value {
			index.analyze(parentField, v, m)
		}
	case []interface{}:
		for _, v := range value {
			index.analyze(parentField, v, m)
		}
	case []string:
		tokens := []string{}
		for _, v := range value {
			tokens = append(tokens, index.Analyze(v)...)
		}
		m[parentField] = tokens
	case *string:
		if value != nil {
			m[parentField] = index.Analyze(*value)
		}
	case string:
		m[parentField] = index.Analyze(value)
	case int:
		// TODO
	}
}

func (index *Index) index(documentID string, document map[string]interface{}) (err error) {
	m := make(map[string][]string)
	index.analyze("", document, m)
	for field, tokens := range m {
		index.indexTokens(documentID, field, tokens)
	}
	return
}

func (index *Index) indexTokens(documentID string, field string, tokens []string) (err error) {
	index.updateDocumentStat(documentID, tokens)
	index.updateTermStat(documentID, tokens)
	if contains(index.FieldNames, field) {
		return
	}
	index.FieldNames = append(index.FieldNames, field)
	return
}

func (index *Index) updateDocumentStat(documentID string, tokens []string) {
	var documentStat DocumentStat

	if index.DocumentStats == nil {
		index.DocumentStats = make(map[string]DocumentStat)
	}

	documentStat, _ = index.fetchDocumentStat(documentID)

	for _, token := range tokens {
		if documentStat.TermFrequency == nil {
			documentStat.TermFrequency = map[string]int{token: 1}
		} else {
			documentStat.TermFrequency[token] += 1
		}
		if documentStat.TermFrequency == nil {
			documentStat.TermFrequency = map[string]int{token: 1}
		} else {
			documentStat.TermFrequency[token] += 1
		}
	}

	index.DocumentStats[documentID] = documentStat
}

func (index *Index) fetchDocumentStat(documentID string) (documentStat DocumentStat, ok bool) {
	documentStat, ok = index.DocumentStats[documentID]
	if !ok && index.ShardCount > 0 {
		shardID := index.calculateShardID(documentID)
		index.loadShard(shardID)
		documentStat, ok = index.DocumentStats[documentID]
	}
	return
}

func (index *Index) updateTermStat(documentID string, tokens []string) {
	var termStat TermStat
	var ok bool

	if index.TermStats == nil {
		index.TermStats = make(map[string]TermStat)
	}

	for _, token := range tokens {
		termStat, ok = index.fetchTermStat(token)
		if ok {
			termStat.DocumentIDs = append(termStat.DocumentIDs, documentID)
		} else {
			if termStat.DocumentIDs == nil {
				termStat.DocumentIDs = []string{documentID}
			} else {
				termStat.DocumentIDs = append(termStat.DocumentIDs, documentID)
			}
		}
		index.TermStats[token] = termStat
	}
}

func (index *Index) removeDocumentFromTermStats(documentID string, tokens []string) {
	for _, token := range tokens {
		termStat, ok := index.fetchTermStat(token)
		if !ok {
			continue
		}

		termStat.DocumentIDs = remove(termStat.DocumentIDs, documentID)
		index.TermStats[token] = termStat
	}
}

func (index *Index) nextDocumentID() (id string) {
	id = generateRandomID(32)
	return
}

func (index *Index) findDocuments(tokens []string) (documentIDs []string, elapsedTime time.Duration) {
	startTime := time.Now()

	var documentIDsSet StringSet

	for _, token := range tokens {
		ids := MakeStringSet([]string{})

		termStat, ok := index.fetchTermStat(token)
		if !ok {
			continue
		}

		for _, id := range termStat.DocumentIDs {
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

func (index *Index) fetchTermStat(token string) (termStat TermStat, ok bool) {
	termStat, ok = index.TermStats[token]
	if !ok && index.ShardCount > 0 {
		shardID := index.calculateShardID(token)
		index.loadShard(shardID)
		termStat, ok = index.TermStats[token]
	}
	return
}

func (index *Index) sortDocuments(documentIDs []string, tokens []string) (sortedDocumentIDs []string, sortedScores []float64, elapsedTime time.Duration) {
	startTime := time.Now()

	scores := make([]float64, len(documentIDs))
	for i, id := range documentIDs {
		scores[i] = index.calculateScore(id, tokens)
	}
	idScores := IDScores{IDs: documentIDs, Scores: scores}
	sort.Sort(sort.Reverse(idScores))

	sortedDocumentIDs = idScores.IDs
	sortedScores = idScores.Scores

	elapsedTime = time.Since(startTime)
	return
}

func (index *Index) calculateScore(documentID string, tokens []string) (score float64) {
	for _, token := range tokens {
		score += float64(index.termFrequency(documentID, token)) * float64(index.inverseDocumentFrequency(token))
	}
	return
}

func (index *Index) fetchHits(documentIDs []string, scores []float64, size int) (hits []Hit) {
	if size == 0 {
		return
	}

	hits = make([]Hit, 0)
	for i, documentID := range documentIDs {
		document, _ := index.fetchDocument(documentID)
		hits = append(hits, Hit{
			ID:     documentID,
			Score:  scores[i],
			Source: document,
		})
		if len(hits) >= size {
			break
		}
	}
	return
}

func (index *Index) fetchDocument(documentID string) (document map[string]interface{}, ok bool) {
	document, ok = index.Documents[documentID]
	if !ok && index.ShardCount > 0 {
		shardID := index.calculateShardID(documentID)
		index.loadShard(shardID)
		document, ok = index.Documents[documentID]
	}
	return
}

// termFrequency returns the number of times a token appears in a certain document
func (index *Index) termFrequency(documentID, token string) (frequency int) {
	documentStat, _ := index.fetchDocumentStat(documentID)
	frequency = documentStat.TermFrequency[token]
	return
}

// inverseDocumentFrequency calculates how rare a token is across all documents
func (index *Index) inverseDocumentFrequency(token string) (frequency float64) {
	frequency = math.Log10(float64(len(index.Documents)) / float64(index.documentFrequency(token)))
	return
}

// documentFrequency returns the number of documents a token is available in
func (index *Index) documentFrequency(token string) (frequency int) {
	frequency = len(index.TermStats[token].DocumentIDs)
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
		}
	}
	return
}
