package main

import (

	"fmt"
	"net/http"
	"io/ioutil"
	"golang.org/x/net/html"
	"bytes"
)

func main() {

	url := "http://macaulaylibrary.org/search?&asset_format_id=1000&collection_type_id=1&layout=1&sort=21&page=1"
	resp, _ := http.Get(url)

	res_bytes, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("HTML:\n\n", string(res_bytes))

	b := resp.Body

	b.Close()
	
	r := bytes.NewReader(res_bytes)

	z := html.NewTokenizer(r)
	fmt.Println(z)

	for{
		tt := z.Next()

		switch { 
		case tt == html.ErrorToken:
			// end of the document, we're done
			return

		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if isAnchor { 
				fmt.Println("found a link~")

				for _, a := range t.Attr{

					if a.Key == "href"{

					fmt.Println("Font href: ", a.Val)
					break

					}
				}
			}
		}

	}

}
