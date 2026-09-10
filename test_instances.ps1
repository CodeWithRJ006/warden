$actor = "finance-operator"
$body = @{
    actor = $actor
    tool = "create_refund"
    amount = 50.0
} | ConvertTo-Json

Write-Host "Sending 9 requests to gateway-1 (port 8081)..."
for ($i=1; $i -le 9; $i++) {
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8081/v1/tools/execute" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
        Write-Host "Req $i Decision:" $response.decision "Reason:" $response.reason
    } catch {
        Write-Host "Req $i Failed:" $_.Exception.Message
    }
}

Write-Host "Sending 10th request to gateway-2 (port 8082)..."
try {
    $response10 = Invoke-RestMethod -Uri "http://localhost:8082/v1/tools/execute" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
    Write-Host "Req 10 Decision:" $response10.decision
} catch {
    Write-Host "Req 10 Failed:" $_.Exception.Message
}

Write-Host "Sending 11th request to gateway-2 (port 8082) [SHOULD BE 429]..."
try {
    $response11 = Invoke-RestMethod -Uri "http://localhost:8082/v1/tools/execute" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
    Write-Host "Req 11 Decision:" $response11.decision
} catch {
    Write-Host "Req 11 Failed:" $_.Exception.Response.StatusCode.value__
}
