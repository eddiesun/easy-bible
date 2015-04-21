package easybible

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	_ "appengine"
	_ "appengine/datastore"
	_ "appengine/user"

	"dataloader"
)

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/load", loadHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	view := template.Must(template.ParseFiles("view/index.html", "view/header.html", "view/footer.html"))

	if err := view.Execute(w, NewIndexContext()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	biblexml, err := ioutil.ReadFile("data/bible_partial.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bible, err := dataloader.UnmarshalBibleXml(biblexml, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "%s", bible.ToJson())

	// fmt.Fprintf(w, "%s", biblexml)
}
