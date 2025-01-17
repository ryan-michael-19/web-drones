$env:GO111MODULE="on"
Set-Location ../src
# TODO: Fix pathing in main.go
# WARNING! This will only work if the server has its
# environment variables set to their default values.
go run ./main.go SCHEMA
Set-Location -
npx tsx ./run_game.ts single
# Set-Location ..
# go run ./main.go SCHEMA
# Set-Location -
# npx tsx ./run_game.ts multi