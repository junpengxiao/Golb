package server

import (
	"appengine"
	"github.com/junpengxiao/Golb/database"
	"github.com/junpengxiao/Golb/datacache"
	"github.com/junpengxiao/Golb/post"
	"github.com/junpengxiao/Golb/postprocessor"
	"net/http"
	"strconv"
)

const (
	pagekey       = "page"
	titlekey      = "title"
	originalkey   = "content"
	itemsEachPage = 5
)

func init() {
	http.HandleFunc("/admin/newpost", newpost)       //show editor to add new post
	http.HandleFunc("/admin/uploadpost", uploadpost) //upload post
	http.HandleFunc("/home", home)                   //homepage
	http.HandleFunc("/post", post)                   //display one post
	http.HandleFunc("/", home)                       //default handler, homepage
}

func home(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	page, err := strconv.Atoi(r.FormValue(pagekey))
	if err != nil || page < 0 {
		page = 0
	}
	result, err := datacache.Query(page*itemsEachPage, itemsEachPage, ctx)
	if err != nil {
		w.Write(err)
	} else {
		w.Write(result)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	title := r.FormValue(titlekey)
	result, err := datacache.Get(title, ctx)
	if err != nil {
		w.Write(err)
	} else {
		w.Write(result)
	}
}

func uploadpost(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var data post.Post
	data.Original = r.FormValue(originalkey)
	if err := postprocessor.Process(&data); err != nil {
		w.Write(err)
		return
	}
	if err := datacache.Put(&data, false, ctx); err != nil {
		w.Write(err)
		return
	} else {
		http.Redirect(w, r, "/", http.StatusAccepted)
	}
}

const testpost = `
<html>
  <body>
    <form action="/admin/uploadpost" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Submit"></div>
    </form>
  </body>
</html>
`

func newpost(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Write(testpost)
}
