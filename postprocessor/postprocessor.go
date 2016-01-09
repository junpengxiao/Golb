//postprocessor process the new post from user and save it into database
//also it defines multimedias name in that post (like images, videos) and tell
//the http handler to require user to upload those files. It also defines the way
//to translate selfdefined language into HTML/JS/CSS
package postprocessor

import (
	"bytes"
	"errors"
	"github.com/junpengxiao/Golb/post"
	"github.com/russross/blackfriday"
	//"log"
	"regexp"
	"strings"
	"time"
)

//Errs
var (
	ErrPostEmpty           = errors.New("Original Post is empty")
	ErrPostMissBorE        = errors.New("Original Post format missed /begin or /end or they are in missorder")
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
	keytitle      = `\title`
	keyauthor     = `\author`
	keytag        = `\tag`
	keybegin      = `\begin`
	keyend        = `\end`
	defaultTitle  = ""
	defaultAuthor = "Junpeng Xiao"
	defaultTag    = ""
)

func extractValue(content, target string) string {
	index := strings.Index(content, target)
	if index != -1 {
		start := index + len(target)
		for ; start < len(content) && content[start] != '{'; start++ {
		} //find first {
		start++
		for ; start < len(content) && (content[start] == ' ' || content[start] == '\t'); start++ {
		} //escape empty character
		end := strings.Index(content[start:], "}")
		end += start
		for ; end-1 > start && (content[end-1] == ' ' || content[end-1] == '\t'); end-- {
		}
		if start < end {
			return content[start:end]
		}
	}
	return ""
}

//Process render the post original md content. Process may returns erorr
func Process(data *post.Post) error {
	if data.Original == "" {
		return ErrPostEmpty
	}
	//Build title,author,tag TODO: define prologue in comment
	prologue := strings.Index(data.Original, keybegin)
	if prologue == -1 {
		return ErrPostMissBorE
	}
	epilogue := strings.LastIndex(data.Original, keyend)
	if epilogue == -1 {
		epilogue = len(data.Original)
	}
	data.Date = time.Now().Round(time.Second)
	//default title is the post time
	if data.Title = extractValue(data.Original[:prologue], keytitle); data.Title == "" {
		data.Title = defaultTitle + data.Date.Format(post.TimeLayout)
	}
	if data.Author = extractValue(data.Original[:prologue], keyauthor); data.Author == "" {
		data.Author = defaultAuthor
	}
	if data.Tag = extractValue(data.Original[:prologue], keytag); data.Tag == "" {
		data.Tag = defaultTag
	}

	//escape SciJax from markdown content
	markdown, sciSlice, err := onepass(data.Original[prologue+len(keybegin) : epilogue])
	if err != nil {
		return err
	}
	data.Content = contentmerge(blackfriday.MarkdownCommon(markdown), sciConvert(sciSlice))
	data.Snapshot = formSnapshot(data.Content)

	return nil
}

//escapeMark is used to mark sci body in original markdown content
const escapeMark = "$$"

//same as escapeMark but with byte format
var escapeMarkB = []byte(escapeMark)

//onepass split original content into 2 parts, one for markdown, the other one for sci handler
//sci handler is a format that I defined for scientitic ussage like graph, math, codes, etc.
//currently math only
func onepass(str string) ([]byte, []string, error) {
	var markdown bytes.Buffer
	sciSlice := make([]string, 0, 1024) //1024 is just an estimate about how many sci part may exists
	last, now := 0, 0
	for now != -1 {
		if now = strings.Index(str[last:], escapeMark); now == -1 {
			break
		}
		now += last
		//write "content$$" into buffer. $$ is the mark that a converted sci content need to be inserted
		markdown.WriteString(str[last : now+len(escapeMark)])
		if last = strings.Index(str[now+len(escapeMark):], escapeMark); last == -1 {
			return nil, nil, ErrPostSciMarkNotMatch
		}
		last += (now + len(escapeMark))
		sciSlice = append(sciSlice, str[now+len(escapeMark):last])
		last += len(escapeMark)
	}
	return markdown.Bytes(), sciSlice, nil
}

//used for maintain order in concurrent processing
type sciConcurrent struct {
	content string
	index   int
}

//how to convert should be defined later. Currently it only support mathjax
func sciConrrentConvert(str string) string {
	return "$$" + str + "$$"
}

//convert sci body concurrently
func sciConvert(sciSlice []string) []string {
	message := make(chan sciConcurrent)
	for i := 0; i != len(sciSlice); i++ {
		go func(str string, order int) {
			var tmp sciConcurrent
			tmp.content = sciConrrentConvert(str)
			tmp.index = order
			message <- tmp
		}(sciSlice[i], i)
	}
	ret := make([]string, len(sciSlice))
	for i := 0; i != len(sciSlice); i++ {
		tmp := <-message
		ret[tmp.index] = tmp.content
	}
	return ret
}

//contentmerge merge the result from markdown and sci string together
func contentmerge(markdown []byte, sciStr []string) string {
	var ret bytes.Buffer
	last, now, index := 0, 0, 0
	for now != -1 {
		if now = bytes.Index(markdown[last:], escapeMarkB); now == -1 {
			break
		}
		now += last
		ret.Write(markdown[last:now])
		ret.WriteString(sciStr[index])
		last = now + len(escapeMarkB)
		index++
	}
	ret.Write(markdown[last:])
	return ret.String()
}

//snapshot extract a snapshot for the html content
func formSnapshot(str string) string {
	var ret bytes.Buffer
	head := regexp.MustCompile(`<h[1-6]>`)
	headIndex := head.FindStringIndex(str)
	bodyIndex := strings.Index(str, `<p>`)
	if headIndex == nil || (bodyIndex != -1 && headIndex[0] > bodyIndex) {
		bodyend := strings.Index(str[bodyIndex:], `</p>`) + bodyIndex
		if headIndex == nil {
			return str[bodyIndex : bodyend+len(`</p>`)]
		} else {
			ret.WriteString(str[bodyIndex : bodyend+len(`</p>`)])
			ret.WriteRune('\n')
		}
	}
	//add head into snapshot
	headmark := `</` + str[headIndex[0]+1:headIndex[1]]
	headend := strings.Index(str, headmark)
	ret.WriteString(str[headIndex[0] : headend+len(headmark)])
	ret.WriteRune('\n')
	//add <p> behind that head into snapshot
	bodyIndex = strings.Index(str[headend:], `<p>`) + headend
	if bodyIndex == -1 {
		return ret.String()
	}
	bodyend := strings.Index(str[bodyIndex:], `</p>`) + bodyIndex
	ret.WriteString(str[bodyIndex : bodyend+len(`</p>`)])
	return ret.String()
}
