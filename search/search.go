package search

import (
	"dataloader"
	"regexp"
)

type (
	Search struct {
		userInput    string
		digitStr     string
		nondigitStr  string
		queryOptions []dataloader.QueryOptions
		entries      []entry
	}

	entry struct {
		bookShortName string
		bookLongName  string
		chapterNumber int
		verseNumber   int
	}
)

func NewSearch(userInput string) *Search {
	s := new(Search)
	s.userInput = userInput

	s.digitStr = s.getDigitStr()
	s.nondigitStr = s.getNondigitStr()

	return s
}

func (s *Search) GetQueryOptions() []dataloader.QueryOptions {
	return nil
}

func (s *Search) getDigitStr() string {
	re := regexp.MustCompile(`\d[\d\s]+`)
	return re.FindString(s.userInput)
}

func (s *Search) getNondigitStr() string {
	re := regexp.MustCompile(`[\D\s]+`)
	return re.FindString(s.userInput)
}
