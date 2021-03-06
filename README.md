# omg-search

[![Docker build](http://dockeri.co/image/cloudwalk/search)](https://registry.hub.docker.com/u/cloudwalk/search/)

We use [gocrawl] to crawl our documentation pages and our [go-ir] package to create an Information Retrieval system, aka, a search system.

## Usage

In starting the server we must pass a token that will be used to authorization in any requests to the API:
```
docker run -d -p 80:5000 cloudwalk/search ./startup.sh "my_token"
```

The API accepts the following GET requests:

  * `/en?query=search+me&access_token=my_token`: search for "search me" in the english documentation.
  * `/pt-BR?query=pesquise+me&access_token=my_token`: search for "pesquise-me" in the portuguese documentation.

The API accepts the following POST requests:

  * `/crawl?access_token=my_token`: recrawl the english and portugues pages and update the Information Retrieval engine.
    * `/crawl/en?access_token=my_token`: recrawl only the english pages and update the Information Retrieval engine.
    * `/crawl/pt-BR?access_token=my_token`: recrawl only the portuguese pages and update the Information Retrieval engine.

[go-ir]:https://github.com/cloudwalkio/go-ir
[gocrawl]:https://github.com/PuerkitoBio/gocrawl

## License

This project is released under the [MIT License](https://opensource.org/licenses/MIT).
