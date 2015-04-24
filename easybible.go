package easybible

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"appengine"
	_ "appengine/datastore"
	_ "appengine/user"

	"dataloader"
	_ "search"
)

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/autocomplete", autocompleteHandler)
	http.HandleFunc("/load", loadHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	view := template.Must(template.ParseFiles("view/index.html", "view/header.html", "view/footer.html"))

	// get bible object
	// c := appengine.NewContext(r)

	// q := datastore.NewQuery("Bible")
	// var verses []dataloader.PersistedVerse
	// if _, err := q.GetAll(c, &verses); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	ic := NewIndexContext()

	if err := view.Execute(w, ic); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// fmt.Fprintf(w, "%+v", verses)
}

func autocompleteHandler(w http.ResponseWriter, r *http.Request) {
	/*
		c := appengine.NewContext(r)

		userQuery := r.URL.Query().Get("query")
		// userQuery := "亞8213"
		s := search.NewSearch(userQuery, c)

		qos := s.GetQueryOptions()

		for _, qo := range qos {
			vs, err := dataloader.Query(c, &qo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			c.Debugf("qo: %+v\n", qo)
			s.AddAutocompleteData(vs, qo, c)
		}

		view := template.Must(template.ParseFiles("view/autocomplete.html"))
		if err := view.Execute(w, NewAutocompleteContext(s.GetAutocompleteResult())); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	c := appengine.NewContext(r)
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
	if err := bc.Persist(c, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Fprintf(w, "%s", bible.ToJson())

	// fmt.Fprintf(w, "%s", biblexml)

	fmt.Fprintf(w, "ok")
}
