FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server ./server
COPY .env.example .env
EXPOSE 8080
CMD ["./server"]
