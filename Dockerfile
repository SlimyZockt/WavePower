# syntax=docker/dockerfile:1
FROM alpine:3.21 AS final
WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    curl \
    tzdata \
    ffmpeg \
    sqlite \
    go \
    nodejs \
    pnpm

RUN curl -sLo templ.tar.gz https://github.com/a-h/templ/releases/download/v0.2.793/templ_Linux_x86_64.tar.gz \
    &&\
    mkdir templ &&\
    tar -xzf templ.tar.gz -C templ

RUN curl -sLo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.16/tailwindcss-linux-x64 \
    &&\
    chmod +x tailwindcss

RUN curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh 

COPY package.json package-lock.json ./
RUN pnpm install

COPY go.mod go.sum ./
RUN go mod download -x

COPY . ./

RUN templ/templ generate

RUN goose -dir=assets/migrations/ sqlite3 app.db up

RUN ./tailwindcss -o include_dir/output.css -m
RUN GOOS=linux go build -o /bin/server
# Expose the port that the application listens on.
EXPOSE 8080

# What the container should run when it is started.
ENTRYPOINT [ "/bin/server" ]

