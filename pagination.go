package canvas

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
)

type pageInitFunction func(int, io.Reader) ([]interface{}, error)

func newPaginatedList(
	d doer,
	path string,
	init pageInitFunction,
	parameters []Option,
) *paginated {
	if parameters == nil {
		parameters = []Option{}
	}
	return &paginated{
		do:      d,
		path:    path,
		query:   asParams(parameters),
		init:    init,
		perpage: 10,
		wg:      new(sync.WaitGroup),
		objects: make(chan interface{}),
		errs:    make(chan error),
	}
}

type paginated struct {
	path  string
	query params
	do    doer

	n       int
	perpage int
	objects chan interface{}
	errs    chan error

	wg   *sync.WaitGroup
	init pageInitFunction
}

// returns <number of pages>, <first response
func (p *paginated) firstReq() (int, *http.Response, error) {
	q := params{"page": {"1"}, "per_page": {fmt.Sprintf("%d", p.perpage)}}
	q.Join(p.query)
	resp, err := get(p.do, p.path, q)
	if err != nil {
		return 0, nil, err
	}
	pages, err := newLinkedResource(resp.Header)
	if err != nil {
		return 0, nil, err
	}
	lastpage, ok := pages.links["last"]
	if !ok {
		return 0, nil, errors.New("could not find last page")
	}
	p.n = lastpage.page
	return p.n, resp, nil
}

func (p *paginated) channel() <-chan interface{} {
	n, resp, err := p.firstReq() // n pages and first request
	if err != nil {
		p.errs <- err
		close(p.errs)
		close(p.objects)
		return nil
	}
	p.wg.Add(n)

	go func() {
		defer resp.Body.Close()
		defer p.wg.Done()
		list, err := p.init(1, resp.Body)
		if err != nil {
			p.errs <- err
			return
		}
		for _, o := range list {
			p.objects <- o
		}
	}()
	for page := 2; page <= n; page++ {
		go func(page int64, path string) {
			defer p.wg.Done()
			q := params{
				"page":     {strconv.FormatInt(page, 10)},
				"per_page": {fmt.Sprintf("%d", p.perpage)}}
			q.Join(p.query)
			resp, err := get(p.do, path, q)
			if err != nil {
				p.errs <- err
				return
			}
			defer resp.Body.Close()
			obs, err := p.init(int(page), resp.Body)
			if err != nil {
				p.errs <- err
				return
			}
			for _, o := range obs {
				p.objects <- o
			}
		}(int64(page), p.path)
	}
	go func() {
		p.wg.Wait()
		close(p.objects)
		close(p.errs)
	}()
	return p.objects
}

func (p *paginated) collect() ([]interface{}, error) {
	p.channel()
	collection := make([]interface{}, 0, p.n*p.perpage)
	for {
		select {
		case err := <-p.errs:
			if err != nil {
				return nil, err
			}
		case obj := <-p.objects:
			if obj == nil {
				return collection, nil
			}
			collection = append(collection, obj)
		}
	}
}

func (p *paginated) ordered() ([]interface{}, error) {
	return nil, nil
}

var resourceRegex = regexp.MustCompile(`<(.*?)>; rel="(.*?)"`)

func newLinkedResource(header http.Header) (*linkedResource, error) {
	var err error
	resource := &linkedResource{
		links: map[string]*link{},
	}
	links := header.Get("Link")
	parts := resourceRegex.FindAllStringSubmatch(links, -1)

	for _, part := range parts {
		resource.links[part[2]], err = newlink(part[1])
		if err != nil {
			return resource, err
		}
	}
	return resource, nil
}

type linkedResource struct {
	links map[string]*link
}

type link struct {
	url  *url.URL
	page int
}

func newlink(urlstr string) (*link, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	page, err := strconv.ParseInt(u.Query().Get("page"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse page num: %w", err)
	}
	return &link{
		url:  u,
		page: int(page),
	}, nil
}
