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

var urlsArray []string

var macaulayUrl = "http://media2.macaulaylibrary.org/Audio/Audio1/"

var searchUrl = "http://macaulaylibrary.org/search?&asset_format_id=1000&collection_type_id=1&layout=1&sort=21&page="

const t = `{{range $i, $v := .}}{{if $i}} or {{$v}}{{else}}{{$v}}{{end}}{{end}}`

func main() {

	var wg sync.WaitGroup

	numOfPagesToVisit := 1 // change this to know how many pages to visit max seems to be 1729

	for i := 1; i < numOfPagesToVisit+1; i++ {
		fmt.Println("page number: ", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			getUrlsFromPage(i)
			fmt.Println("size of array: ", len(urlsArray))
		}()

		if i%10 == 0 {
			time.Sleep(30000 * time.Millisecond)
		}

	}
	wg.Wait()
	fmt.Println(len(urlsArray))

	for _, url := range urlsArray {
		fmt.Println(macaulayUrl + url[7:9] + url[6:]) // get rid of "/audio" (first 6 characters) and get next two
	}

	writeUrlsToFile()
	fmt.Println("done~")
}

func getUrlsFromPage(pageNum int) {

	url := searchUrl + strconv.Itoa(pageNum)
	resp, _ := http.Get(url)
	res_bytes, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("HTML:\n\n", string(res_bytes))
	b := resp.Body

	if b == nil {
		return
	}

	b.Close()
	r := bytes.NewReader(res_bytes)
	z := html.NewTokenizer(r)
	//fmt.Println(z)
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
							//fmt.Println("URL: ", a.Val)
							urlsArray = append(urlsArray, a.Val)
							break
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

	for i := 0; i < len(urlsArray); i++ {

		commaBracket := "\"},"
		if i == len(urlsArray)-1 {

			commaBracket = "\"}"

		}
		_, err := file.WriteString(strings.TrimSpace("{\"url\":\""+
			macaulayUrl+
			urlsArray[i][7:9]+
			urlsArray[i][6:]) +
			commaBracket +
			"\n")

		if err != nil {
			fmt.Println(err)
			break
		}

	}

	file.WriteString(strings.TrimSpace("]}"))
}
