package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Entry struct {
	Description     string
	Type            string
	Price           string
	ManufactureDate string
	Condition       string
	Link            string
}

type Crawler interface {
	Crawl() []Entry
}

type Boot24Crawler struct {
}

func (c Boot24Crawler) Create(s *goquery.Selection) Entry {
	link := s.AttrOr("href", "none")
	price := s.Find(".sr-price").Text()
	desc := s.Find(".sr-objektbox-us").Text()
	details := strings.Fields(s.Find(".details_left").Text())
	boatType := details[0]
	manufactureDate := details[3]
	condition := details[5]

	return Entry{
		Condition:       condition,
		Description:     desc,
		Link:            link,
		ManufactureDate: manufactureDate,
		Price:           price,
		Type:            boatType,
	}
}

func (c Boot24Crawler) GetLastPage(doc *goquery.Document) int {
	pageNumbers := []int{}
	doc.Find(".main-content-norm .sitenumber a.seite").Each(func(i int, s *goquery.Selection) {
		pageNumber, err := strconv.Atoi(s.Text())
		if err != nil {
			log.Fatal(err)
		}
		pageNumbers = append(pageNumbers, pageNumber)
	})
	return pageNumbers[len(pageNumbers)-1]
}

func (c Boot24Crawler) Values() url.Values {
	v := url.Values{}
	v.Add("typ", "t6")
	v.Add("new", "true")
	v.Add("text", "")
	v.Add("kategorie", "")
	v.Add("new_used", "")
	v.Add("material", "")
	v.Add("currency", "EUR")
	v.Add("preisMin", "")
	v.Add("preisMax", "")
	v.Add("motorenEinheit", "1.00")
	v.Add("motorenLeistungMin", "")
	v.Add("motorenLeistungMax", "")
	v.Add("baujahrVon", "")
	v.Add("baujahrBis", "")
	v.Add("laengeMin", "")
	v.Add("laengeMax", "")
	v.Add("antriebArt", "")
	v.Add("antriebWie", "")
	v.Add("liegeplatz", "")
	return v
}

func (c Boot24Crawler) QueryPage(page int) *http.Response {
	url := fmt.Sprintf("http://www.boot24.com/suchergebnis/segelboot.php?page=%d&pagesize=3", page)
	res, err := http.PostForm(url, c.Values())
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (c Boot24Crawler) ParsePage(res *http.Response) *goquery.Document {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func (c Boot24Crawler) ProcessDoc(doc *goquery.Document) []Entry {
	entries := []Entry{}
	doc.Find("#res").Each(func(i int, s *goquery.Selection) {
		s.Find(".sr-objektbox-in").Each(func(i int, s *goquery.Selection) {
			entry := c.Create(s)
			entries = append(entries, entry)
		})
	})
	return entries
}

func (c Boot24Crawler) Crawl() []Entry {
	result := []Entry{}

	// crawl first page
	currentPage := 1
	res := c.QueryPage(currentPage)
	doc := c.ParsePage(res)
	tmp := c.ProcessDoc(doc)
	result = append(result, tmp...)

	// crawl remaining pages
	currentPage += 1
	lastPage := c.GetLastPage(doc) - 1
	for currentPage < lastPage {
		log.Printf("Processing page %d of %d", currentPage, lastPage)
		res := c.QueryPage(currentPage)
		doc := c.ParsePage(res)
		tmp := c.ProcessDoc(doc)
		result = append(result, tmp...)
		currentPage += 1
		log.Println("Finished.")
	}

	return result
}

func main() {
	crawler := Boot24Crawler{}
	fmt.Println(crawler.Crawl())
}
