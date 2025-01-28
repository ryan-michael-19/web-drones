#!/usr/bin/env bash
echo "DATABASE_URL=postgres://user:$(cat ../postgres_pw.txt)@localhost:5432/webdrones" > .env
