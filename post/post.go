package post

import (
	"time"
)

type Post struct {
	Title, Author, Tag, Snapshot, Content, Original string
	Date                                            time.Time
}
