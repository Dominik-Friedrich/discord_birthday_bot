FROM golang:1.27 AS build

WORKDIR /src

# Copy go.mod/go.sum first so `go mod download` is cached across builds and
# only reruns when dependencies actually change, not on every source edit.
COPY app/go.mod app/go.sum ./
RUN go mod download && go mod verify

COPY app/ .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/bot ./bot

USER nonroot:nonroot

ENTRYPOINT ["./bot"]
