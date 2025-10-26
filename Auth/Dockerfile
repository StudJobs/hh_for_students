FROM golang:1.25.1-alpine as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build \
 -a \
 -installsuffix cgo \
 -ldflags='-w -s' \
 -o app \
 ./cmd/main.go

FROM scratch

COPY --from=build /app/configs /configs
COPY --from=build /app/schema /schema
COPY --from=build /app/app /app

EXPOSE 50051

CMD ["/app"]