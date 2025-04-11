#!/usr/bin/env fish

tailwindcss -o ./include_dir/output.css -w 
templ generate -watch &
air . &

wait
