# syntax=docker/dockerfile:1

FROM golang:1.24.4

WORKDIR /app

COPY . .
RUN go mod download

WORKDIR /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server

# whatever port you want
ENV PORT="8080"
EXPOSE $PORT

CMD ["sh","-c","/server -p $PORT"]
