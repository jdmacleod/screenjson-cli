FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/screenjson ./cmd/screenjson

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    poppler-utils \
    ca-certificates \
    curl \
    file \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/screenjson /usr/local/bin/screenjson

ENV SCREENJSON_PDFTOHTML=/usr/bin/pdftohtml

EXPOSE 8080

ENTRYPOINT ["screenjson"]
CMD ["help"]
