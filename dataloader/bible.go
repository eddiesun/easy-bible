package dataloader

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"time"
	"util"
)

type (
	BibleCollection struct {
		Bibles    map[string]*Bible
		LiteBooks []LiteBook
	}

	Bible struct {
		Version string `xml:"version,attr"`
		Books   []Book `xml:"book"`
	}
	PBible struct {
		Type    string
		Key     *datastore.Key
		Version string
	}

	Book struct {
		Id        int       `xml:"id,attr"`
		ShortName string    `xml:"shortName,attr"`
		LongName  string    `xml:"name,attr"`
		OtherName string    `xml:"pinyin,attr"`
		Chapters  []Chapter `xml:"chapter"`
	}
	PBook struct {
		Type         string
		Key          *datastore.Key
		BibleVersion string
		Id           int
		ShortName    string
		LongName     string
		OtherName    string
	}
	LiteBook struct {
		Id          int
		ShortName   string
		LongName    string
		OtherName   string
		NumChapters int
	}

	Chapter struct {
		Number int     `xml:"number,attr"`
		Verses []Verse `xml:"verse"`
	}
	PChapter struct {
		Type          string
		Key           *datastore.Key
		BookId        int
		ChapterNumber int
	}

	Verse struct {
		Number int    `xml:"number,attr"`
		Text   string `xml:",innerxml"`
	}
	PVerse struct {
		Type          string
		Key           *datastore.Key
		ChapterNumber int
		VerseNumber   int
		Text          string
	}
)

func UnmarshalBibleXml(biblexml []byte) (*Bible, error) {
	var bible Bible
	err := xml.Unmarshal(biblexml, &bible)
	return &bible, err
}

func (bible *Bible) ToJson() []byte {
	j, e := json.Marshal(bible)
	if e != nil {
		panic(e)
	}
	return j
}

func (bible *Bible) Book(bookId int) *Book {
	if bookId >= 1 && bookId <= len(bible.Books) {
		return &bible.Books[bookId-1]
	}
	return nil
}

func (bible *Bible) SafeBook(bookId int) *Book {
	if bookId < 1 {
		bookId = 1
	}
	if bookId > len(bible.Books) {
		bookId = len(bible.Books)
	}
	return &bible.Books[bookId-1]
}

func (b *Book) Chapter(number int) *Chapter {
	if number > len(b.Chapters) {
		return nil
	}
	if number == 0 {
		number = 1
	}
	return &b.Chapters[number-1]
}

func (b *Book) SafeChapter(number int) *Chapter {
	if number <= 0 {
		number = 1
	}
	if number > len(b.Chapters) {
		number = len(b.Chapters)
	}
	return &b.Chapters[number-1]
}

func NewLiteBook(b *Book) *LiteBook {
	lb := new(LiteBook)
	lb.Id = b.Id
	lb.ShortName = b.ShortName
	lb.LongName = b.LongName
	lb.OtherName = b.OtherName
	lb.NumChapters = len(b.Chapters)
	return lb
}

func (c *Chapter) GetVerses(from int, to int) []Verse {
	if from > len(c.Verses) {
		return nil
	}
	if from == 0 {
		from = 1
	}
	if to == 0 || to > len(c.Verses) {
		to = len(c.Verses)
	}
	return c.Verses[from-1 : to]
}

func (c *Chapter) SafeGetVerses(from int, to int) []Verse {
	if from == 0 || from > len(c.Verses) {
		from = 1
	}
	if to == 0 || to > len(c.Verses) {
		to = len(c.Verses)
	}
	return c.Verses[from-1 : to]
}

func (bc *BibleCollection) Add(b *Bible) {
	if bc.Bibles == nil {
		bc.Bibles = make(map[string]*Bible)
	}
	bc.Bibles[b.Version] = b
}

func (bc *BibleCollection) FirstBible() *Bible {
	for _, v := range bc.Bibles {
		return v
	}
	return nil
}

func (bc *BibleCollection) AddLiteBooks(b *Bible) {
	bc.LiteBooks = make([]LiteBook, len(b.Books))
	var i = 0
	for _, book := range b.Books {
		bc.LiteBooks[i] = *(NewLiteBook(&book))
		i++
	}
}

func LoadXmlBible(c appengine.Context, pathToBibleXml string) (*Bible, error) {
	defer util.LogTime(c, time.Now(), "    Loading xml bible took ")

	biblexml, err := ioutil.ReadFile(pathToBibleXml)
	if err != nil {
		return nil, err
	}
	bible, err := UnmarshalBibleXml(biblexml)
	if err != nil {
		return nil, err
	}
	return bible, nil
}

func NewBibleCollection(c appengine.Context, pathToBibleXml string) (*BibleCollection, error) {
	bc := new(BibleCollection)
	bible, err := LoadXmlBible(c, pathToBibleXml)
	if err != nil {
		return bc, err
	}
	// add bible to bible collection
	bc.Add(bible)
	bc.AddLiteBooks(bible)
	return bc, nil
}
