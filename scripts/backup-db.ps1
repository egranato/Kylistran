# Copies the live SQLite DB out of the kylistran-api container onto the
# external drive, with a rotation window and a basic integrity check.

$ErrorActionPreference = "Stop"

# --- config: adjust to match your NUC ---
$ContainerName     = "kylistran-api"
$DbPathInContainer = "/data/kylistran.db"
$BackupDir         = "D:\kylistran"
$RetentionDays     = 30
# ---

New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$dest = Join-Path $BackupDir "kylistran-$timestamp.db"

docker cp "${ContainerName}:${DbPathInContainer}" $dest

if (-not (Test-Path $dest) -or (Get-Item $dest).Length -eq 0) {
    Write-Error "backup failed: $dest is empty or missing"
    Remove-Item -Force -ErrorAction SilentlyContinue $dest
    exit 1
}

$sqlite3 = Get-Command sqlite3 -ErrorAction SilentlyContinue
if ($sqlite3) {
    $result = & sqlite3 $dest "PRAGMA integrity_check;"
    if ($result -ne "ok") {
        Write-Error "backup failed integrity check: $dest ($result)"
        exit 1
    }
}

Get-ChildItem -Path $BackupDir -Filter "kylistran-*.db" |
    Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$RetentionDays) } |
    Remove-Item -Force

Write-Output "backup ok: $dest"
