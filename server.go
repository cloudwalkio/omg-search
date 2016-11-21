package main

import (
    "github.com/codegangsta/martini"
    "github.com/PuerkitoBio/gocrawl"
    "github.com/PuerkitoBio/goquery"
    "github.com/cloudwalkio/go-ir"
    "net/http"
    "fmt"
    "time"
    "regexp"
    "encoding/json"
    "flag"
    "os"
)

var token = flag.String("token", "", "access_token that must be validated.")

// Search has structures to keep the search results
type Search struct {
    Results []SearchResult `json:"results"`
}

// SearchResult is a container for our search responses
type SearchResult struct {
    URL string          `json:"url"`
    Title string        `json:"title"`
    Description string  `json:"description"`
}

// CrawlerData is used to aggregate structures to the crawler
type CrawlerData struct {
    Engine *ir.Engine // The Information Retrieval Engine
    Description map[string] string
    Title map[string] string
    filter *regexp.Regexp // Filtering regex
    domain *regexp.Regexp // String to remove from maps
    rootURL string
}


// MessageReturn is just a wrapper for a string
type MessageReturn struct {
    Message string `json:"message"`
}

// IRExtender is based on the gocrawl-provided DefaultExtender,
// because we don't want/need to override all methods.
type IRExtender struct {
    gocrawl.DefaultExtender // Will use the default implementation of all but Visit() and Filter()
    Data *CrawlerData // Extra data for the crawler
}

// Visit overrides original Visit for our need.
func (extender *IRExtender) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {

        url := extender.Data.domain.ReplaceAllString(ctx.URL().String(), "")

        // This is Cloudwalk specific: the div<class="span9"> contains the main content of each page
        body, err := doc.Find("div[class=\"span9\"]").Html()

        if err != nil {
            fmt.Printf("[%s] div[class=\"span9\"] not found: %s\n", url, err)
            return nil, false
        }

        // Add the html page to the engine. The ID is url, the text is body
        extender.Data.Engine.AddDocument(url, body)

        // Get the page description
        desc, find := doc.Find("meta[name=\"docs:description\"]").Attr("content")
        if !find {
            fmt.Printf("[%s] meta[name=\"docs:description\"] not found: %t\n", url, find)
            desc = "Without description"
        }

        // Add the description to the map
        extender.Data.Description[url] = desc

        // Add the title to the map
        title, _ := doc.Find("div[class=\"span9\"]").Find("h1").Html()
        extender.Data.Title[url] = title

        // Return nil and true - let gocrawl find the links
        return nil, true
}

// Filter overrides original Filter for our need.
func (extender *IRExtender) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
  return !isVisited && extender.Data.filter.MatchString(ctx.NormalizedURL().String())
}

// Crawl runs the crawler, starting at rootURL.
// Return an Information Retrieval Engine that uses the crawled documents.
func Crawl(crawlerData *CrawlerData) {
        // Create a new Extender with the above engine and regex
        ext := IRExtender{Data:crawlerData}

        // Create a new Options struct with the above Extender
        opts := gocrawl.NewOptions(&ext)
        opts.CrawlDelay = 0 * time.Second
        opts.LogFlags = gocrawl.LogInfo
        opts.MaxVisits = 500 // Will halt before that, after visiting all pages

        // Create a new crawler with the above Options
        c := gocrawl.NewCrawlerWithOptions(opts)

        // Run the crawler
        c.Run(crawlerData.rootURL)

        // Vectorize the now populated engine
        crawlerData.Engine.Vectorize()
}

func main() {
    flag.Parse()

    m := martini.Classic()

    // Our Engine uses this Regex to remove everything but: words (a-z), numbers (0-9) and
    // unicode portuguse signed characters, keeping intact words like "princípio", "autômato" e "criação"
    patternToRemove := regexp.MustCompile("[^\\w\\d\\-\\xE0\\xE1\\xE2\\xE3\\xE4\\xE5\\xE6\\xE7\\xE8\\xE9\\xEA\\xEB\\xEC\\xED\\xEE\\xEF\\xF0\\xF1\\xF2\\xF3\\xF4\\xF5\\xF6\\xF7\\xF8\\xF9\\xFA\\xFB\\xFC\\xFD\\xFE\\xFF]")

    crawlerEn := CrawlerData{Engine:ir.NewEngine("en", patternToRemove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), rootURL:"https://docs.cloudwalk.io/en/introduction"}
    crawlerPtBr := CrawlerData{Engine:ir.NewEngine("pt", patternToRemove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), rootURL:"https://docs.cloudwalk.io/pt-BR/introduction"}

    // Crawl and populate the information retrieval engines
    go Crawl(&crawlerEn)
    go Crawl(&crawlerPtBr)

    // Update the engines crawling again
    m.Post("/crawl", func(w http.ResponseWriter, req *http.Request) (int,string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        accessToken := req.FormValue("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        patternToRemove := regexp.MustCompile("[^\\w\\d\\.\\-\\xE0\\xE1\\xE2\\xE3\\xE4\\xE5\\xE6\\xE7\\xE8\\xE9\\xEA\\xEB\\xEC\\xED\\xEE\\xEF\\xF0\\xF1\\xF2\\xF3\\xF4\\xF5\\xF6\\xF7\\xF8\\xF9\\xFA\\xFB\\xFC\\xFD\\xFE\\xFF]")

        // Auxiliaries crawlers
        crawlerEnAux := CrawlerData{Engine:ir.NewEngine("en", patternToRemove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), rootURL:"https://docs.cloudwalk.io/en/introduction"}
        crawlerPtBrAux := CrawlerData{Engine:ir.NewEngine("pt", patternToRemove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), rootURL:"https://docs.cloudwalk.io/pt-BR/introduction"}

        go Crawl(&crawlerEnAux)
        go Crawl(&crawlerPtBrAux)

        // Copy the new crawlers to the global crawlers
        crawlerEn = crawlerEnAux
        crawlerPtBr = crawlerPtBrAux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Update the english engine
    m.Post("/crawl/en", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        accessToken := req.FormValue("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        crawlerEnAux := CrawlerData{Engine:ir.NewEngine("en"), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), rootURL:"https://docs.cloudwalk.io/en/introduction"}
        go Crawl(&crawlerEnAux)
        crawlerEn = crawlerEnAux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Update the pt engine
    m.Post("/crawl/pt-BR", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        accessToken := req.FormValue("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        crawlerPtBrAux := CrawlerData{Engine:ir.NewEngine("pt"), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), rootURL:"https://docs.cloudwalk.io/pt-BR/introduction"}
        go Crawl(&crawlerPtBrAux)
        crawlerPtBr = crawlerPtBrAux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Return a Json of the engines
    m.Get("/engine", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        accessToken := req.URL.Query().Get("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, "[\n" + string(crawlerPtBr.Engine.Json()) + ",\n" + string(crawlerEn.Engine.Json()) + "\n]"
    })
    // Return a Json of the pt engine
    m.Get("/engine/pt-BR", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        accessToken := req.URL.Query().Get("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, string(crawlerPtBr.Engine.Json())
    })

    // Return a Json of the en engine
    m.Get("/engine/en", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        accessToken := req.URL.Query().Get("access_token")
        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, string(crawlerEn.Engine.Json())
    })

    // Do a search: /en?query=searching+for+this
    m.Get("/:search", func(params martini.Params, w http.ResponseWriter, req *http.Request) (int, string)  {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")

        s := make([]SearchResult, 0)

        // Parse the url to get the query paramenter named "query" and convert to int
        query := req.URL.Query().Get("query")

        // Get the access token
        accessToken := req.URL.Query().Get("access_token")

        if accessToken != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        if params["search"] == "en" {
            for _, v := range crawlerEn.Engine.Query(query) {
                s = append(s, SearchResult{URL:v.Id, Title:crawlerEn.Title[v.Id], Description:crawlerEn.Description[v.Id]})
            }
        } else {
            if params["search"] == "pt-BR" {
                for _, v := range crawlerPtBr.Engine.Query(query) {
                   s = append(s, SearchResult{URL:v.Id, Title:crawlerPtBr.Title[v.Id], Description:crawlerPtBr.Description[v.Id]})
                }
            }
        }
        b,_ := json.MarshalIndent(Search{s}, "", "  ")
        return http.StatusOK, string(b)
    })

    port := os.Getenv("OMG_SEARCH_PORT")
    if port != "" {
        fmt.Printf("[martini] Listening on port " + port + "\n")
		err := http.ListenAndServe("0.0.0.0:" + port, m)
        if err != nil {
            fmt.Printf("Error: %s", err)
        }
    } else {
        fmt.Printf("Environment variable OMG_SEARCH_PORT not found \n")
    }
}
