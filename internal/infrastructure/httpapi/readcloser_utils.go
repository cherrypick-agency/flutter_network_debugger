package httpapi

import (
	"bytes"
	"io"
)

type readCloser struct {
	io.Reader
	io.Closer
}

type closeFunc func() error

func (c closeFunc) Close() error { return c() }

func multiCloser(closers ...io.Closer) io.Closer {
	return closeFunc(func() error {
		var firstErr error
		for _, cl := range closers {
			if cl == nil {
				continue
			}
			if err := cl.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	})
}

func bodyWithCloser(body io.ReadCloser, r io.Reader) io.ReadCloser {
	if body == nil {
		return io.NopCloser(r)
	}
	return &readCloser{
		Reader: r,
		Closer: body,
	}
}

func prependBody(prefix []byte, body io.ReadCloser) io.ReadCloser {
	if body == nil {
		return io.NopCloser(bytes.NewReader(prefix))
	}
	if len(prefix) == 0 {
		return body
	}
	return &readCloser{
		Reader: io.MultiReader(bytes.NewReader(prefix), body),
		Closer: body,
	}
}
