param(
    [string]$ConfigPath = (Join-Path $PSScriptRoot '..\config.local.json'),
    [string]$BaseUrl = 'http://127.0.0.1:8080'
)
$ErrorActionPreference = 'Stop'
$config = Get-Content -Raw -LiteralPath $ConfigPath | ConvertFrom-Json
$headers = @{ Authorization = "Bearer $($config.token)" }

# Run this after starting the Go server. Never print the private headers/config.
$health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get
if (-not $health.success) { throw 'Health check failed' }
foreach ($action in @('hello', 'success', 'boolean', 'echo')) {
    $params = @()
    if ($action -eq 'echo') { $params = @('Alex', 42, $true, @{ message = 'dynamic params' }) }
    $body = @{ action = $action; params = $params; time = [DateTime]::UtcNow.ToString('o') } | ConvertTo-Json -Depth 10
    $response = Invoke-RestMethod -Uri "$BaseUrl/api" -Method Post -Headers $headers -ContentType 'application/json' -Body $body
    if (-not $response.success) { throw "Action $action failed" }
    if ($action -eq 'hello' -and $response.result -cne 'hello world') { throw 'Wrong hello result' }
    if ($action -eq 'success' -and $response.result -cne 'successful') { throw 'Wrong success result' }
    if ($action -eq 'boolean' -and ($response.result -isnot [bool] -or -not $response.result)) { throw 'Result must be boolean true' }
    if ($action -eq 'echo' -and $response.result.Count -ne 4) { throw 'Echo did not preserve params' }
    Write-Output ($response | ConvertTo-Json -Depth 10 -Compress)
}
try {
    Invoke-RestMethod -Uri "$BaseUrl/api" -Method Post -ContentType 'application/json' -Body '{"action":"hello","params":[]}' | Out-Null
    throw 'Unauthenticated request unexpectedly succeeded'
} catch {
    if ($null -eq $_.Exception.Response -or [int]$_.Exception.Response.StatusCode -ne 401) { throw }
    Write-Output 'Unauthenticated request rejected: 401'
}
