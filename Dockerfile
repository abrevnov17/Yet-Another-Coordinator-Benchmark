FROM golang:1.14

ENV GO111MODULE=on

RUN mkdir /yac
WORKDIR /yac

COPY go.mod .
COPY go.sum .

RUN go mod download

ADD src src

COPY run.sh .
RUN chmod +x run.sh

EXPOSE 8080
ENTRYPOINT ["./run.sh"]