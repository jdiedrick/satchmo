package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"os"
	"strings"
)

var urlsArray []string

var macaulayUrl = "http://media2.macaulaylibrary.org/Audio/Audio1/"

func main() {

	var wg sync.WaitGroup
	
	numOfPagesToVisit := 5 // change this to know how many pages to visit

	for i := 1; i < numOfPagesToVisit + 1; i++ {
		fmt.Println("page number: ", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			getUrlsFromPage(i)
			fmt.Println("size of array: ", len(urlsArray))
		}()

	}
	wg.Wait()
	fmt.Println(len(urlsArray))
	
	for _, url := range urlsArray{
		fmt.Println(macaulayUrl + url[7:9] + url[6:]) // get rid of "/audio" (first 6 characters) and get next two
	}

	writeUrlsToFile()
	fmt.Println("done~")
}

func getUrlsFromPage(pageNum int) {

	url := "http://macaulaylibrary.org/search?&asset_format_id=1000&collection_type_id=1&layout=1&sort=21&page=" + strconv.Itoa(pageNum)
	resp, _ := http.Get(url)
	res_bytes, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("HTML:\n\n", string(res_bytes))
	b := resp.Body
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


func writeUrlsToFile(){
	var(
		file *os.File
		err error
	)

		
	if file, err = os.Create("audio_urls.txt"); err != nil{
		return
	}
	defer file.Close()

	for _, url := range urlsArray {

		_, err := file.WriteString(strings.TrimSpace(macaulayUrl + url[7:9] + url[6:]) + "\n");

		if err != nil{
			fmt.Println(err)
			break
		}

	}
}

