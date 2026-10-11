FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bot ./cmd/bot

FROM gcr.io/distroless/static-debian12
COPY --from=build /bot /bot
ENTRYPOINT ["/bot"]
