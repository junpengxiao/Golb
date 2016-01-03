//use google cloud storage to store media files like image
//since go cloud storage package is in experimental and may make backwards-incompatible changes,
//this  package should be updated correspondingly
package datastorage

import (
	"appengine"
	"errors"
	"google.golang.org/cloud/storage"
)

//used in PutMulti to enforce that number of keys should equal with number of data
var ErrNumberNotMatch = errors.New("#key isn't same as #data")

//bucket name, TODO: read from config file instead of hardwire value here
const bucket string = "golbucket"

//Put stores data into bucket under with filrname "key". if returns a url that can be used to read this data
//what if data already exist?
func Put(key string, data []byte, ctx appengine.Context) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	writer := client.Bucket(bucket).Object(key).NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return writer.MediaLink, nil
}

//wraped Put result, in order to keep the order of returned result in PutMulti function
type orderResult struct {
	order int
	url   string
	err   error
}

//PutMulti stores multiple item in cloud storage in one stroke. It use goroutine to do multitasking
//It will return error[0] with ErrNumberNotMatch if the #key isn't same as #datas
//It will return a slice of strings as url if that item is stored successfully.
//But if there is an error to store that item, then an error will be put in corresponding space.
func PutMulti(key []string, data [][]byte, ctx appengine.Context) ([]string, []error) {
	if len(key) != len(data) {
		err := make([]error)
		err = append(err, ErrNumberNotMatch)
		return nil, err
	}
	message := make(chan orderResult)
	for i := range key {
		go func(key string, data []byte, order int) {
			var result orderResult
			result.order = order
			result.url, result.err = Put(key, data, ctx)
			message <- result
		}(key[i], data[i], i)
	}
	urls := make([]string, len(key))
	errs := make([]error, len(key))
	for i := 0; i != len(key); i++ {
		tmp := <-message
		urls[tmp.order] = tmp.url
		errs[tmp.order] = tmp.err
	}
	return urls, errs
}
