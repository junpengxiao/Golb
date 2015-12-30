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
	if tmp, err := Get(p.Title, ctx); err != nil {
		t.Fatal(err)
	} else {
		p.Original = ""
		if p != *tmp {
			t.Errorf("Result doesn't match", *tmp)
		}
		p.Original = "Original"
	}

	if tmp, err := Get(p.Title+mdsuffix, ctx); err != nil {
		t.Fatal(err)
	} else {
		p.Content = ""
		if p != *tmp {
			t.Errorf("Result doesn't match", *tmp)
		}
	}
}
