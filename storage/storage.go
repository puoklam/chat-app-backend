package storage

import (
	"context"
	"log"
	"time"

	st "cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

const bucketName = "instant-messenger-ab5e3.appspot.com"

var app *firebase.App
var client *storage.Client
var bucket *st.BucketHandle

func init() {
	cfg := &firebase.Config{
		StorageBucket: bucketName,
	}
	opt := option.WithCredentialsFile("data/firebase-cred.json")
	var err error
	app, err = firebase.NewApp(context.Background(), cfg, opt)
	if err != nil {
		log.Fatalln(err)
	}
	client, err = app.Storage(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	bucket, err = client.DefaultBucket()
	if err != nil {
		log.Fatalln(err, "okok")
	}
}

func Upload(ctx context.Context, data []byte, pre string) (string, error) {
	name, err := Put(ctx, data, pre)
	if err != nil {
		return "", err
	}
	if err := public(ctx, name); err != nil {
		return "", err
	}
	return signed(name)
	// return Get(ctx, name)
}

func Put(ctx context.Context, data []byte, pre string) (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	s := pre + uuid.String()
	w := bucket.Object(s).NewWriter(ctx)
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return "", err
	}
	return s, err
}

func Get(ctx context.Context, name string) (*st.ObjectAttrs, error) {
	return bucket.Object(name).Attrs(ctx)
}

func public(ctx context.Context, name string) error {
	o := bucket.Object(name)
	return o.ACL().Set(ctx, st.AllUsers, st.RoleReader)
}

func signed(name string) (string, error) {
	return bucket.SignedURL(name, &st.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(24 * time.Hour),
	})
}
