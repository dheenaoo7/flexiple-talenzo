FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/server ./server
COPY data ./data

ENV PORT=8080
ENV DB_PATH=/data/talent_search.db
ENV PROFILES_PATH=/app/data/profiles.json

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["./server"]
