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

// Structures to keep the search results
type Search struct {
    Results []SearchResult `json:"results"`
}

type SearchResult struct {
    Url string          `json:"url"`
    Title string        `json:"title"`
    Description string  `json:"description"`
}

// Used to aggregate structures to the crawler
type CrawlerData struct {
    Engine *ir.Engine // The Information Retrieval Engine
    Description map[string] string
    Title map[string] string
    filter *regexp.Regexp // Filtering regex
    domain *regexp.Regexp // String to remove from maps
    root_url string
}

type MessageReturn struct {
    Message string `json:"message"`
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

        // Get the page description
        desc, find := doc.Find("meta[name=\"docs:description\"]").Attr("content")
        if !find {
            fmt.Printf("[%s] meta[name=\"docs:description\"] not found: %s\n", url, find)
            desc = "Without description"
        }

        // Add the description to the map
        this.Data.Description[url] = desc

        // Add the title to the map
        title, _ := doc.Find("div[class=\"span9\"]").Find("h1").Html()
        this.Data.Title[url] = title

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
        opts.MaxVisits = 500 // Will halt before that, after visiting all pages

        // Create a new crawler with the above Options
        c := gocrawl.NewCrawlerWithOptions(opts)

        // Run the crawler
        c.Run(crawler_data.root_url)

        // Vectorize the now populated engine
        crawler_data.Engine.Vectorize()
}

func main() {
    flag.Parse()

    m := martini.Classic()

    // Our Engine uses this Regex to remove everything but: words (a-z), numbers (0-9) and
    // unicode portuguse signed characters, keeping intact words like "princípio", "autômato" e "criação"
    pattern_to_remove := regexp.MustCompile("[^\\w\\d\\-\\xE0\\xE1\\xE2\\xE3\\xE4\\xE5\\xE6\\xE7\\xE8\\xE9\\xEA\\xEB\\xEC\\xED\\xEE\\xEF\\xF0\\xF1\\xF2\\xF3\\xF4\\xF5\\xF6\\xF7\\xF8\\xF9\\xFA\\xFB\\xFC\\xFD\\xFE\\xFF]")

    crawler_en := CrawlerData{Engine:ir.NewEngine("en", pattern_to_remove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), root_url:"https://docs.cloudwalk.io/en/introduction"}
    crawler_pt_br := CrawlerData{Engine:ir.NewEngine("pt", pattern_to_remove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), root_url:"https://docs.cloudwalk.io/pt-BR/introduction"}

    // Crawl and populate the information retrieval engines
    go Crawl(&crawler_en)
    go Crawl(&crawler_pt_br)

    // Update the engines crawling again
    m.Post("/crawl", func(w http.ResponseWriter, req *http.Request) (int,string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        access_token := req.FormValue("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        pattern_to_remove := regexp.MustCompile("[^\\w\\d\\.\\-\\xE0\\xE1\\xE2\\xE3\\xE4\\xE5\\xE6\\xE7\\xE8\\xE9\\xEA\\xEB\\xEC\\xED\\xEE\\xEF\\xF0\\xF1\\xF2\\xF3\\xF4\\xF5\\xF6\\xF7\\xF8\\xF9\\xFA\\xFB\\xFC\\xFD\\xFE\\xFF]")

        // Auxiliaries crawlers
        crawler_en_aux := CrawlerData{Engine:ir.NewEngine("en", pattern_to_remove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), root_url:"https://docs.cloudwalk.io/en/introduction"}
        crawler_pt_br_aux := CrawlerData{Engine:ir.NewEngine("pt", pattern_to_remove), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), root_url:"https://docs.cloudwalk.io/pt-BR/introduction"}

        go Crawl(&crawler_en_aux)
        go Crawl(&crawler_pt_br_aux)

        // Copy the new crawlers to the global crawlers
        crawler_en = crawler_en_aux
        crawler_pt_br = crawler_pt_br_aux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Update the english engine
    m.Post("/crawl/en", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        access_token := req.FormValue("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        crawler_en_aux := CrawlerData{Engine:ir.NewEngine("en"), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/en(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/en)`), root_url:"https://docs.cloudwalk.io/en/introduction"}
        go Crawl(&crawler_en_aux)
        crawler_en = crawler_en_aux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Update the pt engine
    m.Post("/crawl/pt-BR", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Access-Control-Allow-Origin", "*")

        // Get the access token
        access_token := req.FormValue("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        crawler_pt_br_aux := CrawlerData{Engine:ir.NewEngine("pt"), Title:make(map[string] string), Description:make(map[string] string), filter:regexp.MustCompile(`http(s*)://docs\.cloudwalk\.io/pt-BR(.*)`), domain:regexp.MustCompile(`.*(docs.cloudwalk.io/pt-BR)`), root_url:"https://docs.cloudwalk.io/pt-BR/introduction"}
        go Crawl(&crawler_pt_br_aux)
        crawler_pt_br = crawler_pt_br_aux

        b,_ := json.MarshalIndent(MessageReturn{"Crawling web pages"}, "", "  ")
        return http.StatusOK, string(b)
    })

    // Return a Json of the engines
    m.Get("/engine", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        access_token := req.URL.Query().Get("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, "[\n" + string(crawler_pt_br.Engine.Json()) + ",\n" + string(crawler_en.Engine.Json()) + "\n]"
    })
    // Return a Json of the pt engine
    m.Get("/engine/pt-BR", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        access_token := req.URL.Query().Get("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, string(crawler_pt_br.Engine.Json())
    })

    // Return a Json of the en engine
    m.Get("/engine/en", func(w http.ResponseWriter, req *http.Request) (int, string) {
        w.Header().Set("Content-Type", "application/json")

        // Get the access token
        access_token := req.URL.Query().Get("access_token")
        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        return http.StatusOK, string(crawler_en.Engine.Json())
    })

    // Do a search: /en?query=searching+for+this
    m.Get("/:search", func(params martini.Params, w http.ResponseWriter, req *http.Request) (int, string)  {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")

        s := make([]SearchResult, 0)

        // Parse the url to get the query paramenter named "query" and convert to int
        query := req.URL.Query().Get("query")

        // Get the access token
        access_token := req.URL.Query().Get("access_token")

        if access_token != *token {
            b,_ := json.MarshalIndent(MessageReturn{"Not authorized"}, "", "  ")
            return http.StatusUnauthorized , string(b)
        }

        if params["search"] == "en" {
            for _, v := range crawler_en.Engine.Query(query) {
                s = append(s, SearchResult{Url:v.Id, Title:crawler_en.Title[v.Id], Description:crawler_en.Description[v.Id]})
            }
        } else {
            if params["search"] == "pt-BR" {
                for _, v := range crawler_pt_br.Engine.Query(query) {
                   s = append(s, SearchResult{Url:v.Id, Title:crawler_pt_br.Title[v.Id], Description:crawler_pt_br.Description[v.Id]})
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
