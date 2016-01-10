package template

import (
	"fmt"
	"github.com/junpengxiao/Golb/post"
	"log"
	"net/http"
	"os"
	"text/template"
)

const (
	navNone    = -1
	navHome    = 0
	navProject = 1
	navResume  = 2
	haveNext   = "true"
)

var templ *template.Template

func init() {
	var err error
	log.Println(os.Getwd())
	templ, err = template.ParseFiles("stylesheets/template.html")
	if err != nil {
		log.Println(err)
	}
}

type Content struct {
	NavNum int //highlight the NavNum nav-button. -1 none, 0 home, 1 project, 2 resume,
	Posts  []post.Post
}

func PostList(list []post.Post, next bool, w http.ResponseWriter) {
	var data Content
	data.NavNum = navHome
	data.Posts = list
	if next && len(data.Posts) > 0 {
		data.Posts[0].Content = haveNext
	}
	log.Println("----> Template Debug", data)
	if err := templ.ExecuteTemplate(w, "HomePage", data); err != nil {
		fmt.Fprintln(w, err)
	}
}

func DisplayPost(content *post.Post, w http.ResponseWriter) {
	var data Content
	data.NavNum = navNone
	data.Posts = append(data.Posts, *content)
	if err := templ.ExecuteTemplate(w, "Article", data); err != nil {
		fmt.Fprintln(w, err)
	}
}
