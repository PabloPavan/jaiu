#!/usr/bin/env bash
set -euo pipefail

npx tailwindcss -c ./web/tailwind/tailwind.config.js -i ./web/tailwind/input.css -o ./web/static/css/app.css --watch
