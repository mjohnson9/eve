package eve

import "time"

// The most common time format used in the EVE API.
const TimeFormat = "2006-01-02 15:04:05"

// The timezone that the EVE API uses. New Eden Dev says that this is server
// local time, which should be UTC.
var Timezone = time.UTC

// Metadata is the metadata gained from parsing XML inside of Decode.
type Metadata struct {
	// Expires gives the result of cachedUntil.
	Expires time.Time

	// RowSets are the row sets found inside of the parsed XML.
	RowSets []*RowSet
}

// RowSet is information about a single row set.
type RowSet struct {
	// Name contains the name of the row set.
	Name string
	// Keys contain the keys of the row set, in order of importance.
	Keys []string
	// Columns contain the columns present in the row set.
	Columns []string
}
