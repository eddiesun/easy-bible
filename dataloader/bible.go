package dataloader

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

type (
	BibleCollection struct {
		Bibles map[string]Bible
	}

	Bible struct {
		Version string `xml:"version,attr"`
		Books   []Book `xml:"book"`
	}
	PBible struct {
		Type    string
		Key     *datastore.Key
		Version string
	}

	Book struct {
		Id        int       `xml:"id,attr"`
		ShortName string    `xml:"shortName,attr"`
		LongName  string    `xml:"name,attr"`
		OtherName string    `xml:"pinyin,attr"`
		Chapters  []Chapter `xml:"chapter"`
	}
	PBook struct {
		Type         string
		Key          *datastore.Key
		BibleVersion string
		Id           int
		ShortName    string
		LongName     string
		OtherName    string
	}

	Chapter struct {
		Number int     `xml:"number,attr"`
		Verses []Verse `xml:"verse"`
	}
	PChapter struct {
		Type          string
		Key           *datastore.Key
		BookId        int
		ChapterNumber int
	}

	Verse struct {
		Number int    `xml:"number,attr"`
		Text   string `xml:",innerxml"`
	}
	PVerse struct {
		Type          string
		Key           *datastore.Key
		ChapterNumber int
		VerseNumber   int
		Text          string
	}

	// PersistedVerse struct {
	// 	BibleVersion  string
	// 	BookId        int
	// 	BookShortName string
	// 	BookLongName  string
	// 	BookOtherName string
	// 	ChapterNumber int
	// 	VerseNumber   int
	// 	VerseText     string
	// }

	QueryOptions struct {
		BookId        int
		BookShortName string
		ChapterNumber int
		VerseFrom     int
		VerseTo       int
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

// func Key(c appengine.Context, bc *BibleCollection, bibleVersion string, bookId int, chapterNum int, verseNum int) *datastore.Key {
// 	key := bibleVersion + "-" + strconv.Itoa(bookId) + "-" + strconv.Itoa(chapterNum) + "-" + strconv.Itoa(verseNum)
// 	return datastore.NewKey(c, "Bible", key, 0, nil)
// }

func (bc *BibleCollection) Persist(c appengine.Context, w http.ResponseWriter) error {
	// Persist Bibles
	for _, bible := range bc.Bibles {
		fmt.Fprintf(w, "Persisting Bible: %v\n", bible.Version)
		c.Infof("Persisting Bible: %v\n", bible.Version)
		pb := &PBible{
			Type:    "Bible",
			Key:     datastore.NewKey(c, "Bible", "Bible-"+bible.Version, 0, nil),
			Version: bible.Version,
		}
		_, err := datastore.Put(c, pb.Key, pb)
		if err != nil {
			return err
		}

		// Persist Books
		for _, book := range bible.Books {
			fmt.Fprintf(w, "    Persisting Book: %v ...... ", book.LongName)
			c.Infof("    Persisting Book: %v ", book.LongName)
			pbk := &PBook{
				Type:         "Book",
				Key:          datastore.NewKey(c, "Book", "Book-"+pb.Version+"-"+strconv.Itoa(book.Id), 0, pb.Key),
				BibleVersion: pb.Version,
				Id:           book.Id,
				ShortName:    book.ShortName,
				LongName:     book.LongName,
				OtherName:    book.OtherName,
			}
			_, err := datastore.Put(c, pbk.Key, pbk)
			if err != nil {
				return err
			}

			// Persist Chapters
			//     Batch Chapters and Verses
			batchChapterKey := make([]*datastore.Key, 0)
			batchChapterStruct := make([]*PChapter, 0)
			batchVerseKey := make([]*datastore.Key, 0)
			batchVerseStruct := make([]*PVerse, 0)
			for _, chapter := range book.Chapters {
				pc := &PChapter{
					Type:          "Chapter",
					Key:           datastore.NewKey(c, "Chapter", "Chapter-"+pb.Version+"-"+strconv.Itoa(pbk.Id)+"-"+strconv.Itoa(chapter.Number), 0, pbk.Key),
					BookId:        pbk.Id,
					ChapterNumber: chapter.Number,
				}
				batchChapterKey = append(batchChapterKey, pc.Key)
				batchChapterStruct = append(batchChapterStruct, pc)

				// Persist Verses
				for _, verse := range chapter.Verses {
					pv := &PVerse{
						Type:          "Verse",
						Key:           datastore.NewKey(c, "Verse", "Verse-"+pb.Version+"-"+strconv.Itoa(pbk.Id)+"-"+strconv.Itoa(pc.ChapterNumber)+"-"+strconv.Itoa(verse.Number), 0, pc.Key),
						ChapterNumber: pc.ChapterNumber,
						VerseNumber:   verse.Number,
						Text:          verse.Text,
					}
					batchVerseKey = append(batchVerseKey, pv.Key)
					batchVerseStruct = append(batchVerseStruct, pv)
				}
			}
			_, err = datastore.PutMulti(c, batchVerseKey, batchVerseStruct)
			if err != nil {
				return err
			}
			_, err = datastore.PutMulti(c, batchChapterKey, batchChapterStruct)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "Done\n")
			c.Infof("    Done")

			// flush to browser
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		fmt.Fprintf(w, "Finished Persisting Bible: %v\n", bible.Version)
		c.Infof("Finished Persisting Bible: %v\n", bible.Version)

	}
	return nil
}

// func NewPersistedVerse(bible *Bible, book *Book, chapter *Chapter, verse *Verse) *PersistedVerse {
// 	return &PersistedVerse{
// 		BibleVersion:  bible.Version,
// 		BookId:        book.Id,
// 		BookShortName: book.ShortName,
// 		BookLongName:  book.LongName,
// 		BookOtherName: book.OtherName,
// 		ChapterNumber: chapter.Number,
// 		VerseNumber:   verse.Number,
// 		VerseText:     verse.Text,
// 	}
// }

/*
func Query(c appengine.Context, qo *QueryOptions) ([]PersistedVerse, error) {
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

	if qo.VerseFrom != 0 {
		q = q.Filter("VerseNumber >=", qo.VerseFrom)
	}

	if qo.VerseTo != 0 {
		q = q.Filter("VerseNumber <=", qo.VerseTo)
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

	var vs []PersistedVerse
	_, err := q.GetAll(c, &vs)
	if err != nil {
		return nil, err
	}

	return vs, err
}
*/
