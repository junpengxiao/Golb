//postprocessor process the new post from user and save it into database
//also it defines multimedias name in that post (like images, videos) and tell
//the http handler to require user to upload those files. It also defines the way
//to translate selfdefined language into HTML/JS/CSS
package postprocessor

import (
	"appengine"
	"bytes"
	"errors"
	"github.com/junpengxiao/Golb/datacache"
	"github.com/junpengxiao/Golb/post"
	"github.com/russross/blackfriday"
	"strings"
	"time"
)

//Errs
var ErrPostEmpty = errors.New("Original Post is empty")
var ErrPostMissBorE = errors.New("Original Post format missed /begin or /end or they are in missorder")

/*
	/title{Title}
	/author{Author}
	/tag{Tag}
	/begin

	/end
*/
const (
	keytitle  = "/title"
	keyauthor = "/author"
	keytag    = "/tag"
	keybegin  = "/begin"
	keyend    = "/end"
)

func extractValue(content, target string) string {
	index := strings.Index(content, target)
	if index != -1 {
		start := index + len(target)
		for ; start < len(content) && content[start] != '{'; start++ {
		} //find first {
		for ; start < len(content) && (content[start] == ' ' || content[start] == '\t'); start++ {
		} //escape empty character
		end := strings.index(content[start:], "}")
		for ; end-1 > start && (content[end-1] == ' ' || content[end-1] == '\t'); end-- {
		}
		if start < end {
			return content[start:end]
		}
	}
	return ""
}

//Process render the post original md content. if the originalName is provided,
//that means this post is updated from an older post. Process returns a slice of strings
//to represent which additional media is required to load into this function
func Process(data *post.Post, originalName string, ctx appengine.Context) ([]string, error) {
	if post.PostOriginal == "" {
		return nil, ErrPostEmpty
	}
	if originalName != "" {
		datacache.Delete(originalName, ctx)
	}
	//Build title,author,tag
	prologue := strings.Index(post.Post.Original, keybegin)
	if prologue == -1 {
		return nil, ErrPostMissBorE
	}
	epilogue := strings.LastIndex(post.Post.Original, keyend)
	if epilogue == -1 {
		epilogue = len(post.Post.Original)
	}
	post.Post.Date = time.Now().Round(time.Second)
	post.Post.Title = extractValue(post.Post.Original[:prologue], keytitle)
	if post.Post.Title == "" {
		post.Post.Title = post.Post.Date.Format(post.TimeLayout)
	}
	post.Post.Author = extractValue(post.Post.Original[:prologue], keyauthor)
	post.Post.Tag = extractValue(post.Post.Original[:prologue], keytag)

	//escape SciJax from markdown content
	markdown, sciChan, number := onepass(post.Post.Original)
	post.Post.Content = contentmerge(blackfriday.MarkdownCommon(markdown), sciConvert(sciChan, number))
	post.Post.Snapshot = formSnapshot(post.Post.Content)
}

const escapeMark = "$$"

func onepass(str string) ([]byte, chan string, int) {
	var markdown bytes.Buffer
	sciChan := make(chan string)
	num := 0
}
