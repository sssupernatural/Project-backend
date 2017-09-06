package types

const (
	defaultUserCacheSize   = 100
	defaultSearchResultMax = 1
)

type IndexerInitOptions struct {
	UserCacheSize   int
	SearchResultMax int
}

func (options *IndexerInitOptions) Init() {
	if options.UserCacheSize == 0 {
		options.UserCacheSize = defaultUserCacheSize
		options.SearchResultMax = defaultSearchResultMax
	}
}