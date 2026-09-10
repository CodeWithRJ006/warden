$ErrorActionPreference = "Stop"

function Invoke-Warden {
    param($Uri, $Actor, $Tool, $Amount)
    $body = @{ tool=$Tool; amount=$Amount } | ConvertTo-Json
    try {
        return Invoke-RestMethod -Uri $Uri -Method Post -Body $body -ContentType "application/json" -Headers @{"X-Warden-Actor"=$Actor}
    } catch {
        return $_
    }
}

Write-Host "DEMO: Local model down + RESTRICTED data (John Doe)"
$res = Invoke-Warden "http://localhost:8081/v1/tools/execute" "support-agent" "get_payment for John Doe" 10.0
if ($res -is [System.Management.Automation.ErrorRecord]) {
    Write-Host "Failed: $($res.Exception.Message)" -ForegroundColor Red
} else {
    Write-Host "Success: allowed"
}

Write-Host "---"
Write-Host "DEMO: Local model down + CLEAN data (no PII)"
$res = Invoke-Warden "http://localhost:8081/v1/tools/execute" "support-agent" "get_payment" 10.0
if ($res -is [System.Management.Automation.ErrorRecord]) {
    Write-Host "Failed: $($res.Exception.Message)" -ForegroundColor Red
} else {
    Write-Host "Success: $($res.decision)" -ForegroundColor Green
}
