# syntax=docker/dockerfile:1

################################################################################
# Create a stage for generating templ 
FROM ghcr.io/a-h/templ:latest AS generate-templ
COPY --chown=65532:65532  . /app
WORKDIR /app


RUN ["templ", "generate"]

################################################################################
FROM alpine:latest AS final
COPY --from=generate-templ /app /app
WORKDIR /app

RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
    curl \
    tzdata \
    ffmpeg \
    sqlite \
    go \
    nodejs \
    npm 

RUN npm install
RUN curl -sLo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.16/tailwindcss-linux-x64 \
    &&\
    chmod +x tailwindcss
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

RUN curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh    && \ 
    goose -dir=assets/migrations/ sqlite3 app.db up
RUN ./tailwindcss -o include_dir/output.css -m
RUN GOOS=linux go build -o /bin/server
# Expose the port that the application listens on.
EXPOSE 8080

# What the container should run when it is started.
ENTRYPOINT [ "/bin/server" ]

