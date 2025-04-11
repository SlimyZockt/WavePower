#!/usr/bin/env bash

pnpm exec tailwindcss -o ./include_dir/output.css -w  &
templ generate -watch &
air . &

wait
