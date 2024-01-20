FROM golang:1.16
WORKDIR /app
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["/app/main"]