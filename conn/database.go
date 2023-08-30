package conn

import (
	"context"
	"os"

	"github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
)

func ConnectionDB() *kivik.Client {
	client, err := kivik.New("couch", os.Getenv("default_url"))
	if err != nil {
		panic(err)
	}
	client.Authenticate(context.TODO(), couchdb.BasicAuth(os.Getenv("CouchDB_Username"), os.Getenv("CouchDB_Password")))

	return client
}
