package dataloader

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	// "fmt"
	"appengine"
	"appengine/datastore"
	"net/http"
)

type (
	BibleCollection struct {
		Bibles map[string]Bible
	}

	Bible struct {
		Version string `xml:"version,attr"`
		Books   []Book `xml:"book"`
	}

	Book struct {
		Id        int       `xml:"id,attr"`
		ShortName string    `xml:"shortName,attr"`
		LongName  string    `xml:"name,attr"`
		OtherName string    `xml:"pinyin,attr"`
		Chapters  []Chapter `xml:"chapter"`
	}

	Chapter struct {
		Number int     `xml:"number,attr"`
		Verses []Verse `xml:"verse"`
	}

	Verse struct {
		Number int    `xml:"number,attr"`
		Text   string `xml:",innerxml"`
	}

	PersistedChapter struct {
		BibleVersion  string
		BookId        int
		BookShortName string
		BookLongName  string
		BookOtherName string
		ChapterNumber int
		Verses        []Verse
	}

	QueryOptions struct {
		BookId        int
		BookShortName string
		ChapterNumber int
		OrderBy       string
		Project       []string
		Limit         int
	}
)

func UnmarshalBibleXml(biblexml []byte, w http.ResponseWriter) (*Bible, error) {
	var bible Bible

	err := xml.Unmarshal(biblexml, &bible)

	return &bible, err
}

func (bible *Bible) ToJson() []byte {
	j, e := json.Marshal(bible)
	if e != nil {
		panic(e)
	}
	return j
}

func (bc *BibleCollection) Add(b *Bible) {
	if bc.Bibles == nil {
		bc.Bibles = make(map[string]Bible)
	}
	bc.Bibles[b.Version] = *b
}

func Key(c appengine.Context, bc *BibleCollection, bibleVersion string, bookId int, chapterNum int) *datastore.Key {
	key := bibleVersion + "-" + strconv.Itoa(bookId) + "-" + strconv.Itoa(chapterNum)
	return datastore.NewKey(c, "Bible", key, 0, nil)
}

func (bc *BibleCollection) Persist(r *http.Request) error {
	c := appengine.NewContext(r)

	// var temp = []struct {
	// 	key  *datastore.Key
	// 	data interface{}
	// }{}

	for _, bible := range bc.Bibles {
		for _, book := range bible.Books {
			for _, chapter := range book.Chapters {
				_, err := datastore.Put(c, Key(c, bc, bible.Version, book.Id, chapter.Number), NewPersistedChapter(&bible, &book, &chapter))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func NewPersistedChapter(bible *Bible, book *Book, chapter *Chapter) *PersistedChapter {
	return &PersistedChapter{
		BibleVersion:  bible.Version,
		BookId:        book.Id,
		BookShortName: book.ShortName,
		BookLongName:  book.LongName,
		BookOtherName: book.OtherName,
		ChapterNumber: chapter.Number,
		Verses:        chapter.Verses,
	}
}

func Query(c appengine.Context, qo *QueryOptions) ([]PersistedChapter, error) {
	q := datastore.NewQuery("Bible")

	if qo.BookId != 0 {
		q = q.Filter("BookId =", qo.BookId)
	}

	if qo.BookShortName != "" {
		q = q.Filter("BookShortName =", qo.BookShortName)
	}

	if qo.ChapterNumber != 0 {
		q = q.Filter("ChapterNumber =", qo.ChapterNumber)
	}

	if qo.OrderBy != "" {
		q = q.Order(qo.OrderBy)
	} else {
		q = q.Order("ChapterNumber")
	}

	if qo.Project != nil {
		q = q.Project(qo.Project...)
	}

	if qo.Limit > 0 {
		q = q.Limit(qo.Limit)
	}

	var chapters []PersistedChapter
	_, err := q.GetAll(c, &chapters)
	if err != nil {
		return nil, err
	}

	return chapters, err
}
