package search

import (
	"appengine"
	"dataloader"
	"regexp"
	"strconv"
	_ "strings"
)

type (
	Search struct {
		userInput    string
		digitStr     string
		nondigitStr  string
		queryOptions []dataloader.QueryOptions
		fragment     fragments
		// resultSet    []dataloader.PersistedVerse
		// entries      []entry
		autocompleteResult map[string]*AutocompleteResult
	}

	fragments struct {
		bookShortNames    []string
		bookLongNames     []string
		chapterVerseRange map[string]chapterVerseRange
	}

	chapterVerseRange struct {
		c  string
		v1 string
		v2 string
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

func NewSearch(userInput string, c appengine.Context) *Search {
	s := new(Search)
	s.userInput = userInput

	s.digitStr = s.getDigitStr()
	s.nondigitStr = s.getNondigitStr()

	c.Debugf("Input tokens: nondigitStr=%s, digitStr=%s\n", s.nondigitStr, s.digitStr)

	// add fragments for chapters and verses
	if len(s.digitStr) > 0 {
		// handle the case for chapter X verse Y
		for i := 0; i < len(s.digitStr); i++ {
			var c = s.digitStr[:i+1]
			var v = s.digitStr[i+1:]
			if len(c) > 3 || len(v) > 3 {
				continue
			}
			s.addChapterVerseRangeToFragments(c, v, v)
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
				s.addChapterVerseRangeToFragments(c, v1, v2)
			}
		}

	} else {
		s.addChapterVerseRangeToFragments("1", "", "")
	}

	// add fragments for book short names
	s.fragment.bookShortNames = append(s.fragment.bookShortNames, s.nondigitStr)

	return s
}

func (s *Search) GetQueryOptions() []dataloader.QueryOptions {
	qs := make([]dataloader.QueryOptions, 0)

	for _, r := range s.fragment.chapterVerseRange {
		var q = dataloader.QueryOptions{}
		q.ChapterNumber, _ = strconv.Atoi(r.c)
		if r.v1 != "" {
			q.VerseFrom, _ = strconv.Atoi(r.v1)
		}
		if r.v2 != "" {
			q.VerseTo, _ = strconv.Atoi(r.v2)
		}
		qs = append(qs, q)
	}

	return qs
}

func (s *Search) getDigitStr() string {
	re := regexp.MustCompile(`\d[\d\s]*`)
	return re.FindString(s.userInput)
}

func (s *Search) getNondigitStr() string {
	re := regexp.MustCompile(`[\D\s]+`)
	return re.FindString(s.userInput)
}

func (s *Search) addChapterVerseRangeToFragments(chapterNum string, verse1 string, verse2 string) {
	if s.fragment.chapterVerseRange == nil {
		s.fragment.chapterVerseRange = make(map[string]chapterVerseRange)
	}
	tmp := chapterVerseRange{
		chapterNum,
		verse1,
		verse2,
	}
	key := chapterNum + "-" + verse1 + "-" + verse2
	s.fragment.chapterVerseRange[key] = tmp
}

/*
func (s *Search) AddAutocompleteData(set []dataloader.PersistedVerse, q dataloader.QueryOptions, c appengine.Context) {
	if len(set) > 0 {
		c.Debugf("Book Filter: %s, %+v\n", s.nondigitStr, set)
		for _, v := range set {
			c.Debugf("\t%+v\n", v)
			if s.nondigitStr == "" ||
				strings.Contains(v.BookOtherName, s.nondigitStr) ||
				strings.Contains(s.nondigitStr, v.BookShortName) ||
				strings.Contains(v.BookLongName, s.nondigitStr) {
				c.Debugf("\tMatch\n")
				s.addAutocompleteResult(v, q)
			} else {
				c.Debugf("\tNo Match\n")
			}
		}
	}
}

func (s *Search) addAutocompleteResult(v dataloader.PersistedVerse, q dataloader.QueryOptions) {
	if s.autocompleteResult == nil {
		s.autocompleteResult = make(map[string]*AutocompleteResult)
	}
	key := strconv.Itoa(v.BookId) + "-" + strconv.Itoa(q.ChapterNumber) + "-" + strconv.Itoa(q.VerseFrom) + "-" + strconv.Itoa(q.VerseTo)
	if arOnFile, exists := s.autocompleteResult[key]; !exists {
		var ar = AutocompleteResult{
			BibleVersion:  v.BibleVersion,
			BookId:        v.BookId,
			BookShortName: v.BookShortName,
			BookLongName:  v.BookLongName,
			BookOtherName: v.BookOtherName,
			ChapterNumber: v.ChapterNumber,
			VerseFrom:     q.VerseFrom,
			VerseText:     make([]dataloader.Verse, 0),
		}
		ar.VerseText = append(ar.VerseText, dataloader.Verse{Number: v.VerseNumber, Text: v.VerseText})
		ar.VerseTo = min(q.VerseTo, ar.VerseText[len(ar.VerseText)-1].Number)
		s.autocompleteResult[key] = &ar
	} else {
		arOnFile.VerseText = append(arOnFile.VerseText, dataloader.Verse{Number: v.VerseNumber, Text: v.VerseText})
		arOnFile.VerseTo = min(q.VerseTo, arOnFile.VerseText[len(arOnFile.VerseText)-1].Number)
	}
}
*/
func (s *Search) GetAutocompleteResult() map[string]*AutocompleteResult {
	return s.autocompleteResult
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
