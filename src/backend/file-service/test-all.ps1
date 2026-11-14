$ErrorActionPreference = "Stop"

Write-Host "Running Go tests..."
go test -v -count=1 ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Starting MongoDB..."
docker compose -f ../compose.yaml -p backend up -d --wait mongo-file-service
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Running MongoDB integration tests..."
$backend = (Resolve-Path ..).Path

docker run --rm `
  --network backend_default `
  --env-file .env `
  -e MONGO_DB=files_test_db `
  -e RUN_MONGO_INTEGRATION=1 `
  -v "${backend}:/app" `
  -w /app/file-service `
  golang:1.26.10-alpine `
  go test -v -count=1 ./internal/infrastructure/repository

if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "All tests passed!"
