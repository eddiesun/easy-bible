package search

import (
	"appengine"
	"appengine/datastore"
	"dataloader"
	"regexp"
	"strconv"
	"strings"
)

const (
	AUTOCOMPLETE_MAX_NUM_BOOKS = 5
)

type (
	Search struct {
		userInput         string
		digitStr          string
		nondigitStr       string
		filteredBooks     []dataloader.PBook
		queryOptions      []dataloader.QueryOptions
		chapterVerseRange map[string]chapterVerseRange
		// resultSet    []dataloader.PersistedVerse
		// entries      []entry
		autocompleteResult []AutocompleteResult
	}

	chapterVerseRange struct {
		c  int
		v1 int
		v2 int
	}

	AutocompleteResult struct {
		BibleVersion  string
		BookId        int
		BookShortName string
		BookLongName  string
		BookOtherName string
		ChapterNumber int
		VerseFrom     int
		VerseTo       int
		VerseText     []dataloader.Verse
	}
)

func NewSearch(userInput string, c appengine.Context) (*Search, error) {
	s := new(Search)
	s.userInput = strings.Replace(userInput, " ", "", -1)

	s.digitStr = s.getDigitStr()
	s.nondigitStr = s.getNondigitStr()

	c.Debugf("Input tokens: nondigitStr=%s, digitStr=%s\n", s.nondigitStr, s.digitStr)

	// First filter by books
	//     Get all books from datastore
	books, err := dataloader.GetAllBooks(c)
	if err != nil {
		return nil, err
	}
	//     Filter books by name
	s.filterBooks(c, books)
	c.Debugf("Matched Books:\n%+v\n", s.filteredBooks)

	// Secondly filter by chapters and verses
	// generate chapter verse queries
	s.generateChapterVerseRange(c)

	for _, b := range s.filteredBooks {

		if len(s.chapterVerseRange) <= 0 {
			// TO DO
		} else {
			for _, cvr := range s.chapterVerseRange {
				c.Debugf("Chapter and Verses: %+v\n", cvr)
				q := datastore.NewQuery("Verse").Ancestor(b.Key).Filter("ChapterNumber =", cvr.c).Order("ChapterNumber").Order("VerseNumber")
				if cvr.v1 != 0 {
					q = q.Filter("VerseNumber >=", cvr.v1)
				}
				if cvr.v2 != 0 {
					q = q.Filter("VerseNumber <=", cvr.v2)
				}

				var verses []dataloader.PVerse
				_, err := q.GetAll(c, &verses)
				if err != nil {
					return nil, err
				}
				c.Debugf("Fetched Verses: %+v\n", verses)
				s.addAutocompleteResult(b, verses, cvr)
			}
		}
	}

	return s, err
}

func (s *Search) generateChapterVerseRange(c appengine.Context) {
	if len(s.digitStr) > 0 {
		// handle the case for chapter X verse Y
		for i := 0; i < len(s.digitStr); i++ {
			var c = s.digitStr[:i+1]
			var v = s.digitStr[i+1:]
			if len(c) > 3 || len(v) > 3 {
				continue
			}
			s.addChapterVerseRange(c, v, v)
		}

		// handle the case for chapter X verse Y to verse Z
		for i := 0; i < len(s.digitStr); i++ {
			for j := i + 1; j < len(s.digitStr); j++ {
				var c = s.digitStr[:i+1]
				var v1 = s.digitStr[i+1 : j+1]
				var v2 = s.digitStr[j+1:]

				if len(c) > 3 || len(v1) > 3 || len(v2) > 3 {
					continue
				}
				if v1 == "" || v2 == "" {
					continue
				}
				iv1, err1 := strconv.Atoi(v1)
				iv2, err2 := strconv.Atoi(v2)
				if err1 != nil || err2 != nil {
					continue
				}
				if iv1 >= iv2 {
					continue
				}
				s.addChapterVerseRange(c, v1, v2)
			}
		}

	} else {
		s.addChapterVerseRange("1", "", "")
	}
}

func (s *Search) getDigitStr() string {
	re := regexp.MustCompile(`\d[\d\s]*`)
	return re.FindString(s.userInput)
}

func (s *Search) getNondigitStr() string {
	re := regexp.MustCompile(`[\D\s]+`)
	return re.FindString(s.userInput)
}

func (s *Search) addChapterVerseRange(chapterNum string, verse1 string, verse2 string) {
	if s.chapterVerseRange == nil {
		s.chapterVerseRange = make(map[string]chapterVerseRange)
	}
	iChapterNum, _ := strconv.Atoi(chapterNum)
	iVerse1, _ := strconv.Atoi(verse1)
	iVerse2, _ := strconv.Atoi(verse2)
	tmp := chapterVerseRange{
		iChapterNum,
		iVerse1,
		iVerse2,
	}
	key := chapterNum + "-" + verse1 + "-" + verse2
	s.chapterVerseRange[key] = tmp
}

func (s *Search) filterBooks(c appengine.Context, books []dataloader.PBook) {
	s.filteredBooks = books[:0]
	if len(s.nondigitStr) > 0 {
		for _, book := range books {
			var matchBookOtherName = strings.Contains(book.OtherName, s.nondigitStr)
			var matchBookShortName = strings.Contains(s.nondigitStr, book.ShortName)
			var matchBookLongName = strings.Contains(book.LongName, s.nondigitStr)
			c.Debugf("BOOKS Filter: matchBookOtherName=%v, matchBookShortName=%v, matchBookLongName=%v,\n", matchBookOtherName, matchBookShortName, matchBookLongName)
			if matchBookOtherName || matchBookShortName || matchBookLongName {
				s.filteredBooks = append(s.filteredBooks, book)
				if len(s.filteredBooks) > AUTOCOMPLETE_MAX_NUM_BOOKS {
					break
				}
			}
		}
	} else {
		s.filteredBooks = books[:AUTOCOMPLETE_MAX_NUM_BOOKS]
	}
}

func (s *Search) addAutocompleteResult(book dataloader.PBook, verses []dataloader.PVerse, cvr chapterVerseRange) {
	if len(verses) <= 0 {
		return
	}

	t := AutocompleteResult{
		BibleVersion:  book.BibleVersion,
		BookId:        book.Id,
		BookShortName: book.ShortName,
		BookLongName:  book.LongName,
		BookOtherName: book.OtherName,
		ChapterNumber: cvr.c,
		VerseFrom:     cvr.v1,
		VerseTo:       min(cvr.v2, verses[len(verses)-1].VerseNumber),
	}
	for _, v := range verses {
		t.VerseText = append(t.VerseText, dataloader.Verse{v.VerseNumber, v.Text})
	}

	s.autocompleteResult = append(s.autocompleteResult, t)
}

func (s *Search) GetAutocompleteResult() *[]AutocompleteResult {
	return &s.autocompleteResult
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
