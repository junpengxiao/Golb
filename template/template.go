package template

import (
	"github.com/junpengxiao/Golb/post"
	"net/http"
	"text/template"
)

const (
	navNone    = -1
	navHome    = 0
	navProject = 1
	navResume  = 2
	haveNext   = "true"
)

var templ, _ = template.ParseFiles("template.html")

type Content struct {
	NavNum int //highlight the NavNum nav-button. -1 none, 0 home, 1 project, 2 resume,
	Posts  []post.Post
}

func PostList(list []post.Post, next bool, w http.ResponseWriter) error {
	var data Content
	data.NavNum = navHome
	data.Posts = list
	if next && len(data.Posts) > 0 {
		data.Posts[0].Content = haveNext
	}
	templ.ExecuteTemplate(w, "HomePage", data)
}

func CertainPost(list post.Post, w http.ResponseWriter) error {
	var data Content
	data.NavNum = navNone
	data.Posts = list
	templ.ExecuteTemplate(w, "Article", data)
}
