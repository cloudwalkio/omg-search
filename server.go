package main

import (
    "github.com/codegangsta/martini"
    "github.com/PuerkitoBio/gocrawl"
    "github.com/PuerkitoBio/goquery"
    "github.com/allanino/go-ir"
    "net/http"
    "fmt"
    "time"
    "regexp"
    "encoding/json"
)

// Structures to keep the search results
type Search struct {
    Results []SearchResult
}

type SearchResult struct {
    Url string
    Description string
}

// Used to aggregate structures to the crawler
type CrawlerData struct {
    Engine *ir.Engine // The Information Retrieval Engine
    Description map[string] string 
    filter *regexp.Regexp // Filtering regex
    domain *regexp.Regexp // String to remove from maps
    root_url string
}

// Create the Extender implementation, based on the gocrawl-provided DefaultExtender,
// because we don't want/need to override all methods.
type IRExtender struct {
    gocrawl.DefaultExtender // Will use the default implementation of all but Visit() and Filter()
    Data *CrawlerData // Extra data for the crawler
}

// Override Visit for our need.
func (this *IRExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {

        url := this.Data.domain.ReplaceAllString(ctx.URL().String(), "")

        // This is Cloudwalk specific: the div<class="span9"> contains the main content of each page 
        body, err := doc.Find("div[class=\"span9\"]").Html()

        if err != nil {
            fmt.Printf("[%s] div[class=\"span9\"] not found: %s\n", url, err)
            return nil, false
        }

        // Add the html page to the engine. The ID is url, the text is body
        this.Data.Engine.AddDocument(url, body)

        desc, find := doc.Find("meta[name=\"docs:description\"]").Attr("content")
        if !find {
            fmt.Printf("[%s] meta[name=\"docs:description\"] not found: %s\n", url, find)
            desc = "Without description"
        }

        // Add the description to the map
        this.Data.Description[url] = desc

        // Return nil and true - let gocrawl find the links
        return nil, true
}

// Override Filter for our need.
func (this *IRExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
  return !isVisited && this.Data.filter.MatchString(ctx.NormalizedURL().String())
}

// Run the crawler, starting at root_url.
// Return an Information Retrieval Engine that uses the crawled documents.
func Crawl(crawler_data *CrawlerData) {


        // Create a new Extender with the above engine and regex
        ext := IRExtender{Data:crawler_data}
        
        // Create a new Options struct with the above Extender
        opts := gocrawl.NewOptions(&ext)
        opts.CrawlDelay = 0 * time.Second
        opts.LogFlags = gocrawl.LogInfo
        opts.MaxVisits = 2

        // Create a new crawler with the above Options
        c := gocrawl.NewCrawlerWithOptions(opts)
        
        // Run the crawler
        c.Run(crawler_data.root_url)
        
        // Vectorize the now populated engine
        crawler_data.Engine.Vectorize()
}

// Rerun the crawler to recalculate the engine vectorization
func Reindex(crawler_data *CrawlerData) {
    Crawl(crawler_data)
}

func main() {

    m := martini.Classic()

    crawler_en := CrawlerData{Engine:ir.NewEngine(), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(en)`), root_url:"https://docs.cloudwalk.io/en/introduction"}
    crawler_pt_br := CrawlerData{Engine:ir.NewEngine(), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(pt-BR)`), root_url:"https://docs.cloudwalk.io/pt-BR/introduction"}

    // Index the engines
    go Reindex(&crawler_en)
    go Reindex(&crawler_pt_br)

    // Update the engines crawling again
    m.Get("/reindex", func() {
    go Reindex(&crawler_en)
    go Reindex(&crawler_pt_br)
    })

    // Update the english engine
    m.Get("/reindex/en", func() {
        go Reindex(&crawler_en)
    })

    // Update the pt engine
    m.Get("/reindex/pt-BR", func() {
        go Reindex(&crawler_pt_br)
    })

    // Return a Json of the engines
    m.Get("/engine", func(w http.ResponseWriter) string {
        w.Header().Set("Content-Type", "application/json")
        return "[\n" + string(crawler_pt_br.Engine.Json()) + ",\n" + string(crawler_en.Engine.Json()) + "\n]"
    })
    // Return a Json of the pt engine
    m.Get("/engine/pt-BR", func(w http.ResponseWriter) string {
        w.Header().Set("Content-Type", "application/json")
        return string(crawler_pt_br.Engine.Json())
    })

    // Return a Json of the en engine
    m.Get("/engine/en", func(w http.ResponseWriter) string {
        w.Header().Set("Content-Type", "application/json")
        return string(crawler_en.Engine.Json())
    })


    // Do a search: /en?query=searching+for+this
    m.Get("/:search", func(params martini.Params, w http.ResponseWriter, req *http.Request) string  {
        w.Header().Set("Content-Type", "application/json")

        s := make([]SearchResult, 0)
        
        // Parse the url to get the query paramenter named "query" and convert to int
        query := req.URL.Query().Get("query")

        if params["search"] == "en" {
            for _, v := range crawler_en.Engine.Query(query) {
                s = append(s, SearchResult{Url:v.Id, Description:crawler_en.Description[v.Id]})
            }
        } else {
            if params["search"] == "pt-BR" {
                for _, v := range crawler_pt_br.Engine.Query(query) {
                   s = append(s, SearchResult{Url:v.Id, Description:crawler_pt_br.Description[v.Id]})
                }
            } 
        }
        b,_ := json.MarshalIndent(Search{s}, "", "  ")
        return string(b)
    })

    fmt.Printf("[martini] Listening on port 5000\n")
    http.ListenAndServe("0.0.0.0:5000", m)
}