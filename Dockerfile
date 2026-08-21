FROM golang:1.27 AS build

WORKDIR /src

# The player feature needs a C toolchain: gopus is a cgo binding to libopus,
# and this is the only thing in the codebase that needs CGO_ENABLED=1.
RUN apt-get update && apt-get install -y --no-install-recommends \
        gcc \
        libopus-dev \
        pkg-config \
        curl \
        unzip \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy go.mod/go.sum first so `go mod download` is cached across builds and
# only reruns when dependencies actually change, not on every source edit.
COPY app/go.mod app/go.sum ./
RUN go mod download && go mod verify

COPY app/ .

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

# yt-dlp ships a standalone Linux binary (bundles its own Python via
# PyInstaller), so no system python/pip is needed. Deliberately not pinned:
# yt-dlp breaks against YouTube's anti-bot changes often enough that staying
# on "latest" here is more reliable long-term than a version that goes stale.
RUN curl -fL -o /out/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux \
    && chmod +x /out/yt-dlp

# yt-dlp needs a JS runtime to solve YouTube's signature challenges
# yt-dlp uses Deno automatically when it's on PATH.
RUN curl -fL -o /tmp/deno.zip https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip \
    && unzip -o /tmp/deno.zip -d /out \
    && chmod +x /out/deno

FROM debian:12-slim

# ffmpeg/libopus0 are external tools the player execs at runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
        ffmpeg \
        libopus0 \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home --uid 1000 bot

WORKDIR /app
COPY --from=build /out/bot ./bot
COPY --from=build /out/yt-dlp /usr/local/bin/yt-dlp
COPY --from=build /out/deno /usr/local/bin/deno

USER bot

ENTRYPOINT ["./bot"]
