FROM golang:1.23

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go install github.com/air-verse/air@v1.61.7 || \
	go install github.com/air-verse/air@v1.61.7 || \
	go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./

RUN go mod download || \
	go mod download || \
	go mod download

COPY . .

CMD ["air"]
