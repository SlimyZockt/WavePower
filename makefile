templ:
	@templ generate -watch --proxy="http://localhost:8080"

tailwind:
	@npx tailwindcss -o .\include_dir\output.css --watch

install:
	@go install github.com/a-h/templ/cmd/templ@latest

build: 
	@npx tailwindcss -o .\include_dir\output.css
	@templ generate
	@go build bin/server.go .

air: 
	@air .

dev: 
	make -j 3 templ tailwind air

