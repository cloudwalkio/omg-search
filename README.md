omg-search
==========

We use [gocrawl] to crawl our documentation pages and our [go-ir] package to create a Information Retrieval system, aka, a search system.

Usage
-----

In starting the server we must pass a token that will be used to authorization in any requests to the API: 
```
docker run -d -p 80:5000 allanino/search -token "my_token"
```

The API accepts the following GET requests:

  * `/en?query=search+me&access-token=my_token`: search for "search me" in the english documentation.
  * `/pt-BR?query=pesquise+me&access-token=my_token`: search for "pesquise-me" in the portuguese documentation.
  * `/crawl?access-token=my_token`: recrawl the english and portugues pages and update the Information Retrieval engine.
    * `/crawl/en?access-token=my_token`: recrawl only the english pages and update the Information Retrieval engine.
    * `/crawl/pt-BR?access-token=my_token`: recrawl only the portuguese pages and update the Information Retrieval engine.

[go-ir]:https://github.com/allanino/go-ir
[gocrawl]:https://github.com/PuerkitoBio/gocrawl
