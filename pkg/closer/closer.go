package closer

import "io"

func MustClose(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}
