package search

import (
	"appengine"
	"dataloader"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util"
)

const (
	AUTOCOMPLETE_MAX_NUM_BOOKS = 8
)

type (
	Search struct {
		userInput          string
		digitStr           string
		nondigitStr        string
		chapterVerseRange  map[string]chapterVerseRange
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

func NewSearch(c appengine.Context, userInput string) *Search {
	s := new(Search)
	s.userInput = strings.Replace(userInput, " ", "", -1)
	s.userInput = strings.ToLower(s.userInput)

	c.Debugf("    Input string: %s\n", s.userInput)

	s.digitStr = s.getDigitStr()
	s.nondigitStr = s.getNondigitStr()

	c.Debugf("    Input tokens: nondigitStr=%s, digitStr=%s\n", s.nondigitStr, s.digitStr)

	// generate chapter verse queries
	s.generateChapterVerseRange(c)

	return s
}

func (s *Search) MatchedLiteBooks(c appengine.Context, books []dataloader.LiteBook) []dataloader.LiteBook {
	defer util.LogTime(c, time.Now(), "Filtering Books took ")

	filteredLiteBooks := make([]dataloader.LiteBook, 0, 10)
	for _, book := range books {
		if strings.Contains(book.OtherName, s.nondigitStr) || strings.Contains(book.LongName, s.nondigitStr) {
			c.Infof("        Matched Book: %s\n", book.LongName)
			filteredLiteBooks = append(filteredLiteBooks, book)
		}
		if len(filteredLiteBooks) > AUTOCOMPLETE_MAX_NUM_BOOKS {
			break
		}
	}
	return filteredLiteBooks
}

func (s *Search) FilterCV(c appengine.Context, partialBible *dataloader.Bible) {
	defer util.LogTime(c, time.Now(), "Filtering Chapters and Verses took ")
	// First filter by books
	for _, book := range partialBible.Books {
		if s.chapterVerseRange == nil {
			c.Infof("        No Specified Chapter or Verses, assume Chapter 1 Verse 1\n")
			chapter := book.Chapter(1)
			verses := chapter.GetVerses(1, 1)
			r := AutocompleteResult{
				BibleVersion:  "和合本",
				BookId:        book.Id,
				BookShortName: book.ShortName,
				BookLongName:  book.LongName,
				BookOtherName: book.OtherName,
				ChapterNumber: 1,
				VerseFrom:     1,
				VerseTo:       1,
				VerseText:     verses,
			}
			s.autocompleteResult = append(s.autocompleteResult, r)
		} else {
			// specified chapter and/or verses
			for _, cvr := range s.chapterVerseRange {
				chapter := book.Chapter(cvr.c)
				if chapter == nil {
					continue
				}
				verses := chapter.GetVerses(cvr.v1, cvr.v2)
				if verses == nil {
					continue
				}
				c.Infof("        Matched Chapter: %v, Verse %v - %v\n", cvr.c, cvr.v1, cvr.v2)
				r := AutocompleteResult{
					BibleVersion:  "和合本",
					BookId:        book.Id,
					BookShortName: book.ShortName,
					BookLongName:  book.LongName,
					BookOtherName: book.OtherName,
					ChapterNumber: cvr.c,
					VerseFrom:     cvr.v1,
					VerseTo:       min(cvr.v2, verses[len(verses)-1].Number),
					VerseText:     verses[:min(3, len(verses))],
				}
				s.autocompleteResult = append(s.autocompleteResult, r)
			}
		}
	}
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

func (s *Search) GetAutocompleteResult() *[]AutocompleteResult {
	return &s.autocompleteResult
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
