package database

import (
	"appengine/aetest"
	//"appengine/datastore"
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
	if err := Put(&p, false, ctx); err != nil {
		t.Fatal(err)
	}
	p.Content = "NewContent"
	if err := Put(&p, true, ctx); err != nil {
		t.Fatal(err)
	}
	if tmp, err := Get(p.Title+mdsuffix, ctx); err != nil {
		t.Fatal(err)
	} else {
		p.Content = ""
		if p != *tmp {
			t.Errorf("Result doesn't match", *tmp)
		}
		p.Content = "NewContent"
	}
	if tmp, err := Get(p.Title, ctx); err != nil {
		t.Fatal(err)
	} else {
		p.Original = ""
		if p != *tmp {
			t.Errorf("Result doesn't match", *tmp)
		}
		p.Original = "Original"
	}
}

func TestDelete(t *testing.T) {
	ctx, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()
	p := post.Post{"Title", "Author", "Tag", "Snapshot", "Content", "Original", time.Now().Round(time.Second)}
	Put(&p, false, ctx)
	if err := Delete("Title", ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := Get("Title", ctx); err == nil {
		t.Fatal("Entity should be deleted")
	} else {
		t.Log(err)
	}
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

	//since query may not consistent, we need to use key retrieve item first before we query them.
	//Otherwise, I cannot get the new added item through query
	Get("T1", ctx)
	Get("T2", ctx)
	Get("T3", ctx)
	if posts, _, str, err := Query(1, 10, "", ctx); err != nil {
		t.Fatal(err)
	} else {
		if len(posts) == 0 {
			t.Fatal("Test Query, the result is empty")
		}
		p.Content = ""
		p.Original = ""
		if posts[len(posts)-1] != p {
			t.Errorf("Test Query, the result isn't match, expect: ", p, " Get: ", posts[len(posts)-1])
		}

		if post2, _, _, err := Query(1, 10, str, ctx); err != nil {
			t.Fatal(err)
		} else {
			if len(post2) == 0 {
				t.Fatal("Test Query, the 2nd result is empty")
			}
			if post2[len(post2)-1] != p {
				t.Errorf("Test Query, the result isn't match, expect: ", p, " Get: ", post2[len(post2)-1])
			}
		}
	}
}
