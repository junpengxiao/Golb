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
var (
	ErrPostEmpty = errors.New("Original Post is empty")
	ErrPostMissBorE = errors.New("Original Post format missed /begin or /end or they are in missorder")
	ErrPostSciMarkNotMatch = errors.New("Original Post contants single $$, which is a SciMark that need to be paired")
)
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
const escapeMarkB = []byte(escapeMark)

//onepass split original content into 2 parts, one for markdown, the other one for sci handler
func onepass(str string) ([]byte, chan string, int, error) {
	var markdown bytes.Buffer
	sciChan := make(chan string)
	num, last, now := 0, 0, 0
	for ;now = strings.Index(str[last:], escapeMark); now != -1 {
		markdown.WriteString(str[last:now+len(escapeMark)])
		last = strings.Index(str[now+len(escapeMark):],escapeMark)
		if last == -1 {
			return nil, nil, 0, ErrPostSciMarkNotMatch
		}
		sciChan <- str[now+len(escapeMark):last]
		num++
		last+=len(escapeMark)
	}
	return markdown.Bytes(), sciChan, num, nil
}

func sciConvert(sciChan chan string, num int) {
	ret := make([]string, num)
	for i:=0; i!=num; i++ {
		go func(result *string, )
	}
}

//contentmerge merge the result from markdown and sci string together
func contentmerge(markdown []byte, sciStr []string) string {
	var ret bytes.Buffer
	last, now := 0, 0
	for index := 0; now = bytes.Index(markdown[last:], escapeMarkB); now != -1 {
		ret.WriteByte(markdown[last:now])
		ret.WriteString(sciStr[index++])
		last = now + len(escapeMarkB)
	}
	ret.WriteByte(markdown[last:])
	return ret.String()
} 

//snapshot extract a snapshot for the html content
func formSnapshot(str string) string{
	var ret bytes.Buffer
	head := regexp.MustCompile( `<h[1-6]>` )
	headIndex := head.FindStringIndex(str)
	bodyIndex := strings.Index(str, `<p>`)
	if headIndex == nil || (bodyIndex!=-1 && headIndex[0]>bodyIndex) {
		bodyend := strings.Index(str[bodyIndex:], `<\p>`)
		if headIndex == nil {
			return str[bodyIndex:bodyend+len(`<\p>`)]
		} else {
			ret.WriteString(str[bodyIndex:bodyend+len(`<\p>`)])
			ret.WriteRune('\n')
		}
	}
	//add head into snapshot
	headmark := `<\` + str[headIndex[0]+1:headIndex[1]]
	headend := strings.Index(str, headmark)
	ret.WriteString(str[headIndex[0]:headend+len(headmark)])
	//add <p> behind that head into snapshot
	bodyIndex = strings.Index(str[headend:],`<p>`)
	if bodyIndex == -1 {
		return ret.String()
	}
	ret.WriteRune('\n')
	bodyend := strings.Index(str[bodyIndex:], `<\p>`)
	ret.WriteString(str[bodyIndex:bodyend+len(`<\p>`)])
	return ret.String()
}
