package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"os"
	"text/template"
	"encoding/json"
	"sync"
)

var goodreadsResponse GoodreadsResponse

type GoodreadsResponse struct {
    XMLName xml.Name `xml:"GoodreadsResponse"`
    Reviews []Review `xml:"reviews>review"`
		mux sync.Mutex
}

type Review struct {
		Id      string  `xml:"id"`
		Book    Book    `xml:"book"`
		Date    string  `xml:"date_added"`
}

type Book struct {
		Title   string  `xml:"title"`
		Image   string  `xml:"image_url"`
		Description   string  `xml:"description"`
		Link   string  `xml:"link"`
}

type APIKey struct {
    Key string `json:"key"`
    ID string `json:"id"`
}

var tmpls, _ = template.ParseFiles("bookshelf.html")
var tmpl = tmpls.Lookup("bookshelf.html")


func main() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	goodreadsResponse.mux.Lock()
	tmpl.Execute(w, goodreadsResponse.Reviews)
	goodreadsResponse.mux.Unlock()
	go MakeRequest(w)
}

func gohandle(w http.ResponseWriter) {

}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func cleanGoodreads(r *GoodreadsResponse) {
	for index, review := range r.Reviews {
	    // index is the index where we are
	    // element is the element from someSlice for where we are
			if (len(review.Book.Description) >= 500) {
				r.Reviews[index].Book.Description =
					review.Book.Description[:500] + "</b></i>..."
			}
				r.Reviews[index].Date = review.Date[:10] + review.Date[len(review.Date)-5:]

	}
}

func LoadConfiguration(file string) APIKey {
    var key APIKey
    keyFile, err := os.Open(file)
    defer keyFile.Close()
    if err != nil {
        fmt.Println(err.Error())
    }
    jsonParser := json.NewDecoder(keyFile)
    jsonParser.Decode(&key)
    return key
}

func MakeRequest(w http.ResponseWriter) {

	var key APIKey = LoadConfiguration("api-key.json")
	var goodreads_url string = fmt.Sprintf("https://www.goodreads.com/review/list?key=%s&v=2&id=%s?&per_page=200", key.Key, key.ID)
	resp, err := http.Get(goodreads_url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	goodreadsResponse.mux.Lock()
	goodreadsResponse.Reviews = nil
	err = xml.Unmarshal(body, &goodreadsResponse)
	if err != nil {
			fmt.Println("Fatal error ", err.Error())
			os.Exit(1)
	}

	cleanGoodreads(&goodreadsResponse);
	goodreadsResponse.mux.Unlock()
}


func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
