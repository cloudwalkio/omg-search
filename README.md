omg-search
==========

We use [gocrawl] to crawl our documentation pages and our [go-ir] package to create a Information Retrieval system, aka, a search system.

Usage
-----

The API accepts the following GET requests:

  * `/en?query=search+me`: search for "search me" in the english documentation.
  * `/pt-BR?query=pesquise+me`: search for "pesquise-me" in the portuguese documentation.
  * `/crawl`: recrawl the english and portugues pages and update the Information Retrieval engine.
    * `/crawl/en`: recrawl only the english pages and update the Information Retrieval engine.
    * `/crawl/pt-BR`: recrawl only the portuguese pages and update the Information Retrieval engine.

[go-ir]:https://github.com/allanino/go-ir
[gocrawl]:https://github.com/PuerkitoBio/gocrawl
