FROM golang:latest AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app

FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /app/app .
CMD ["./app"]