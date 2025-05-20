#!/usr/bin/env bash

go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/a-h/templ/cmd/templ@latest
goose -dir=assets/migrations/ sqlite3 app.db up
pnpm install
pnpm exec tailwindcss -o ./include_dir/output.css
templ generate
echo "Don't Forget to add the .env and TLS Key(\"server.key\" & \"server.pem\")!!!"
