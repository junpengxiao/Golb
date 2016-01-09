package postprocessor

import (
	"fmt"
	"github.com/junpengxiao/Golb/post"
	"testing"
)

var standardPost = post.Post{
	Original: `\title{TestContent}
\author{Junpeng Xiao}
\begin

这里是序言

# This is the first head 第一章

Hello World $$x_2^5$$
\end
`,
}

func TestProcess(t *testing.T) {
	if err := Process(&standardPost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Title: ", standardPost.Title)
	fmt.Println("Author: ", standardPost.Author)
	fmt.Println("Tag: ", standardPost.Tag)
	fmt.Println("Date: ", standardPost.Date.Format(post.TimeLayout))
	fmt.Println("Snapshot: ", standardPost.Snapshot)
	fmt.Println("Content: ", standardPost.Content)
	fmt.Println("Original: ", standardPost.Original)
}
