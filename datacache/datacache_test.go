package datacache

import (
	"appengine/aetest"
	"appengine/memcache"
	//"github.com/junpengxiao/Golb/database"
	"github.com/junpengxiao/Golb/post"
	"testing"
	"time"
)

func TestPutAndGet(t *testing.T) {
	ctx, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	p := post.Post{"Title", "Author", "Tag", "Snapshot", "Content", "Original", time.Now().Round(time.Second)}
	//check put, get and it should be catched
	if err := Put(&p, false, ctx); err != nil {
		t.Fatal(err)
	}
	if tmp, err := Get(p.Title, ctx); err != nil {
		t.Fatal(err)
	} else {
		p.Original = ""
		if p != *tmp {
			t.Errorf("Result doesn't match, original: ", p, "retrived item: ", *tmp)
		}
		//use memcache to check the item directly
		var tmp2 post.Post
		_, err2 := memcache.Gob.Get(ctx, p.Title, &tmp2)
		if err2 != nil || p != tmp2 {
			t.Errorf("Error in TestP&G Err: ", err2, " Retrieved Item :", tmp2)
		}
	}
	//check update, the cache should be expired
	if err := Put(&p, true, ctx); err != nil {
		t.Fatal(err)
	}
	var tmp2 post.Post
	_, err2 := memcache.Gob.Get(ctx, p.Title, &tmp2)
	if err2 != memcache.ErrCacheMiss {
		t.Errorf("Error in TestP&G, data not expired. Err: ", err2, "Retrived : ", tmp2)
	}
}

//check if 2 slice of post.Post are same
func check(arr1, arr2 []post.Post) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, v := range arr1 {
		if v != arr2[i] {
			return false
		}
	}
	return true
}

func TestQuery(t *testing.T) {
	ctx, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	p := post.Post{"T1", "Author", "Tag", "Snapshot", "Content", "Original", time.Now().Round(time.Second)}
	err = Put(&p, false, ctx)
	if err != nil {
		t.Fatal("Test Query, Error while put items")
	}
	p.Title = "T2"
	err = Put(&p, false, ctx)
	if err != nil {
		t.Fatal("Test Query, Error while put items")
	}
	p.Title = "T3"
	err = Put(&p, false, ctx)
	if err != nil {
		t.Fatal("Test Query, Error while put items")
	}
	p.Title = "T4"
	err = Put(&p, false, ctx)
	if err != nil {
		t.Fatal("Test Query, Error while put items")
	}
	p.Title = "T5"
	err = Put(&p, false, ctx)
	if err != nil {
		t.Fatal("Test Query, Error while put items")
	}

	//since query may not consistent, we need to use key retrieve item first before we query them.
	//Otherwise, I cannot get the new added item through query
	Get("T1", ctx)
	Get("T2", ctx)
	Get("T3", ctx)
	Get("T4", ctx)
	Get("T5", ctx)

	//query with data store
	ret1, next1, err1 := Query(1, 2, ctx)
	if err1 != nil {
		t.Fatal(err1)
	}
	//query from memcache
	ret2, next2, err2 := Query(1, 2, ctx)
	if err2 != nil {
		t.Fatal(err2)
	}
	if check(ret1, ret2) || next1 != next2 {
		t.Errorf("Test Query original and memcache aren't match, original: ", ret1, " memcache: ", ret2,
			" original next: ", next1, " memcache next: ", next2)
	}

	//query from cursor
	memcache.Delete(ctx, buildQueryKey(1, 2))
	ret3, next3, err3 := Query(1, 2, ctx)
	if err3 != nil {
		t.Fatal(err3)
	}
	if check(ret1, ret3) || next1 != next3 {
		t.Errorf("Test Query original and cursor aren't match, original: ", ret1, " cursor: ", ret3,
			"original next: ", next1, "cursor next: ", next3)
	}
}
