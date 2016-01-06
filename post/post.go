package post

import (
	"time"
)

type Post struct {
	Title, Author, Tag, Snapshot, Content, Original string
	Date                                            time.Time
}

const TimeLayout string = "Jan 2 15:04:05 PST 2006"
