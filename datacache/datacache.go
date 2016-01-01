//datacache wraps database. It utilizes appengine/memcache to reduce the workload in database
package datacache

import (
	"appengine"
	"appengine/memcache"
	"github.com/junpengxiao/Golb/database"
	"github.com/junpengxiao/Golb/post"
	"log"
)

const mdsuffix = ".text"   // for markdown item. It isn't necessary to be same as variable in database
const itemsuffix = ".post" //again, no need to be same as corresponding varible in database

//wraps datastore.Put. If update change the snapshot of that item and it is really
//important to update that snapshot immediately, user should use a special
//link to flush the entire memcache. Nonetheless, new snapshot should be updated within 1 week
func Put(data *post.Post, isUpdate bool, ctx appengine.Context) error {
	if !isUpdate {
		return database.Put(data, isUpdate, ctx)
	} else {
		//no need to handle error
		memcache.Delete(ctx, data.Title)
		memcache.Delete(ctx, data.Title+mdsuffix)
		//no need to add into cache. It save into cache when Get func is called
		return database.Put(data, isUpdate, ctx)
	}
}

//wraps datastore.Get. If item exist in cache, return. Otherwise, retrieve from db and store in cache.
func Get(key string, ctx appengine.Context) (*post.Post, error) {
	ret := new(post.Post)
	if _, err := memcache.Gob.Get(ctx, key, ret); err != nil {
		ret, err = database.Get(key, ctx)
		if err != nil {
			return nil, err
		}
		item := &memcache.Item{
			Key:    key,
			Object: *ret,
		}
		if err := memcache.Gob.Set(ctx, item); err != nil {
			log.Println("Error in datacache, Get: ", err)
		}
		return ret, nil
	}
	return ret, nil
}

//wraps delete. Like Put, this function is lazy to snapshot. Blog owner need to flush the whole memcache
//if he or she need the corresponding snapshot be deleted immediately.
func Delete(title string, ctx appengine.Context) error {
	memcache.Delete(ctx, title)
	memcache.Deleta(ctx, title+itemsuffix)
	memcache.Delete(ctx, title+mdsuffix)
	database.Delete(title, ctx)
}
