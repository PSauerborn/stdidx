FROM golang:1.25 AS tests

WORKDIR /app/tests

COPY go.mod go.sum ./
RUN go mod download

RUN go install gotest.tools/gotestsum@latest

COPY tests ./tests
COPY *.go ./

CMD ["gotestsum", "--format", "testdox"]

FROM golang:1.25 AS build

ARG CPU_ARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o api .
