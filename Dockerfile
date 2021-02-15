FROM golang:1.14

WORKDIR /go/src/todolist
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8000

CMD ["todolist"]