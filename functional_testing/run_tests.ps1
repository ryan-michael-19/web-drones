$env:GO111MODULE="on"
Set-Location ..
# TODO: Fix pathing in main.go
go run ./main.go SCHEMA
Set-Location -
npx tsx ./run_game.ts single
# Set-Location ..
# go run ./main.go SCHEMA
# Set-Location -
# npx tsx ./run_game.ts multi