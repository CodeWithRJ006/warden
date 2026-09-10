Write-Host "DEMO: Local model down + RESTRICTED data (John Doe)"
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/v1/tools/execute" -Method Post -Body '{"actor":"finance-operator","tool":"create_refund for John Doe","amount":50.0}' -ContentType "application/json" -ErrorAction Stop
    Write-Host "Decision:" $response.decision "Reason:" $response.reason
} catch {
    Write-Host "Failed:" $_.Exception.Message
}
Write-Host "---"

Write-Host "DEMO: Local model down + CLEAN data (no PII)"
try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:8081/v1/tools/execute" -Method Post -Body '{"actor":"finance-operator","tool":"create_refund","amount":50.0}' -ContentType "application/json" -ErrorAction Stop
    Write-Host "Decision:" $response2.decision "Reason:" $response2.reason
} catch {
    Write-Host "Failed:" $_.Exception.Message
}
