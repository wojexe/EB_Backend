FROM golang:1.24-alpine
EXPOSE 8080

RUN apk --no-cache add alpine-sdk

WORKDIR /app

COPY . .
RUN go mod download \
&& go build -v -o compiled_app

CMD ["./compiled_app"]

