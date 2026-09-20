package qualtrics

import (
	"context"
	"net/url"
	"strconv"
)

type ListOptions struct {
	Offset int
	Limit  int
}

func (o ListOptions) Query() url.Values {
	q := url.Values{}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	return q
}

func pathWithQuery(path string, opts ListOptions) string {
	q := opts.Query()
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

type Page[T any] struct {
	Items      []T
	Total      int
	NextOffset int
	HasMore    bool
}

func NewPage[T any](items []T, total, nextOffset int) Page[T] {
	return Page[T]{
		Items:      items,
		Total:      total,
		NextOffset: nextOffset,
		HasMore:    nextOffset >= 0,
	}
}

type PageFetcher[T any] func(context.Context, ListOptions) (Page[T], error)

func CollectAll[T any](ctx context.Context, fetch PageFetcher[T], pageSize int) ([]T, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	all := make([]T, 0, pageSize)
	offset := 0
	for {
		page, err := fetch(ctx, ListOptions{Offset: offset, Limit: pageSize})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if !page.HasMore {
			return all, nil
		}
		if page.NextOffset <= offset {
			return all, nil
		}
		offset = page.NextOffset
	}
}
