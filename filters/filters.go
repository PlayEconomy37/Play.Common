package filters

import (
	"strings"

	"github.com/PlayEconomy37/Play.Common/validator"
)

// Holds filtering parameters
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string // Supported sort column values
}

// Validate filters received as query parameters
func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values
	v.Check(validator.Between(f.Page, 0, 10_000_000), "page", "must be greater or equal to 0 and lower or equal to 10 million")
	v.Check(validator.Between(f.PageSize, 0, 100), "page_size", "must be greater or equal to 0 and lower or equal to 100")

	// Check that the sort parameter matches a value in the safelist
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// Check that the client-provided `Sort` field matches one of the entries in our safelist
// and if it does, extract the column name from the `Sort` field by stripping the leading
// hyphen character (if one exists)
func (f Filters) SortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// Prevent SQL injection attack.
	// Before calling this method, we should have validated the Sort field
	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction ("ASC" or "DESC") depending on the prefix character of the
// Sort field
func (f Filters) SortDirection() int8 {
	// Descending order
	if strings.HasPrefix(f.Sort, "-") {
		return -1
	}

	// Ascending order
	return 1
}

// Returns the number of records to be returned in the query
func (f Filters) Limit() int {
	return f.PageSize
}

// Returns the number of rows to skip before starting to return records from the query
func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}
