package filter

import "net/http"

type Filter func(r *http.Request) bool

type MultiFilter Filter

func NewMultiFilter(fs ...Filter) MultiFilter {
	return func(r *http.Request) bool {
		for _, f := range fs {
			if !f(r) {
				return false
			}
		}

		return true
	}
}

func (mf MultiFilter) Append(f Filter) MultiFilter {
	return NewMultiFilter(Filter(mf), f)
}
