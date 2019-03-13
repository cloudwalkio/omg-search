# omg-search

We use [gocrawl] to crawl our documentation pages and our [go-ir] package to create an Information Retrieval system, aka, a search system.

## Usage

Export two environment variables:

```
export PORT=port_number
export TOKEN=choose_a_token
```

The token will be requested when making HTTP requests to this service.

Generate the binary and run it:

```
$ go build
$ ./omg-search
```

It will start crawling the documentation pages in both languages, English and Portuguese and send logs to the console.

## GAE (Google App Engine)

This service is configured to run on GAE, hence the `app.yaml` file.

To deploy a new version:

```
gcloud app deploy --project GCP-PROJECT
```

The deploy will happen based on the configuration information inside `app.yaml` and your current `gcloud` settings. **Make sure everything is correct** .

## API calls
The API accepts the following GET requests:

  * `/en?query=search+me&access_token=my_token`: search for "search me" in the English documentation.
  * `/pt-BR?query=pesquise+me&access_token=my_token`: search for "pesquise-me" in the Portuguese documentation.

The API accepts the following POST requests:

  * `/crawl?access_token=my_token`: recrawl the English and Portuguese pages and update the Information Retrieval engine.
    * `/crawl/en?access_token=my_token`: recrawl only the English pages and update the Information Retrieval engine.
    * `/crawl/pt-BR?access_token=my_token`: recrawl only the Portuguese pages and update the Information Retrieval engine.

[go-ir]:https://github.com/cloudwalkio/go-ir
[gocrawl]:https://github.com/PuerkitoBio/gocrawl

## License

This project is released under the [MIT License](https://opensource.org/licenses/MIT).
