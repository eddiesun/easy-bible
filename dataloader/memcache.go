package dataloader

import (
	"appengine"
	"appengine/memcache"
	"encoding/json"
	"strconv"
	"time"
	"util"
)

type ()

var bibleVersions = []string{"和合本"}
var numOfBooksInBible = 66
var liteBooksKey = "Lite Books"

func getMemcacheKey(version string, bookId int) string {
	return "bible-" + version + "-Book-" + strconv.Itoa(bookId)
}

func MemcachePut(c appengine.Context, bc *BibleCollection) error {
	defer util.LogTime(c, time.Now(), "    Saving Bible Collection to memcache: ")

	// prepare storing all books
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
	// prepare storing lite books
	liteBooksJson, err := json.Marshal(bc.LiteBooks)
	if err != nil {
		return err
	}
	memcacheItemList = append(memcacheItemList, &memcache.Item{
		Key:   liteBooksKey,
		Value: liteBooksJson,
	})

	// actual call to mamcache store
	err = memcache.SetMulti(c, memcacheItemList)
	if err != nil {
		return err
	}
	return nil
}

func MemcacheGet(c appengine.Context) (*BibleCollection, error) {
	defer util.LogTime(c, time.Now(), "    Getting bible collection from memcache ")

	// prepare all books
	var memcacheKeyList []string
	for _, version := range bibleVersions {
		for bookId := 1; bookId <= numOfBooksInBible; bookId++ {
			memcacheKeyList = append(memcacheKeyList, getMemcacheKey(version, bookId))
		}
	}

	// prepare lite books
	memcacheKeyList = append(memcacheKeyList, liteBooksKey)

	// fetching json data from memcache
	keyBookMap, err := memcache.GetMulti(c, memcacheKeyList)
	if err != nil || keyBookMap == nil || len(keyBookMap) <= 0 {
		return nil, err
	}

	// convert json into object for chapters and verses
	bc := new(BibleCollection)
	bc.Bibles = make(map[string]*Bible)
	for _, version := range bibleVersions {
		var bible Bible
		for bookId := 1; bookId <= numOfBooksInBible; bookId++ {
			key := getMemcacheKey(version, bookId)
			var book Book
			// c.Debugf("*** key: %+v\n *** keyBookMap[key]%+v\n", key, keyBookMap[key])
			if _, exist := keyBookMap[key]; !exist {
				// some books are not in memcache. Consider the whole book is not in cache, For Now!
				c.Debugf("***!!!!!!!*** NOT IN CACHE, key: %+v\n", key)
				return nil, err
			}
			err := json.Unmarshal(keyBookMap[key].Value, &book)
			if err != nil {
				return nil, err
			}
			bible.Books = append(bible.Books, book)
		}
		bc.Bibles[version] = &bible
	}

	// convert book name id map json to object
	if _, exist := keyBookMap[liteBooksKey]; !exist {
		c.Debugf("    NOT IN CACHE, key: %+v\n", liteBooksKey)
		return nil, err
	}
	err = json.Unmarshal(keyBookMap[liteBooksKey].Value, &bc.LiteBooks)
	if err != nil {
		return bc, nil
	}

	return bc, nil
}

func MemcacheGetBook(c appengine.Context, version string, bookId int) (*Book, error) {
	defer util.LogTime(c, time.Now(), "    Getting one Book from memcache ")

	item, err := memcache.Get(c, getMemcacheKey(version, bookId))
	if err == memcache.ErrCacheMiss {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var book Book
	if err := json.Unmarshal(item.Value, &book); err != nil {
		return nil, err
	}

	return &book, err
}

func MemcacheGetLiteBooks(c appengine.Context) ([]LiteBook, error) {
	defer util.LogTime(c, time.Now(), "    Getting LiteBooks from memcache ")
	item, err := memcache.Get(c, liteBooksKey)
	if err == memcache.ErrCacheMiss {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var liteBooks []LiteBook
	if err := json.Unmarshal(item.Value, &liteBooks); err != nil {
		return nil, err
	}

	return liteBooks, nil
}
