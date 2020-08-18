package crawler

import (
	"crawlerJob/model"
	"crawlerJob/telegram"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	URL = "https://www.104.com.tw/jobs/search/?keyword=golang&area=6001001000,6001008000&jobsource=2018indexpoc&ro=0&page="
)

var ch104 = make(chan bool, 1)

type c104 struct{}

func New104() Action {
	return c104{}
}

func (c104) Entry() {
	fmt.Println("Entry 104")
}

func (c104) Crawler() string {
	var page int = 1
	for {
		select {
		case <-ch104:
			fmt.Println("stop 104 crawler")
			return Crawler_Init
		default:
			crawler(page)
			page++
			time.Sleep(time.Second)
		}
	}

	// crawler(page)
}

func crawler(page int) {
	url := fmt.Sprintf("%s%d", URL, page)
	fmt.Println(url)
	stories := []jobInfo{}

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: www.104.com.tw
		colly.AllowedDomains("www.104.com.tw"),
		// Parallelism
		colly.Async(true),
	)

	c.OnHTML(".b-block--nodata", func(e *colly.HTMLElement) {
		ch104 <- true
		return
	})

	// On every a element which has .top-matter attribute call callback
	// This class is unique to the div that holds all information about a story
	c.OnHTML(".js-job-item", func(e *colly.HTMLElement) {
		tmp := jobInfo{}
		tmp.Id = e.Attr("data-job-no")
		tmp.Company = e.Attr("data-cust-name")
		tmp.Title = e.Attr("data-job-name")
		tmp.Salary = e.ChildText(".b-tag--default")
		tmp.Content = e.ChildText(".job-list-item__info")
		tmp.Link = e.ChildAttr("a", "href")
		stories = append(stories, tmp)
	})

	// Set max Parallelism and introduce a Random Delay
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	// Before making a request print "Visiting ..."
	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())

	// })

	c.Visit(url)

	c.Wait()

	for _, v := range stories {
		// fmt.Println("ID: ", v.Id)
		// fmt.Println("公司: ", v.Company)
		// fmt.Println("職缺: ", v.Title)
		// fmt.Println("薪資: ", v.Salary)
		// fmt.Println("内容: ", v.Content)
		// fmt.Println("連結: ", v.Link)
		result := model.InsertJob(v.Id, "golang", v.Company, v.Title, v.Salary, v.Content, v.Link, "104")
		if result == true {
			telegram.Send(v.String())
		}
	}
	fmt.Println(len(stories))
}

func (c104) Exit() {
	fmt.Println("Exit 104")
	time.Sleep(4 * time.Hour)
}
