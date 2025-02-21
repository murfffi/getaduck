package sclerr

import "io"

func CloseQuietly(c io.Closer) {
	_ = c.Close()
}
