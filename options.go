package chatgpt

import (
	"net/url"
	"strconv"
)

type ListOptions struct {
	After *string // Identifier for the last event from the previous pagination request.
	Limit *int    // Number of events to retrieve. Defaults to 20.
}

func (opts *ListOptions) Encode() string {
	if opts == nil {
		return ""
	}

	values := url.Values{}
	if opts.After != nil {
		values.Add("after", *opts.After)
	}
	if opts.After != nil {
		values.Add("limit", strconv.Itoa(*opts.Limit))
	}

	e := values.Encode()
	if e == "" {
		return e
	}

	return "?" + e
}
