package easybible

import (
	_ "fmt"
	"html/template"
	"net/http"
	"strconv"

	"appengine"
	_ "appengine/memcache"

	"dataloader"
	"search"
)

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/autocomplete", autocompleteHandler)
	http.HandleFunc("/partial", partialHandler)
	// http.HandleFunc("/load", loadHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// c := appengine.NewContext(r)

	ic := NewIndexContext()

	if b := r.URL.Query().Get("b"); b != "" {
		ic.InitBookId, _ = strconv.Atoi(b)
	}
	if c := r.URL.Query().Get("c"); c != "" {
		ic.InitChapter, _ = strconv.Atoi(c)
	}
	if v1 := r.URL.Query().Get("v1"); v1 != "" {
		ic.InitVerseBegin, _ = strconv.Atoi(v1)
	}
	if v2 := r.URL.Query().Get("v2"); v2 != "" {
		ic.InitVerseEnd, _ = strconv.Atoi(v2)
	}

	view := template.Must(template.ParseFiles("view/index.html", "view/header.html", "view/footer.html"))
	if err := view.Execute(w, ic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func partialHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Partial Request *****\n")

	iBook, _ := strconv.Atoi(r.URL.Query().Get("b"))
	iChapter, _ := strconv.Atoi(r.URL.Query().Get("c"))
	iFromVerse, _ := strconv.Atoi(r.URL.Query().Get("v1"))
	iToVerse, _ := strconv.Atoi(r.URL.Query().Get("v2"))

	pc := PartialContext{}

	bc := getBibleCollection(c, w)
	for _, bible := range bc.Bibles {
		book := bible.SafeBook(iBook)
		pc.BookId = book.Id
		pc.BookLongName = book.LongName
		pc.BookShortName = book.ShortName
		pc.BookOtherName = book.OtherName

		chapter := book.SafeChapter(iChapter)
		pc.ChapterNumber = chapter.Number
		pc.MaxChapterNumber = len(book.Chapters)

		verses := chapter.SafeGetVerses(iFromVerse, iToVerse)
		pc.Verses = verses
		pc.MaxVerseNumber = len(chapter.Verses)
	}

	// render view
	view := template.Must(template.ParseFiles("view/partial.html"))
	if err := view.Execute(w, pc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Autocomplete Request *****\n")

	userQuery := r.URL.Query().Get("query")
	if userQuery == "" {
		return
	}

	// load bible object from memcache
	bc := getBibleCollection(c, w)

	// ready to generate search conditions
	c.Infof("Begin querying\n")
	s, err := search.NewSearch(c, userQuery, bc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// render view
	view := template.Must(template.ParseFiles("view/autocomplete.html"))
	if err := view.Execute(w, NewAutocompleteContext(s.GetAutocompleteResult())); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getBibleCollection(c appengine.Context, w http.ResponseWriter) *dataloader.BibleCollection {
	// load bible object from memcache
	c.Infof("Try to load Bible collection from memcache\n")
	bc, err := dataloader.MemcacheGet(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// if cache miss, load from file
	if bc == nil {
		c.Infof("Memcache MISS for Bible collection\n")
		// read bible collection xml file
		bc, err = dataloader.NewBibleCollection(c, "data/bible_processed.xml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Save bible collection to memcache
		if dataloader.MemcachePut(c, bc) != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		c.Infof("Memcache HIT for Bible collection\n")
	}
	return bc
}

/*
func loadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	c := appengine.NewContext(r)
	bc := new(dataloader.BibleCollection)

	biblexml, err := ioutil.ReadFile("data/bible_processed.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bible, err := dataloader.UnmarshalBibleXml(biblexml)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// add bible to bible collection
	bc.Add(bible)

	// store the processed bible
	if err := bc.Persist(c, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Fprintf(w, "%s", bible.ToJson())

	// fmt.Fprintf(w, "%s", biblexml)

	fmt.Fprintf(w, "ok")
}
*/
