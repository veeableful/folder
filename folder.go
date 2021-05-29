package folder

import (
	"io/fs"
	"time"
)

// Index contains all the information needed to search and return matching documents.
type Index struct {
	Name                      string
	FieldNames                []string
	Documents                 map[string]map[string]interface{}
	DocumentStats             map[string]DocumentStat
	TermStats                 map[string]TermStat
	LoadedDocumentsShards     map[int]struct{}
	LoadedDocumentStatsShards map[int]struct{}
	LoadedTermStatsShards     map[int]struct{}
	ShardCount                int
	f                         fs.FS
	baseURL                   string
}

// New creates an empty index.
func New() (index *Index) {
	index = &Index{}
	index.Documents = make(map[string]map[string]interface{})
	index.DocumentStats = make(map[string]DocumentStat)
	index.TermStats = make(map[string]TermStat)
	index.LoadedDocumentStatsShards = make(map[int]struct{})
	index.LoadedDocumentsShards = make(map[int]struct{})
	index.LoadedTermStatsShards = make(map[int]struct{})
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

// SearchOptions contains options that can be used to alter the search operation and result.
type SearchOptions struct {
	UseCache bool // Whether to use and/or keep relevant data in memory
	Size     int  // Number of documents to return
	From     int  // Starting offset for returned documents
}

// DefaultSearchOptions returns the default search options.
var DefaultSearchOptions = SearchOptions{
	UseCache: true,
	Size:     10,
	From:     0,
}

// Hit contains metadata of a document such as its ID and score, and also the document iself.
type Hit struct {
	ID     string
	Score  float64
	Source map[string]interface{}
}

// IndexWithID indexes a document into the index but with user-specified document ID.
func (index *Index) IndexWithID(document map[string]interface{}, desiredDocumentID string) (documentID string, err error) {
	documentID = desiredDocumentID
	err = index.Update(documentID, document)
	return
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

	debug("Delete", documentID)

	m := make(map[string][]string)
	index.analyze("", document, m)

	allTokens := MakeStringSet([]string{})
	for _, tokens := range m {
		for _, token := range tokens {
			allTokens.Add(token)
		}
	}

	delete(index.DocumentStats, documentID)
	err = index.removeDocumentFromTermStats(documentID, allTokens.List())
	return
}

// Fetch fetches a document with specific ID.
func (index *Index) Fetch(documentID string) (document map[string]interface{}, err error) {
	shardID := index.CalculateShardID(documentID)
	document, err = index.fetchDocumentFromShard(shardID, documentID)
	return
}

// Search searches terms in an index and returns matching documents from the index along with some
// metadata. It is equivalent to SearchWithOptions using the default search options.
func (index *Index) Search(s string) (res SearchResult, err error) {
	return index.SearchWithOptions(s, DefaultSearchOptions)
}

// SearchWithOptions searches a term just like Search but it also accepts user-provided SearchOptions.
func (index *Index) SearchWithOptions(s string, opts SearchOptions) (res SearchResult, err error) {
	if !opts.UseCache {
		debug("Search", s, "(not cached)")
		tmp := New()
		tmp.Name = index.Name
		tmp.ShardCount = index.ShardCount
		tmp.f = index.f
		tmp.baseURL = index.baseURL
		return tmp.searchWithOptions(s, opts)
	}
	debug("Search", s, "(cached)")
	return index.searchWithOptions(s, opts)
}

func (index *Index) searchWithOptions(s string, opts SearchOptions) (res SearchResult, err error) {
	var matchedDocumentIDs []string
	var sortedDocumentIDs []string
	var scores []float64

	startTime := time.Now()
	tokens := index.Analyze(s)
	debug("  Analyzed", s, "into", tokens)

	matchedDocumentIDs, res.Time.Match, err = index.findDocuments(tokens)
	if err != nil {
		return
	}

	sortedDocumentIDs, scores, res.Time.Sort, err = index.sortDocuments(matchedDocumentIDs, tokens)
	if err != nil {
		return
	}

	res.Hits, err = index.fetchHits(sortedDocumentIDs, scores, opts.Size, opts.From)
	if err != nil {
		return
	}

	res.Count = len(sortedDocumentIDs)
	res.Time.Total = time.Since(startTime)
	return
}

// AnalyzeString breaks down string into list of tokens with some metadata such positions.
func (index *Index) Analyze(s string) (tokens []string) {
	tokens = splitWithRunes(s, "、,　 ")
	tokens = LowercaseFilter(tokens)
	tokens = PunctuationFilter(tokens)
	tokens = StopWordFilter(tokens)
	return
}

func (index *Index) CalculateShardID(s string) (shardID uint32) {
	const Q = 123456789

	result := uint32(0)
	for _, r := range s {
		result += Q + uint32(r)*uint32(r)
	}
	result *= Q
	shardID = uint32(result % uint32(index.ShardCount))
	return
}
