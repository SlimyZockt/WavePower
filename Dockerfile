FROM alpine:3.22 AS final
WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    curl \
    tzdata \
    ffmpeg \
    sqlite \
    build-base \
    musl-dev \
    zig \
    nodejs \
    pnpm

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add templ cgo --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing/


COPY package.json package-lock.json ./
RUN pnpm install

COPY go.mod go.sum ./
RUN cgo mod download -x

COPY . .

RUN templ generate
RUN pnpm exec tailwindcss -o include_dir/output.css -m
# Expose the port that the application listens on.
EXPOSE 8080

RUN cgo build -v -o /bin/server

# What the container should run when it is started.
ENTRYPOINT [ "/bin/server" ]

