templ:
	@templ generate -watch 

# --proxy="http://localhost:8080"

tailwind:
	@tailwindcss -o .\include_dir\output.css --watch

install:
	@go install github.com/a-h/templ/cmd/templ@latest

build: 
	@tailwindcss -o .\include_dir\output.css
	@templ generate
	@go build bin/server.go .

air: 
	@air .

dev: 
	make -j 3 templ tailwind air


refresh_db:
	goose -dir=assets/migrations/ sqlite3 app.db down
	goose -dir=assets/migrations/ sqlite3 app.db up

setup:
	pnpm install
	ggoose -dir=assets/migrations/ sqlite3 app.db up
	echo "Don't Forget to add the .env and TLS Key(\"server.key\" & \"server.pem\")!!!"
