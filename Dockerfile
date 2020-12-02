FROM golang:1.14

RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/serialx/hashring
RUN go get -u github.com/rs/xid

ADD run.sh run.sh
RUN chmod +x run.sh

ADD src src

EXPOSE 8080
ENTRYPOINT ["./run.sh"]