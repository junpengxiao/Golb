//database encapsulate the appengine/datastore to provide support to Golb
package database

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"github.com/junpengxiao/Golb/post"
	//"log"
	"strings"
	"time"
)

type PostItem struct {
	Title, Author, Tag, Snapshot string `datastore:",noindex"`
	Date                         time.Time
}

type PostContent struct {
	Content string `datastore:",noindex"`
}

var ErrPostExists = errors.New("Post with this title exists in datastore")
var ErrPostNotExists = errors.New("Post with this title doesn't exist in datastore")

const itemsuffix string = ".post"
const mdsuffix string = ".text"
const postkind string = "Post"
const contentkind string = "Content"

//postkey return 3 keys used to store one post:
//itemkey: "title.post" used in PostItem.
//contentkey: "title" used in full article content
//contentmdkey: "title.text" used in content markdown format
func postkey(title string, ctx appengine.Context) (*datastore.Key, *datastore.Key, *datastore.Key) {
	itemkey := datastore.NewKey(ctx, postkind, title+itemsuffix, 0, nil)
	contentkey := datastore.NewKey(ctx, contentkind, title, 0, itemkey)
	contentmdkey := datastore.NewKey(ctx, contentkind, title+mdsuffix, 0, itemkey)
	return itemkey, contentkey, contentmdkey
}

//convert a post into 3 parts stored in datastore
func convertPost(data *post.Post) (PostItem, PostContent, PostContent) {
	item := PostItem{data.Title, data.Author, data.Tag, data.Snapshot, data.Date}
	content := PostContent{data.Content}
	contentmd := PostContent{data.Original}
	return item, content, contentmd
}

//build a post from 3 parts stored in datastore
func buildPost(item *PostItem, content, contentmd *PostContent) *post.Post {
	data := post.Post{item.Title, item.Author, item.Tag, item.Snapshot,
		content.Content, contentmd.Content, item.Date}
	return &data
}

//Put stored a post into datastore. It split original data into 3 parts:
//PostItem, Content with HTML, Original Markdown Content. Then store them in a transaction
//if isUpdate is true, then it is a update to original post. otherwise it is a new post.
func Put(data *post.Post, isUpdate bool, ctx appengine.Context) error {
	itemkey, contentkey, contentmdkey := postkey(data.Title, ctx)
	//first check whether post exists
	tmp := new(PostItem)
	if isUpdate {
		if err := datastore.Get(ctx, itemkey, tmp); err != nil {
			return ErrPostNotExists
		}
	} else {
		if err := datastore.Get(ctx, itemkey, tmp); err != datastore.ErrNoSuchEntity {
			return ErrPostExists
		}
	}
	//convert data into 3 parts
	item, content, contentmd := convertPost(data)
	//store 3 parts into datastore in transaction
	err := datastore.RunInTransaction(ctx, func(ctx appengine.Context) error {
		if _, err := datastore.Put(ctx, itemkey, &item); err != nil {
			return err
		}
		if _, err := datastore.Put(ctx, contentkey, &content); err != nil {
			return err
		}
		if _, err := datastore.Put(ctx, contentmdkey, &contentmd); err != nil {
			return err
		}
		return nil
	}, nil)
	return err
}

//Get returns a post from datastore. If the key is item key, it returns item part
//if the key is content key, it returns item along with content part
//if the key is contentmd key, it returns item along with original part
func Get(key string, ctx appengine.Context) (*post.Post, error) {
	item := new(PostItem)
	content := new(PostContent)
	contentmd := new(PostContent)
	//return markdown format, i.e., original
	if strings.HasSuffix(key, mdsuffix) {
		itemkey, _, contentmdkey := postkey(key[:len(key)-len(mdsuffix)], ctx)
		if err := datastore.Get(ctx, itemkey, item); err != nil {
			return nil, err
		}
		if err := datastore.Get(ctx, contentmdkey, contentmd); err != nil {
			return nil, err
		}
		return buildPost(item, content, contentmd), nil
	}

	if strings.HasSuffix(key, itemsuffix) {
		itemkey, _, _ := postkey(key[:len(key)-len(itemsuffix)], ctx)
		if err := datastore.Get(ctx, itemkey, item); err != nil {
			return nil, err
		}
		return buildPost(item, content, contentmd), nil
	}

	itemkey, contentkey, _ := postkey(key, ctx)
	if err := datastore.Get(ctx, itemkey, item); err != nil {
		return nil, err
	}
	if err := datastore.Get(ctx, contentkey, content); err != nil {
		return nil, err
	}
	return buildPost(item, content, contentmd), nil
}

func Delete(title string, ctx appengine.Context) error {
	itemkey, contentkey, contentmdkey := postkey(title, ctx)
	return datastore.RunInTransaction(ctx, func(ctx appengine.Context) error {
		if err := datastore.Delete(ctx, contentkey); err != nil {
			return err
		}
		if err := datastore.Delete(ctx, contentmdkey); err != nil {
			return err
		}
		if err := datastore.Delete(ctx, itemkey); err != nil {
			return err
		}
		return nil
	}, nil)
}

//Query first jumps 'offset' posts, then return 'limit' posts, along with a boolean value to denote
//if database has more posts, a string to denote cursor for that query. Based on GAE document,
//offset has its own cost, so App should store this value, and use it for the same query next time.
func Query(offset, limit int, encodedCursor string, ctx appengine.Context) ([]post.Post, bool, string, error) {
	if limit <= 0 {
		return nil, false, "", nil
	}
	//limit = limit + 1. This is used for determin whether we need to display "next" button in our webpage
	limit++
	query := datastore.NewQuery(postkind).Limit(limit).Order("-Date")
	//if encodedCursor is provided, then try to use it instead of offset
	if encodedCursor != "" {
		cursor, err := datastore.DecodeCursor(encodedCursor)
		if err == nil {
			query = query.Start(cursor)
		}
	} else {
		query = query.Offset(offset)
	}
	it := query.Run(ctx)
	cursor, err := it.Cursor()
	if err != nil {
		return nil, false, "", err
	}
	startPosition := cursor.String()
	//build return value
	haveNext := true
	ret := make([]post.Post, 0, limit)
	for i := 0; i != limit; i++ {
		var item PostItem
		_, err := it.Next(&item)
		if err == datastore.Done {
			haveNext = false
			break
		}
		if err != nil {
			return nil, false, "", err
		}
		ret = append(ret, post.Post{item.Title, item.Author, item.Tag, item.Snapshot, "", "", item.Date})
	}
	if haveNext {
		return ret[:limit-1], haveNext, startPosition, nil
	} else {
		return ret, haveNext, startPosition, nil
	}
}
