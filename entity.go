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
	Bible  dataloader.Bible
	Dump   interface{}
}

type AutocompleteContext struct {
	Result map[string]*search.AutocompleteResult
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

func NewAutocompleteContext(c map[string]*search.AutocompleteResult) *AutocompleteContext {
	return &AutocompleteContext{c}
}
