FROM golang:1.26.5-alpine
WORKDIR /app
COPY . .

RUN apk add --no-cache gcc musl-dev
ENV CGO_ENABLED=1

RUN go build -o main ./cmd/url-shortener
CMD ["./main"]
