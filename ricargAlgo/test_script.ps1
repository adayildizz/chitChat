# PowerShell script to run multiple nodes for testing Ricart-Agrawala algorithm
# Run this script to start 3 nodes (A, B, C) in separate windows

Write-Host "Starting Ricart-Agrawala Algorithm Test" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "This will start 3 nodes in separate PowerShell windows" -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop all nodes" -ForegroundColor Yellow
Write-Host ""

# Start Node A
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; go run . --id A --listen :50051 --mode manual"
Start-Sleep -Milliseconds 500

# Start Node B
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; go run . --id B --listen :50052 --mode manual"
Start-Sleep -Milliseconds 500

# Start Node C
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; go run . --id C --listen :50053 --mode manual"

Write-Host "All nodes started!" -ForegroundColor Green
Write-Host "In each window, press ENTER to request critical section" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test Scenarios:" -ForegroundColor Yellow
Write-Host "1. Request CS from different nodes simultaneously" -ForegroundColor White
Write-Host "2. Request CS sequentially" -ForegroundColor White
Write-Host "3. Verify mutual exclusion (only one node in CS at a time)" -ForegroundColor White
Write-Host "4. Check Lamport timestamps are consistent" -ForegroundColor White

