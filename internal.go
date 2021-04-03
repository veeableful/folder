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

func (index *Index) index(documentID, fieldName string, fieldValue interface{}, parentFieldNames []string) (err error) {
	switch concreteFieldValue := fieldValue.(type) {
	case map[string]interface{}:
		for nextFieldName, nextFieldValue := range concreteFieldValue {
			parentFieldNames = append(parentFieldNames, fieldName)
			err = index.index(documentID, nextFieldName, nextFieldValue, parentFieldNames)
			if err != nil {
				return
			}
			parentFieldNames = parentFieldNames[:len(parentFieldNames)-1]
		}
	case string:
		fullFieldName := strings.Join(append(parentFieldNames, fieldName), ".")
		err = index.indexString(documentID, fullFieldName, concreteFieldValue)
		if err != nil {
			return
		}
	case *string:
		if concreteFieldValue == nil {
			return
		}
		fullFieldName := strings.Join(append(parentFieldNames, fieldName), ".")
		err = index.indexString(documentID, fullFieldName, *concreteFieldValue)
		if err != nil {
			return
		}
	case []interface{}:
		for _, nextFieldValue := range concreteFieldValue {
			err = index.index(documentID, fieldName, nextFieldValue, parentFieldNames)
			if err != nil {
				return
			}
		}
	case []string:
		for _, nextFieldValue := range concreteFieldValue {
			err = index.index(documentID, fieldName, nextFieldValue, parentFieldNames)
			if err != nil {
				return
			}
		}
	}
	return
}

func (index *Index) indexString(documentID, field, value string) (err error) {
	analyzeResult := index.Analyze(value)
	index.updateDocumentStat(documentID, analyzeResult.Tokens)
	index.updateTermStat(documentID, analyzeResult.Tokens)
	if contains(index.FieldNames, field) {
		return
	}
	index.FieldNames = append(index.FieldNames, field)
	return
}

func (index *Index) updateDocumentStat(documentID string, tokens []Token) {
	var documentStat DocumentStat

	if index.DocumentStats == nil {
		index.DocumentStats = make(map[string]DocumentStat)
	}

	documentStat = index.DocumentStats[documentID]

	for _, token := range tokens {
		if documentStat.TermFrequency == nil {
			documentStat.TermFrequency = map[string]int{token.Token: 1}
		} else {
			documentStat.TermFrequency[token.Token] += 1
		}
		if documentStat.TermFrequency == nil {
			documentStat.TermFrequency = map[string]int{token.Token: 1}
		} else {
			documentStat.TermFrequency[token.Token] += 1
		}
	}

	index.DocumentStats[documentID] = documentStat
}

func (index *Index) updateTermStat(documentID string, tokens []Token) {
	var termStat TermStat
	var ok bool

	if index.TermStats == nil {
		index.TermStats = make(map[string]TermStat)
	}

	for _, token := range tokens {
		termStat, ok = index.TermStats[token.Token]
		if ok {
			termStat.DocumentIDs = append(termStat.DocumentIDs, documentID)
		} else {
			if termStat.DocumentIDs == nil {
				termStat.DocumentIDs = []string{documentID}
			} else {
				termStat.DocumentIDs = append(termStat.DocumentIDs, documentID)
			}
		}
		index.TermStats[token.Token] = termStat
	}
}

func (index *Index) nextDocumentID() (id string) {
	id = generateRandomID(32)
	return
}

func (index *Index) findDocuments(tokens []Token) (documentIDs []string, elapsedTime time.Duration) {
	startTime := time.Now()

	var documentIDsSet StringSet

	for _, token := range tokens {
		ids := MakeStringSet([]string{})

		termStat, ok := index.TermStats[token.Token]
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

func (index *Index) sortDocuments(documentIDs []string, tokens []Token) (sortedDocumentIDs []string, sortedScores []float64, elapsedTime time.Duration) {
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

func (index *Index) calculateScore(documentID string, tokens []Token) (score float64) {
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
	for i, id := range documentIDs {
		hits = append(hits, Hit{
			ID:     id,
			Score:  scores[i],
			Source: index.Documents[id],
		})
		if len(hits) >= size {
			break
		}
	}
	return
}

// termFrequency returns the number of times a token appears in a certain document
func (index *Index) termFrequency(id string, token Token) (frequency int) {
	frequency = index.DocumentStats[id].TermFrequency[token.Token]
	return
}

// inverseDocumentFrequency calculates how rare a token is across all documents
func (index *Index) inverseDocumentFrequency(token Token) (frequency float64) {
	frequency = math.Log10(float64(len(index.Documents)) / float64(index.documentFrequency(token)))
	return
}

// documentFrequency returns the number of documents a token is available in
func (index *Index) documentFrequency(token Token) (frequency int) {
	frequency = len(index.TermStats[token.Token].DocumentIDs)
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
