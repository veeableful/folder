package folder

import (
	"strings"
	"time"
)

// Index contains all the information needed to search and return matching documents.
type Index struct {
	FieldNames    []string
	Documents     map[string]map[string]interface{}
	DocumentStats map[string]DocumentStat
	TermStats     map[string]TermStat
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

// AnalyzeResult contains the tokens extracted from an analysis.
type AnalyzeResult struct {
	Tokens []Token
}

// Token contains the token string and its position within the analyzed text.
type Token struct {
	Token    string
	Position int
}

// Hit contains metadata of a document such as its ID and score, and also the document iself.
type Hit struct {
	ID     string
	Score  float64
	Source map[string]interface{}
}

// Index indexes a document into the index.
func (index *Index) Index(document map[string]interface{}) (err error) {
	documentID := index.nextDocumentID()

	// Store the documents a.k.a. field values
	if index.Documents == nil {
		index.Documents = make(map[string]map[string]interface{})
	}
	index.Documents[documentID] = document

	// Store field term stats
	for fieldName, fieldValue := range document {
		parentFieldNames := []string{}
		err = index.index(documentID, fieldName, fieldValue, parentFieldNames)
		if err != nil {
			return
		}
	}

	return
}

// Search searches terms in an index and returns matching documents from the index along with some
// metadata.
func (index *Index) Search(s string) (res SearchResult) {
	var matchedDocumentIDs []string
	var sortedDocumentIDs []string
	var scores []float64

	startTime := time.Now()
	analyzeResult := index.Analyze(s)
	matchedDocumentIDs, res.Time.Match = index.findDocuments(analyzeResult.Tokens)
	sortedDocumentIDs, scores, res.Time.Sort = index.sortDocuments(matchedDocumentIDs, analyzeResult.Tokens)
	res.Hits = index.fetchHits(sortedDocumentIDs, scores, 10)
	res.Count = len(sortedDocumentIDs)
	res.Time.Total = time.Since(startTime)

	return
}

// Analyze breaks down search terms into list of tokens with some metadata such positions.
func (index *Index) Analyze(s string) (res AnalyzeResult) {
	s = strings.Trim(s, " ")
	ss := strings.Split(s, " ")

	tokens := res.Tokens
	for i, v := range ss {
		tokens = append(tokens, Token{
			Token:    v,
			Position: i,
		})
	}

	tokens = LowercaseFilter(tokens)
	tokens = PunctuationFilter(tokens)
	tokens = StopWordFilter(tokens)
	res.Tokens = tokens

	return
}
