package test_utils

import (
	"context"
	"testing"

	"google.golang.org/appengine/datastore"
)

func ClearDatastore(t *testing.T, ctx context.Context, kind string) {
	q := datastore.NewQuery(kind).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = datastore.DeleteMulti(ctx, keys); err != nil {
		t.Fatal(err)
	}
}

func ExpectChange(t *testing.T, ctx context.Context, kind string, diff int, f func()) {
	q0 := datastore.NewQuery(kind)
	c0, err := q0.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	f()
	q1 := datastore.NewQuery(kind)
	c1, err := q1.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if c1-c0 != diff {
		t.Fatalf("Expect diff is %v, but it changed from %v to %v in %v", diff, c0, c1, kind)
	}
}
