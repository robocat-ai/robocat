FROM golang:1.19 as build

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY internal ./internal
COPY *.go ./

RUN CGO_ENABLED=0 go build -v -o main .

FROM ghcr.io/robocat-ai/robocat-base:latest

COPY --from=build /app/main /usr/local/bin/robocat
