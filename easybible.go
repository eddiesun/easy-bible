package easybible

import (
	_ "fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"appengine"

	"dataloader"
	"search"
	"util"
)

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/autocomplete", autocompleteHandler)
	http.HandleFunc("/partial", partialHandler)
	http.HandleFunc("/bookmenu", bookmenuHandler)
	// http.HandleFunc("/load", loadHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Index Request *****\n")
	defer util.LogTime(c, time.Now(), "***** Index Request Served in ")

	ic := NewIndexContext()

	if b := r.URL.Query().Get("b"); b != "" {
		ic.InitBookId, _ = strconv.Atoi(b)
	}
	if ic.InitBookId <= 0 {
		ic.InitBookId = 1
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

	ic.LiteBooks = getLiteBooks(c, w)

	view := template.Must(template.ParseFiles("view/index.html", "view/header.html", "view/footer.html"))
	if err := view.Execute(w, ic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func partialHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Partial Request *****\n")
	defer util.LogTime(c, time.Now(), "***** Partial Request Served in ")

	iBook, _ := strconv.Atoi(r.URL.Query().Get("b"))
	iChapter, _ := strconv.Atoi(r.URL.Query().Get("c"))
	iFromVerse, _ := strconv.Atoi(r.URL.Query().Get("v1"))
	iToVerse, _ := strconv.Atoi(r.URL.Query().Get("v2"))

	pc := PartialContext{}

	book := getBook(c, w, "和合本", iBook) // get from memcache

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

	// render view
	view := template.Must(template.ParseFiles("view/partial.html"))
	if err := view.Execute(w, pc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Autocomplete Request *****\n")
	defer util.LogTime(c, time.Now(), "***** Autocomplete Request Served in ")

	userQuery := r.URL.Query().Get("query")
	if userQuery == "" {
		return
	}

	// get LiteBooks
	liteBooks := getLiteBooks(c, w)

	// get a new Search object
	c.Infof("Begin querying\n")
	s := search.NewSearch(c, userQuery)

	// get matched Books
	matchedLiteBooks := s.MatchedLiteBooks(c, liteBooks)

	// get partial bible that contains these books
	partialBible := getPartialBible(c, w, matchedLiteBooks)

	// search by chapter and verses
	s.FilterCV(c, partialBible)

	// render view
	view := template.Must(template.ParseFiles("view/autocomplete.html"))
	if err := view.Execute(w, NewAutocompleteContext(s.GetAutocompleteResult())); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func bookmenuHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("\n\n\n***** New Bookmenu Request *****\n")
	defer util.LogTime(c, time.Now(), "***** Bookmenu Request Served in ")

	iBook, _ := strconv.Atoi(r.URL.Query().Get("b"))

	// load bible object from memcache
	book := getBook(c, w, "和合本", iBook)

	bmc := PartialContext{
		BookId:           book.Id,
		BookLongName:     book.LongName,
		BookShortName:    book.ShortName,
		BookOtherName:    book.OtherName,
		ChapterNumber:    1,
		MaxChapterNumber: len(book.Chapters),
	}

	funcMap := template.FuncMap{
		"splitChapters": func(maxc int) []int {
			is := make([]int, maxc)
			for i := 0; i < maxc; i++ {
				is[i] = i + 1
			}
			return is
		},
	}

	// render view
	file, _ := ioutil.ReadFile("view/bookmenu.html")
	view, err := template.New("bookmenu").Funcs(funcMap).Parse(string(file))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := view.Execute(w, bmc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getBibleCollection(c appengine.Context, w http.ResponseWriter) *dataloader.BibleCollection {
	// load bible object from memcache
	c.Infof("Try to load Bible collection from memcache or from xml\n")
	defer util.LogTime(c, time.Now(), "Getting Bible collection took ")

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

func getBook(c appengine.Context, w http.ResponseWriter, version string, bookId int) *dataloader.Book {
	c.Infof("Try to get a Book from memcache or call getBibleCollection\n")
	defer util.LogTime(c, time.Now(), "Getting a Book took ")

	book, err := dataloader.MemcacheGetBook(c, version, bookId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// if cache miss
	if book == nil {
		bc := getBibleCollection(c, w)
		return bc.Bibles[version].SafeBook(bookId)
	}

	return book
}

func getLiteBooks(c appengine.Context, w http.ResponseWriter) []dataloader.LiteBook {
	c.Infof("Try to get LiteBooks from memcache or call getBibleCollection\n")
	defer util.LogTime(c, time.Now(), "Getting LiteBooks took ")

	liteBooks, err := dataloader.MemcacheGetLiteBooks(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if liteBooks == nil {
		return getBibleCollection(c, w).LiteBooks
	}
	return liteBooks
}

func getPartialBible(c appengine.Context, w http.ResponseWriter, liteBooks []dataloader.LiteBook) *dataloader.Bible {
	c.Infof("Try to get Partial Bible from memcache or call getBibleCollection\n")
	defer util.LogTime(c, time.Now(), "Getting Partial Bible took ")

	bible, err := dataloader.MemcacheGetPartialBible(c, liteBooks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// if cache miss
	if bible == nil {
		bc := getBibleCollection(c, w)
		var partialBible dataloader.Bible
		fullBible := bc.FirstBible()
		partialBible.Version = fullBible.Version
		for _, book := range liteBooks {
			partialBible.Books = append(partialBible.Books, *fullBible.Book(book.Id))
		}
		return &partialBible
	}

	return bible
}
