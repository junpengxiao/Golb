package database

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"github.com/junpengxiao/Golb/post"
)

type PostItem struct {
	Title, Author, Tag, Snapshot string `datastore:",noindex"`
	Date                         time.Time
}

type PostContent struct {
	content string `datastore:".noindex"`
}

var ErrPostExists = errors.New("Post with this title exists in datastore")

//postkey return 3 keys used to store one post:
//itemkey: "title.post" used in PostItem.
//contentkey: "title" used in full article content
//contentmdkey: "title.text" used in content markdown format
func postkey(title string, ctx appengine.Context) (*datastore.Key, *datastore.Key, *datastore.Key) {
	itemkey := datastore.NewKey(ctx, "Post", title+".post", 0, nil)
	contentkey := datastore.NewKey(ctx, "Content", title, 0, itemkey)
	contentmdkey := datastore.NewKey(ctx, "Content", title+".text", 0, itemkey)
	return itemkey, contentkey, contentmdkey
}

//convert a post into 3 parts stored in datastore
func convertPost(data *post.Post) (PostItem, PostContent, PostContent) {
	item := PostItem{data.Title, data.Author, data.Tag, data.Snapshot, data.Date}
	content := PostContent{data.Content}
	contentmd := PostContent{data.Original}
	return item, content, contentmd
}

//StoreData stored a post into datastore and return the content key. It split original data into 3 parts:
//PostItem, Content with HTML, Original Markdown Content. Then store them in a transaction
func StoreData(data *post.Post, ctx appengine.Context) (*datastore.Key, error) {
	//first check whether post exists
	itemkey, contentkey, contentmdkey := postkey(data.Title, ctx)
	tmp = new(PostItem)
	if err := datastore.Get(ctx, itemkey, tmp); err != datastore.ErrNoSuchEntity {
		return "", ErrPostExists
	}
	//convert data into 3 parts
	item, content, contentmd := convertPost(data)
	//store 3 parts into datastore in transaction
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := datastore.Put(ctx, itemkey, &item)
		if err == nil {
			datastore.Put(ctx, contentkey, &content)
		}
		if err == nil {
			datastore.Put(ctx, contentmdkey, &contentmd)
		}
		return err
	}, nil)
	if err != nil {
		return nil, err
	}
	return contentkey, nil
}
