package mrbiceps

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ProductsData struct {
	httpClient *http.Client
	hrefs []string
}

type Product struct {
	Name  string
	Price string
	Value string
}

func (p ProductsData) Length() int {
	return len(p.hrefs)
}

func (p ProductsData) Each(f func(int, Product)) error { // make smth like worker threads
	reg := regexp.MustCompile(`\d{2,}`)

	for i, l := range p.hrefs {
		req, err := http.NewRequest("GET", l, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", getRandomUserAgent())

		res, err := p.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK { // check for RateLImited
			return fmt.Errorf("bad status code");
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return err
		}


		table := doc.Find("#tab_2 > table")
		if table.Length() == 0 {
			continue
		}

		rows := table.First().Find("tr")
		if rows.Length() == 0 {
			return fmt.Errorf("outdated");
		}

		cols := rows.First().Children();

		idx := -1;

		total := 0
		fraction := 0

		for i := range cols.Length() {
			if match := reg.FindString(cols.Eq(i).Text()); match != "" {
				idx = i
				total , _ = strconv.Atoi(match)
				break;
			}
		}

		if idx == -1 {
			return fmt.Errorf("outdated");
		}

		for i = 1; i < rows.Length(); i++ {
			cols := rows.Eq(i).Find("td")
			if (strings.Contains(strings.TrimSpace(cols.Eq(0).Text()), "Baltym")) {
				if match := reg.FindString(cols.Eq(idx).Text()); match != "" {
					fraction, _ = strconv.Atoi(match)
				} else {
					return fmt.Errorf("outdated");
				}
				break;
			}
		}

		fmt.Println("total:", total, "fraction", fraction)

		f(i, Product{
			Name: doc.Find("div.summary_wrp > h1").Text(),
			Price: strings.TrimSpace(doc.Find("div.current_price").Text()),
			Value: "",
		})
	}
	return nil
}

func GetProductsData() (ProductsData, error) {
	client := &http.Client{}

	link := "https://www.mrbiceps.lt/maisto-papildai/papildai/baltymai-proteinas/";
	page := "";

	hrefs := []string{}

	for {
		req, err := http.NewRequest("GET", link + page, nil)
		if err != nil {
			return ProductsData{}, err
		}

		req.Header.Set("User-Agent", getRandomUserAgent())

		res, err := client.Do(req)
		if err != nil {
			return ProductsData{}, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK { // check for RateLImited
			return ProductsData{}, fmt.Errorf("bad status code");
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return ProductsData{}, err
		}

		products := doc.Find("div.product_element")
		
		if products.Length() == 0 {
			return ProductsData{}, fmt.Errorf("outdated");
		}

		flag := false
		links := []string{}

		for i := products.Length() - 1; i >= 0; i-- {
			s := products.Eq(i)
			if !flag && s.HasClass("no_stock") == false {
				links = make([]string, i + 1)
				flag = true
			}

			if flag {
				l, ok := s.Find("a").Attr("href");
				if !ok {
					return ProductsData{}, fmt.Errorf("outdated");
				}

				links[i] = l
			}
		}

		hrefs = append(hrefs, links...)

		if (products.Length() != len(links)) {
			break;
		}

		s := doc.Find("a.pagination_link")
		if (s.Length() != 2) {
			return ProductsData{}, fmt.Errorf("outdated");
		}

		p, ok := s.Eq(1).Attr("href");
		page = p

		if !ok {
			return ProductsData{}, fmt.Errorf("outdated");
		}

		if (page == "#") {
			break;
		}
	}

	return ProductsData{
		httpClient: client,
		hrefs: hrefs,
	}, nil
}

func getRandomUserAgent() string {
    userAgents := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:119.0) Gecko/20100101 Firefox/119.0",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0) AppleWebKit/537.36 (KHTML, like Gecko) Safari/605.1.15",
        "Mozilla/5.0 (iPhone; CPU iPhone OS 15_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.2 Mobile/15E148 Safari/604.1",
        "Mozilla/5.0 (Android 11; Mobile; rv:109.0) Gecko/109.0 Firefox/109.0",
        "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko",
        "Opera/9.80 (Windows NT 6.0) Presto/2.12.388 Version/12.14",
        "Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Mobile Safari/537.36",
    }

		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
    return userAgents[rand.Intn(len(userAgents))]
}
