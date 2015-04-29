package dataloader

import (
	"appengine"
	"appengine/memcache"
	"encoding/json"
	"strconv"
	"time"
)

type ()

var bibleVersions = []string{"和合本"}
var numOfBooksInBible = 66

func getMemcacheKey(version string, bookId int) string {
	return "bible-" + version + "-Book-" + strconv.Itoa(bookId)
}

func MemcachePut(c appengine.Context, bc *BibleCollection) error {
	timeStart := time.Now()
	defer c.Infof("    Saving Bible Collection to memcache: %s \n", time.Since(timeStart))

	var memcacheItemList []*memcache.Item
	for _, bible := range bc.Bibles {
		for _, book := range bible.Books {
			t, err := json.Marshal(book)
			if err != nil {
				return err
			}
			memcacheItemList = append(memcacheItemList, &memcache.Item{
				Key:   getMemcacheKey(bible.Version, book.Id),
				Value: t,
			})
		}
	}
	err := memcache.SetMulti(c, memcacheItemList)
	if err != nil {
		return err
	}
	return nil
}

func MemcacheGet(c appengine.Context) (*BibleCollection, error) {
	timeStart := time.Now()
	defer c.Infof("    Getting bible collection from memcache %s \n", time.Since(timeStart))

	var memcacheKeyList []string
	for _, version := range bibleVersions {
		for bookId := 1; bookId <= numOfBooksInBible; bookId++ {
			memcacheKeyList = append(memcacheKeyList, getMemcacheKey(version, bookId))
		}
	}
	keyBookMap, err := memcache.GetMulti(c, memcacheKeyList)
	if err != nil || keyBookMap == nil || len(keyBookMap) <= 0 {
		return nil, err
	}

	bc := new(BibleCollection)
	bc.Bibles = make(map[string]*Bible)
	for _, version := range bibleVersions {
		var bible Bible
		for bookId := 1; bookId <= numOfBooksInBible; bookId++ {
			key := getMemcacheKey(version, bookId)
			var book Book
			err := json.Unmarshal(keyBookMap[key].Value, &book)
			if err != nil {
				return nil, err
			}
			// c.Debugf("*** %+v\n%+v\n%+v\n", bc, version, book)
			bible.Books = append(bible.Books, book)
		}
		bc.Bibles[version] = &bible
	}

	return bc, nil
}
