FROM luisbebop/go1.2

RUN \
    go get github.com/codegangsta/martini;\
    go get github.com/PuerkitoBio/gocrawl;\
    go get github.com/cloudwalkio/go-ir;
#RUN

ADD server.go /src/omg-search/server.go
ADD startup.sh /src/omg-search/startup.sh

# Build Go server's binary
RUN \
    cd /src/omg-search/;\
    go build;
#RUN

WORKDIR /src/omg-search/

# Save logs to LOG_DIR
ENV LOG_DIR /var/log/docker/search

ENTRYPOINT ["./startup.sh"]
