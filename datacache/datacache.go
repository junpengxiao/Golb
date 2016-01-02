//datacache wraps database. It utilizes appengine/memcache to reduce the workload in database
package datacache

import (
	"appengine"
	"appengine/memcache"
	"github.com/junpengxiao/Golb/database"
	"github.com/junpengxiao/Golb/post"
	"log"
	"strconv"
)

const mdsuffix = ".text"   // for markdown item. It isn't necessary to be same as variable in database
const itemsuffix = ".post" //again, no need to be same as corresponding varible in database

//Put wraps database.Put and return the value from database.Put. If update change the snapshot of that item and it is really
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

//Get wraps database.Get. It return the same value as database.Get(). Plus a record a log if memcache errors occur
//If item exists in cache, return. Otherwise, retrieve from db and stores in cache.
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
			log.Println("Error in datacache.Get(), while trying to set value", err)
		}
		return ret, nil
	}
	return ret, nil
}

//Delete wraps database.Delete. Like Put, this function is lazy to snapshot. return the same value as database.Delete
//Blog owner need to flush the whole memcache if he or she need the corresponding snapshot be deleted immediately.
func Delete(title string, ctx appengine.Context) error {
	memcache.Delete(ctx, title)
	memcache.Delete(ctx, title+itemsuffix)
	memcache.Delete(ctx, title+mdsuffix)
	return database.Delete(title, ctx)
}

func buildQueryKey(offset, limit int) string {
	return "qo" + strconv.Itoa(offset) + "l" + strconv.Itoa(limit)
}
func buildCursorKey(offset, limit int) string {
	return "co" + strconv.Itoa(offset) + "l" + strconv.Itoa(limit)
}

const haveNext string = "T"

func updateQueryCache(offset, limit int, ret []post.Post, next bool, cursor string, ctx appengine.Context) {
	if offset != 0 {
		//put cursor into memcache
		item := &memcache.Item{
			Key:    buildCursorKey(offset, limit),
			Object: cursor,
		}
		memcache.Gob.Set(ctx, item)
	}
	//put result into memcache
	if next {
		ret[0].Content = haveNext
	}
	item2 := &memcache.Item{
		Key:    buildQueryKey(offset, limit),
		Object: ret,
	}
	memcache.Gob.Set(ctx, item2)
	if next {
		ret[0].Content = ""
	}
}

//Query wraps database.Query. This function opaque the encodedcursor to upper level.
//Since Query only returns post.Post with PostItem part, I use post.Post.Content to store haveNext boolean value
//Notice that the cursor string returns from offset with 0 will not be stored. Since it can not reduce the work load

func Query(offset, limit int, ctx appengine.Context) ([]post.Post, bool, error) {
	//check offset and limit
	if offset < 0 || limit <= 0 {
		return nil, false, nil
	}
	//Search from cache if current Query is catched
	ret := make([]post.Post, 0, limit)
	if _, err := memcache.Gob.Get(ctx, buildQueryKey(offset, limit), &ret); err == nil {
		//retrieve the data successfully.
		if len(ret) == 0 {
			return nil, false, nil
		} else {
			//check if there is more items stored in DB. This is used to determin whether "next" button should be displayed on webpage
			log.Println("Datacache Query,  Query with memcache hit No.2 checked (3 checks in total)")
			if ret[0].Content == haveNext {
				ret[0].Content = ""
				return ret, true, nil
			}
			return ret, false, nil
		}
	} else {
		//fail to retrieve the data from memcache.
		//Try to retrive cursor from memcache to avoid Query Offset workload
		if offset != 0 {
			var cursor string
			if _, err := memcache.Gob.Get(ctx, buildCursorKey(offset, limit), &cursor); err == nil {
				//retrive the cursor successfully.
				if ret, next, newcursor, err := database.Query(offset, limit, cursor, ctx); err != nil {
					return nil, false, err
				} else {
					//update cursor value in case it was out of date because of new posts added.
					log.Println("Datacache Query,  Query with cursor hit No.3 checked (3 checks in total)")
					updateQueryCache(offset, limit, ret, next, newcursor, ctx)
					return ret, next, nil
				}
			}
		}
		//offset == 0 or fail to retrieve cursor, then query database directly
		if ret, next, newcursor, err := database.Query(offset, limit, "", ctx); err != nil {
			return nil, false, err
		} else {
			log.Println("Datacache Query, Query with datastore hit, No.1 checked (3 checks in total)")
			updateQueryCache(offset, limit, ret, next, newcursor, ctx)
			return ret, next, nil
		}
	}
	//all if situation above will return directly, so this line will never be arrived
}
