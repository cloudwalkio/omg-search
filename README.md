omg-search
==========

We use [gocrawl] to crawl our documentation pages and our [go-ir] package to create an Information Retrieval system, aka, a search system.

Usage
-----

In starting the server we must pass a token that will be used to authorization in any requests to the API: 
```
docker run -d -p 80:5000 cloudwalk/search "my_token"
```

The API accepts the following GET requests:

  * `/en?query=search+me&access_token=my_token`: search for "search me" in the english documentation.
  * `/pt-BR?query=pesquise+me&access_token=my_token`: search for "pesquise-me" in the portuguese documentation.
 
The API accepts the following POST requests:

  * `/crawl?access_token=my_token`: recrawl the english and portugues pages and update the Information Retrieval engine.
    * `/crawl/en?access_token=my_token`: recrawl only the english pages and update the Information Retrieval engine.
    * `/crawl/pt-BR?access_token=my_token`: recrawl only the portuguese pages and update the Information Retrieval engine.

Logs
----

The server logs are saved to `LOG_DIR` environment variable, which defaults to `/var/log/docker/search`. To access it from the host, outise the container, we can mount a host directory to LOG_DIR when starting the container, using flag `-v`:
```
docker run -d -v /path/to/host/log/dir:/var/log/docker/search -p 80:5000 cloudwalk/search "my_token"
```

An alternative is to mount the volume from another container, using the `--volumes-from` flag.

[go-ir]:https://github.com/cloudwalkio/go-ir
[gocrawl]:https://github.com/PuerkitoBio/gocrawl
