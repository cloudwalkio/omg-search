FROM luisbebop/go1.2

RUN \
    go get github.com/codegangsta/martini;\
    go get github.com/PuerkitoBio/gocrawl;\
    go get github.com/cloudwalkio/go-ir;\
#RUN

ADD server.go /src/omg-search/server.go

# Build Go server's binary
RUN \
    cd /src/omg-search/;\
    go build;\
#RUN

WORKDIR /src/omg-search/

ENTRYPOINT ["./omg-search"]
