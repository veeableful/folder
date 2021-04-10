package folder

import (
	"io/fs"
	"strings"
	"time"
)

// Index contains all the information needed to search and return matching documents.
type Index struct {
	Name          string
	FieldNames    []string
	Documents     map[string]map[string]interface{}
	DocumentStats map[string]DocumentStat
	TermStats     map[string]TermStat
	LoadedShards  map[int]struct{}
	ShardCount    int
	f             fs.FS
}

// New creates an empty index.
func New() (index *Index) {
	index = &Index{}
	return
}

// DocumentStat contains information specific to documents.
type DocumentStat struct {
	TermFrequency map[string]int
}

// TermStat contains information specific to terms.
type TermStat struct {
	DocumentIDs []string
}

// IDScores is a structure used for sorting document IDs using their respective scores.
type IDScores struct {
	IDs    []string
	Scores []float64
}

// Len returns the number of document IDs.
func (ids IDScores) Len() int {
	return len(ids.IDs)
}

// Less compares the scores between two documents.
func (ids IDScores) Less(i, j int) bool {
	return ids.Scores[i] < ids.Scores[j]
}

// Swap swaps two documents and their respective scores in the arrays.
func (ids IDScores) Swap(i, j int) {
	ids.Scores[i], ids.Scores[j] = ids.Scores[j], ids.Scores[i]
	ids.IDs[i], ids.IDs[j] = ids.IDs[j], ids.IDs[i]
}

// SearchTime contains the elapsed times during various stages in the search process.
type SearchTime struct {
	Match time.Duration
	Sort  time.Duration
	Total time.Duration
}

// SearchResult contains the result of a search such as matching document count, the documents
// themselves with some metadata (a.k.a. the hits), and the search statistics.
type SearchResult struct {
	Count int
	Hits  []Hit
	Time  SearchTime
}

// Hit contains metadata of a document such as its ID and score, and also the document iself.
type Hit struct {
	ID     string
	Score  float64
	Source map[string]interface{}
}

// Index indexes a document into the index.
func (index *Index) Index(document map[string]interface{}) (documentID string, err error) {
	documentID = index.nextDocumentID()
	err = index.Update(documentID, document)
	return
}

// Update updates an existing document in the index with new data.
func (index *Index) Update(documentID string, document map[string]interface{}) (err error) {
	if index.Documents == nil {
		index.Documents = make(map[string]map[string]interface{})
	}

	err = index.Delete(documentID)
	if err != nil {
		return
	}

	index.Documents[documentID] = document
	err = index.index(documentID, document)
	if err != nil {
		return
	}
	return
}

// Delete deletes an existing document in the index.
func (index *Index) Delete(documentID string) (err error) {
	document, ok := index.Documents[documentID]
	if !ok {
		return
	}

	m := make(map[string][]string)
	index.analyze("", document, m)

	allTokens := MakeStringSet([]string{})
	for _, tokens := range m {
		for _, token := range tokens {
			allTokens.Add(token)
		}
	}

	delete(index.DocumentStats, documentID)
	index.removeDocumentFromTermStats(documentID, allTokens.List())
	return
}

// Search searches terms in an index and returns matching documents from the index along with some
// metadata.
func (index *Index) Search(s string) (res SearchResult) {
	var matchedDocumentIDs []string
	var sortedDocumentIDs []string
	var scores []float64

	startTime := time.Now()
	tokens := index.Analyze(s)
	matchedDocumentIDs, res.Time.Match = index.findDocuments(tokens)
	sortedDocumentIDs, scores, res.Time.Sort = index.sortDocuments(matchedDocumentIDs, tokens)
	res.Hits = index.fetchHits(sortedDocumentIDs, scores, 10)
	res.Count = len(sortedDocumentIDs)
	res.Time.Total = time.Since(startTime)
	return
}

// AnalyzeString breaks down string into list of tokens with some metadata such positions.
func (index *Index) Analyze(s string) (tokens []string) {
	s = strings.Trim(s, " ")
	tokens = strings.Split(s, " ")
	tokens = LowercaseFilter(tokens)
	tokens = PunctuationFilter(tokens)
	tokens = StopWordFilter(tokens)
	return
}
