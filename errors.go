package folder

import (
	"errors"
)

var (
	// ErrDocumentMissingIDField is returned when the document being indexed is missing a field that
	// is specified by the user to be used as the document ID.
	ErrDocumentMissingIDField = errors.New("document missing id field")
)
