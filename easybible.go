package easybible

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"appengine"
	"appengine/datastore"
	_ "appengine/user"

	"dataloader"
	"search"
)

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/autocomplete", autocompleteHandler)
	http.HandleFunc("/load", loadHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	view := template.Must(template.ParseFiles("view/index.html", "view/header.html", "view/footer.html"))

	// get bible object
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Bible")
	var chapters []dataloader.PersistedChapter
	if _, err := q.GetAll(c, &chapters); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ic := NewIndexContext()
	// ic.Dump = chapters

	if err := view.Execute(w, ic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "!!!!!%+v", chapters)
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// userQuery := r.URL.Query().Get("query")
	userQuery := "亞82"
	s := search.NewSearch(userQuery)

	_ := s.GetQueryOptions()

	var options = dataloader.QueryOptions{
		BookId:        38,
		ChapterNumber: 8,
	}

	chapters, err := dataloader.Query(c, &options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	view := template.Must(template.ParseFiles("view/autocomplete.html"))
	if err := view.Execute(w, NewAutocompleteContext(chapters)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// fmt.Fprintf(w, "%+v", chapters)
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	bc := new(dataloader.BibleCollection)

	biblexml, err := ioutil.ReadFile("data/bible_partial.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bible, err := dataloader.UnmarshalBibleXml(biblexml, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// add bible to bible collection
	bc.Add(bible)

	// store the processed bible
	if err := bc.Persist(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", bible.ToJson())

	// fmt.Fprintf(w, "%s", biblexml)
}
