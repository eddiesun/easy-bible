package dataloader

import (
	"encoding/json"
	"encoding/xml"
	// "fmt"
	"net/http"
)

type (
	Bible struct {
		Version *string
		Books   []Book `xml:"book"`
	}

	Book struct {
		ShortName string    `xml:"shortName,attr"`
		LongName  string    `xml:"name,attr"`
		OtherName string    `xml:"pinyin,attr"`
		Chapters  []Chapter `xml:"chapter"`
	}

	Chapter struct {
		Number uint32  `xml:"number,attr"`
		Verses []Verse `xml:"verse"`
	}

	Verse struct {
		Number uint32 `xml:"number,attr"`
		Text   string `xml:",innerxml"`
	}
)

func UnmarshalBibleXml(biblexml []byte, w http.ResponseWriter) (*Bible, error) {
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
