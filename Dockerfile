FROM golang:1.20 as build
WORKDIR /app
COPY go.mod ./
RUN go mod install 
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o lb .

FROM apline:latest
RUN apk --no-chache add ce-certificate
WORKDIR /root
COPY --from=build /app/lb .
ENTRYPOINT ["/root/lb"]