#!/usr/bin/env fish

npx tailwindcss -o ./include_dir/output.css -w  &
templ generate -watch &
air . &

wait
