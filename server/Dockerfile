FROM golang:1.21.6-alpine
LABEL author="github.com/kidommoc"
LABEL repo="github.com/kidommoc/gustrody"

COPY go.mod go.sum main.go internal/ /app/
WORKDIR /app
RUN go mod download
RUN go build -o ./build/austrody

ENTRYPOINT ["./build/austrody"]