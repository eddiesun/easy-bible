package easybible

import (
	"dataloader"
	"search"
)

type HeaderContext struct {
	Title   string
	Scripts []string
	Styles  []string
}

type FooterContext struct {
	GATrackingId string
}

type IndexContext struct {
	Header HeaderContext
	Footer FooterContext
	// Bible  dataloader.Bible
	InitBookId     int
	InitChapter    int
	InitVerseBegin int
	InitVerseEnd   int
	Dump           interface{}
}

type AutocompleteContext struct {
	Result *[]search.AutocompleteResult
}

type PartialContext struct {
	BookId           int
	BookLongName     string
	BookShortName    string
	BookOtherName    string
	ChapterNumber    int
	MaxChapterNumber int
	Verses           []dataloader.Verse
	MaxVerseNumber   int
}

func DefaultHeaderContext() HeaderContext {
	return HeaderContext{Title: "Easy Bible Lookup"}
}

func DefaultFooterContext() FooterContext {
	return FooterContext{GATrackingId: "GA-xxxxx-xxxxx"}
}

func NewIndexContext() *IndexContext {
	c := &IndexContext{
		Header: DefaultHeaderContext(),
		Footer: DefaultFooterContext(),
	}

	c.Header.Scripts = append(c.Header.Scripts, "/static/js/index.js")
	c.Header.Styles = append(c.Header.Styles, "/static/css/index.css")

	return c
}

func NewAutocompleteContext(c *[]search.AutocompleteResult) AutocompleteContext {
	return AutocompleteContext{c}
}
