package dataloader

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"net/http"
	"strconv"
)

type (
	QueryOptions struct {
		Type          string
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

			if len(batchVerseKey) > 1300 {
				for i := 0; i < len(batchVerseKey); i += 1000 {
					var upperBound = min(i+1000, len(batchVerseKey))
					c.Infof("        Persisting Verses [%v, %v)\n", i, upperBound)
					_, err = datastore.PutMulti(c, batchVerseKey[i:upperBound], batchVerseStruct[i:upperBound])
					if err != nil {
						return err
					}
				}
			} else {
				_, err = datastore.PutMulti(c, batchVerseKey, batchVerseStruct)
				if err != nil {
					return err
				}
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

func GetAllBooks(c appengine.Context) ([]PBook, error) {
	q := datastore.NewQuery("Book")
	books := make([]PBook, 0)
	_, err := q.GetAll(c, &books)
	return books, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
