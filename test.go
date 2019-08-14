package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func getResponse(url string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Close = true
	if err != nil {
		fmt.Println(err)
	}
	return client.Do(req)
}

type Request struct {
	title    string
	url      string
	selector string
	each     func(i int, s *goquery.Selection) GatherItem
}

type GatherItem struct {
	Index int
	Title string
	Url   string
}

type GatherCache struct {
	t     time.Time
	title string
	items []GatherItem
}

func (p Request) fetch() ([]GatherItem, error) {
	res, err := getResponse(p.url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	root, err := html.Parse(res.Body)
	if err != nil {
		fmt.Println("aaaa", err)
	}

	doc := goquery.NewDocumentFromNode(root)

	result := []GatherItem{}
	doc.Find(p.selector).Each(func(i int, s *goquery.Selection) {
		GatherItem := p.each(i, s)
		if len(GatherItem.Title) > 0 && len(GatherItem.Url) > 0 {
			if strings.Index(GatherItem.Url, "http") != 0 {
				u, err := url.Parse(p.url)
				if err == nil {
					if len(u.Scheme) > 0 {
						GatherItem.Url = u.Scheme + "://" + u.Host + GatherItem.Url
					} else {
						GatherItem.Url = u.Host + GatherItem.Url
					}
				}
			}
			result = append(result, GatherItem)
		}
	})
	return result, nil
}

type Gather struct {
	request []*Request
	cache   map[string]GatherCache
}

func (p *Gather) Init() {
	if p.cache == nil {
		p.cache = make(map[string]GatherCache)
	}

	p.Add(&Request{
		title:    "csdn",
		url:      "https://blog.csdn.net",
		selector: ".right_box .feed_company li",
		each: func(i int, s *goquery.Selection) GatherItem {
			a := s.Find(".content>h3>a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "博客园-今日热门",
		url:      "https://news.cnblogs.com/n/digg?type=today",
		selector: "#news_list .news_block",
		each: func(i int, s *goquery.Selection) GatherItem {
			fmt.Println(i)
			a := s.Find(".content>.news_entry>a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "酷壳-最新文章",
		url:      "https://coolshell.cn/",
		selector: "#recent-posts-2 ul li",
		each: func(i int, s *goquery.Selection) GatherItem {
			fmt.Println(i)
			a := s.Find("a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "阮一峰的网络日志-近期文章",
		url:      "http://www.ruanyifeng.com/blog/archives.html",
		selector: ".module-content ul.module-list li",
		each: func(i int, s *goquery.Selection) GatherItem {
			fmt.Println(i)
			a := s.Find("a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "廖雪峰的官方网站-最新发表",
		url:      "https://www.liaoxuefeng.com/",
		selector: "#x-content .uk-margin.uk-clearfix",
		each: func(i int, s *goquery.Selection) GatherItem {
			a := s.Find("a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "码农网",
		url:      "http://www.codeceo.com/",
		selector: ".central .content .excerpt",
		each: func(i int, s *goquery.Selection) GatherItem {
			a := s.Find("h3>a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})

	p.Add(&Request{
		title:    "ITeye",
		url:      "https://www.iteye.com/",
		selector: "#main .after .main_left ul li",
		each: func(i int, s *goquery.Selection) GatherItem {
			a := s.Find("a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			if len(url) > 0 && !strings.Contains(url, "http") {
				return GatherItem{i, title, url}
			} else {
				return GatherItem{}
			}
		},
	})

	p.Add(&Request{
		title:    "小程序资讯",
		url:      "https://www.newrank.cn/public/news.html",
		selector: ".media-main-left-news-list",
		each: func(i int, s *goquery.Selection) GatherItem {
			fmt.Println(i)
			a := s.Find("a")
			title := strings.Trim(a.Text(), " ")
			url, _ := a.Attr("href")
			return GatherItem{i, title, url}
		},
	})
}

func (p *Gather) Add(req *Request) {
	p.request = append(p.request, req)
}

func (p *Gather) Get(title string) []GatherItem {
	cache, ok := p.cache[title]
	if ok {
		if time.Since(cache.t) < (1 * time.Minute) {
			fmt.Println("xxxxxxxxxxxxxxxxxxx")
			return cache.items
		}
	}

	for _, req := range p.request {
		if req.title == title {
			result, err := req.fetch()
			if err != nil {
				fmt.Println(err)
			}
			if len(result) > 0 {
				p.cache[title] = GatherCache{time.Now(), title, result}
			}
			return result
		}
	}
	return []GatherItem{}
}

func (p *Gather) List() []string {
	ls := []string{}
	for _, req := range p.request {
		ls = append(ls, req.title)
	}
	return ls
}

func main() {
	gather := new(Gather)
	gather.Init()
	fmt.Println(gather.Get("csdn"))
}
