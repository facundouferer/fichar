#!/usr/bin/env pwsh

Write-Host "Validating docker-compose.yml configuration..." -ForegroundColor Green

try {
    $config = docker-compose config
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ docker-compose.yml is valid" -ForegroundColor Green
        exit 0
    }
} catch {
    Write-Host "✗ docker-compose.yml has errors: $_" -ForegroundColor Red
    exit 1
}
