#!/usr/bin/env pwsh
# Test script - Only process the last week to save tokens and time
# Usage: .\test_last_week.ps1

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   🧪 TEST MODE: Last Week Only" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if pipeline exists
if (-not (Test-Path ".\pipeline.exe")) {
    Write-Host "Building pipeline..." -ForegroundColor Yellow
    go build -o pipeline.exe main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "❌ Build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Build successful" -ForegroundColor Green
    Write-Host ""
}

Write-Host "� Running pipeline in TEST mode..." -ForegroundColor Green
Write-Host "⏱️  Expected: ~2 minutes for 10 kids" -ForegroundColor Yellow
Write-Host ""

# Set environment variable to only process last week
$env:TEST_LAST_WEEK_ONLY = "true"

# Run pipeline
.\pipeline.exe

$exitCode = $LASTEXITCODE

# Clear the test mode variable
$env:TEST_LAST_WEEK_ONLY = ""

Write-Host ""
if ($exitCode -eq 0) {
    Write-Host "============================================" -ForegroundColor Green
    Write-Host "   ✅ Test Completed Successfully" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "📁 Check outputs:" -ForegroundColor Cyan
    Write-Host "   - data\kids_analysis_week_*.json (Silver Layer)" -ForegroundColor Gray
    Write-Host "   - data\kids_reports_week_*.json (Gold Layer)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "💰 Token costs are shown in the output above" -ForegroundColor Yellow
} else {
    Write-Host "============================================" -ForegroundColor Red
    Write-Host "   ❌ Test Failed" -ForegroundColor Red
    Write-Host "============================================" -ForegroundColor Red
    exit $exitCode
}
