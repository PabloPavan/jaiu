#!/usr/bin/env bash
set -euo pipefail

npx tailwindcss -i ./web/tailwind/input.css -o ./web/static/css/app.css --watch
