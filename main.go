package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// https://www.golang-book.com/books/intro/9
type Recording struct {
	catalogNumber     int
	date              string
	length            string
	speciesCommon     string
	speciesScientific string
	recordist         string
	url               string

	//location
	//sound type
	//quality
}

var recordings []Recording

var macaulayUrl = "http://media2.macaulaylibrary.org/Audio/Audio1/"

var searchUrl = "http://macaulaylibrary.org/search?&asset_format_id=1000&collection_type_id=1&layout=1&sort=21&page="

func main() {

	var wg sync.WaitGroup

	if len(os.Args) == 1 {
		fmt.Println("you need to pass in num of search pages to scrape. 1729 is currently the max")
		return
	}

	numOfPagesToVisit, _ := strconv.Atoi(os.Args[1])

	for i := 1; i < numOfPagesToVisit+1; i++ {
		fmt.Println("page number: ", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			getUrlsFromPage(i - 1)
			fmt.Println("size of array: ", len(recordings))
		}()

		if i%10 == 0 {
			time.Sleep(30000 * time.Millisecond)
		}

	}
	wg.Wait()

	fmt.Println("num of recordings: ", len(recordings))
	writeUrlsToFile()
	fmt.Println("done~")
}

func getUrlsFromPage(pageNum int) {

	url := searchUrl + strconv.Itoa(pageNum)

	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Fatal(err)
	}

	doc.Find("div.catalog").Each(func(i int, s *goquery.Selection) {

		recordingNumber := strings.TrimSpace(s.Text())
		
		url := ""
		if(len(recordingNumber) > 2 ){
			url = macaulayUrl + recordingNumber[:2] + "/" + recordingNumber
		}
		recording := Recording{url: url}
		recordings = append(recordings, recording)

	})

	//get species scientific name
	doc.Find("div.search-results").Find("div.subject").Each(func(i int, s *goquery.Selection) {
		scientificName := s.Find("h4.indent").Children().Text()

		if scientificName == "" {
			scientificName = s.Find("h4").Text()
		}

		recordings[i].speciesScientific = scientificName
	})

	//get species common name
	doc.Find("div.search-results").Find("div.subject").Each(func(i int, s *goquery.Selection) {

		name := ""

		commonName := s.Find("h4.indent").Text()

		if commonName != "" {
			name = strings.TrimSpace(s.Find("h4.indent").Children().Remove().End().Text())
		} else {
			// no common name
			name = "No Common Name"
		}

		recordings[i].speciesCommon = name

	})

	//get date
	doc.Find("div.search-results").Find("div.date").Each(func(i int, s *goquery.Selection) {

		date := strings.TrimSpace(s.Text())
		recordings[i].date = date

	})

	//get recordist, need to fixup first name
	doc.Find("div.search-results").Find("div.recordist").Each(func(i int, s *goquery.Selection) {

		recordist := ""
		lastName := strings.TrimSpace(s.Children().Remove().End().Text())
		firstName := strings.TrimSpace(s.Find("div.indent").Text())
		recordist = lastName + firstName
		recordings[i].recordist = recordist

	})

	//get length
	doc.Find("div.search-results").Find("div.length").Each(func(i int, s *goquery.Selection) {

		length := strings.TrimSpace(s.Text())
		recordings[i].length = length

	})

	//get catalog number
	doc.Find("div.catalog").Each(func(i int, s *goquery.Selection) {

		catalogNumber, _ := strconv.Atoi(strings.TrimSpace(s.Text()))

		recordings[i].catalogNumber = catalogNumber

	})

}

//from so: http://stackoverflow.com/questions/5884154/golang-read-text-file-into-string-array-and-write

func writeUrlsToFile() {
	var (
		file *os.File
		err  error
	)

	if file, err = os.Create("audio_urls.json"); err != nil {
		return
	}
	defer file.Close()

	file.WriteString(strings.TrimSpace("{\"urls\":["))

	for i := 0; i < len(recordings); i++ {

		commaBracket := "\"},"
		if i == len(recordings)-1 {

			commaBracket = "\"}"

		}
		_, err := file.WriteString(strings.TrimSpace(
			"{" +
				"\"catalogNumber\":\"" +
				strings.TrimSpace(strconv.Itoa(recordings[i].catalogNumber)) +
				"\"" +

				", " +
				"\"date\":\"" +
				strings.TrimSpace(recordings[i].date) +
				"\"" +

				", " +
				"\"length\":\"" +
				strings.TrimSpace(recordings[i].length) +
				"\"" +

				", " +
				"\"speciesCommon\":\"" +
				strings.TrimSpace(recordings[i].speciesCommon) +
				"\"" +

				", " +
				"\"speciesScientific\":\"" +
				strings.TrimSpace(recordings[i].speciesScientific) +
				"\"" +

				", " +
				"\"recordist\":\"" +
				strings.TrimSpace(recordings[i].recordist) +
				"\"" +

				", " +
				"\"url\":\"" +
				recordings[i].url +

				commaBracket +
				"\n"))

		if err != nil {
			fmt.Println(err)
			break
		}

	}

	file.WriteString(strings.TrimSpace("]}"))
}
