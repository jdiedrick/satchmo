package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// https://www.golang-book.com/books/intro/9
type Recording struct {
	catalogNumber     int
	speciesCommon     string
	speciesScientific string
	soundType         string
	location          string
	recordist         string
	date              string
	length            string
	quality           int
	url               string
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
	resp, _ := http.Get(url)
	res_bytes, _ := ioutil.ReadAll(resp.Body)
	b := resp.Body

	// so: http://stackoverflow.com/questions/26726203/runtime-error-invalid-memory-address-or-nil-pointer-dereference/26738639#comment42044181_26726203
	if b == nil {
		return
	}

	b.Close()
	r := bytes.NewReader(res_bytes)
	z := html.NewTokenizer(r)
	recordingNum := 0
loop:

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			//fmt.Println("Len urls %i: ", len(urls))
			//fmt.Println(urls)
			// end of the document, we're done
			break loop

		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				//fmt.Println("found a link~")
				for _, a := range t.Attr {
					if a.Key == "href" {
						if len(a.Val) == 13 && a.Val[:7] == "/audio/" {
							fmt.Println("recordings size: ", len(recordings))
							fmt.Println("URL: ", a.Val)
							fmt.Print("Recording num: ", recordingNum)
							recording := Recording{url: a.Val}
							recordings = append(recordings, recording)

						}
					}
				}
			}

			isH4 := t.Data == "h4"

			if isH4 {
				for _, h4 := range t.Attr {
					if h4.Key == "class" {
						if h4.Val == "indent" {
							z.Next()
							t := z.Token()
							fmt.Println(t.Data)
							recordings[recordingNum].speciesCommon = string(t.Data)
							recordingNum++

						}
					}
				}
			}

		}
	}
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
			"{\"url\":\"" +
				macaulayUrl +
				recordings[i].url[7:9] +
				recordings[i].url[6:] +
				"\"" +
				", \"speciesCommon\":\"" +
				strings.TrimSpace(recordings[i].speciesCommon) +
				commaBracket +
				"\n"))

		if err != nil {
			fmt.Println(err)
			break
		}

	}

	file.WriteString(strings.TrimSpace("]}"))
}
