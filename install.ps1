# glyph installer — Windows (PowerShell 5.1+)
#
# Interactive usage:
#   .\install.ps1
#
# Non-interactive / scripted:
#   .\install.ps1 -All
#   .\install.ps1 -Tools pin,ask
#   irm https://github.com/reky0/glyph/releases/latest/download/install.ps1 | iex
#
# Parameters:
#   -Tools   Comma-separated list of tools to install (pin, ask, diff, stand)
#   -All     Install all tools without prompting
#   -Version Specific release version to install (default: latest)
#   -InstallDir  Directory to install binaries (default: %LOCALAPPDATA%\glyph\bin)

[CmdletBinding()]
param(
    [string[]]$Tools,
    [switch]$All,
    [string]$Version,
    [string]$InstallDir
)

$ErrorActionPreference = "Stop"

$REPO     = "reky0/glyph"
$ALL_TOOLS = @("pin", "ask", "diff", "stand")
$DESC = @{
    pin   = "clipboard for URLs, commands, paths and notes"
    ask   = "ask an AI with automatic directory context"
    diff  = "explain a git diff in plain English"
    stand = "generate a standup from recent git commits"
}

# ── colours ───────────────────────────────────────────────────────────────────
function Write-Info    { Write-Host "  -> " -NoNewline -ForegroundColor Cyan;   Write-Host $args }
function Write-Success { Write-Host "  v  " -NoNewline -ForegroundColor Green;  Write-Host $args }
function Write-Warn    { Write-Host "  !  " -NoNewline -ForegroundColor Yellow; Write-Host $args }
function Write-Err     { Write-Host "  x  " -NoNewline -ForegroundColor Red;    Write-Host $args -ForegroundColor Red }

function Die([string]$msg) { Write-Err $msg; exit 1 }

# ── detect architecture ───────────────────────────────────────────────────────
function Get-Platform {
    $arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else {
        Die "32-bit Windows is not supported."
    }
    # ARM64 detection
    $cpu = (Get-WmiObject Win32_Processor -ErrorAction SilentlyContinue).Architecture
    if ($cpu -eq 12) { $arch = "arm64" }   # 12 = ARM64
    return "windows-$arch"
}

# ── fetch latest version ──────────────────────────────────────────────────────
function Get-LatestVersion {
    try {
        $resp = Invoke-RestMethod "https://api.github.com/repos/$REPO/releases/latest"
        return $resp.tag_name
    } catch {
        Die "Could not fetch latest version from GitHub: $_"
    }
}

# ── banner ────────────────────────────────────────────────────────────────────
function Show-Banner {
    Write-Host ""
    Write-Host "  glyph installer" -ForegroundColor White
    Write-Host "  Small CLI tools for your terminal memory" -ForegroundColor DarkGray
    Write-Host ""
}

# ── interactive selection ─────────────────────────────────────────────────────
function Select-Tools {
    Write-Host "  Available tools:" -ForegroundColor White
    Write-Host ""
    for ($i = 0; $i -lt $ALL_TOOLS.Count; $i++) {
        $t = $ALL_TOOLS[$i]
        Write-Host ("    {0}  {1,-7}  {2}" -f ($i+1), $t, $DESC[$t]) -ForegroundColor DarkGray
    }
    Write-Host ""
    Write-Host "  Enter numbers to install (e.g. " -NoNewline
    Write-Host "1 3" -NoNewline -ForegroundColor White
    Write-Host "), or " -NoNewline
    Write-Host "a" -NoNewline -ForegroundColor White
    Write-Host " for all:"
    Write-Host -NoNewline "  > "

    $input = Read-Host

    if ($input -match '^[Aa](ll)?$') {
        return $ALL_TOOLS
    }

    $selected = @()
    foreach ($token in ($input -split '\s+')) {
        if ($token -match '^[1-4]$') {
            $selected += $ALL_TOOLS[[int]$token - 1]
        } elseif ($token -ne "") {
            Write-Warn "Ignoring unknown selection: $token"
        }
    }
    if ($selected.Count -eq 0) { Die "No tools selected." }
    return $selected
}

# ── install a single tool ──────────────────────────────────────────────────────
function Install-Tool([string]$tool, [string]$ver, [string]$platform, [string]$dir) {
    $url  = "https://github.com/$REPO/releases/download/$ver/glyph-$tool-$platform.zip"
    $tmp  = Join-Path $env:TEMP "glyph-$tool-$([System.IO.Path]::GetRandomFileName())"
    $zip  = "$tmp.zip"

    Write-Info "Downloading $tool $ver ($platform)..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
    } catch {
        Die "Failed to download ${tool}: $_"
    }

    Expand-Archive -Path $zip -DestinationPath $tmp -Force
    Remove-Item $zip -Force

    $exe = Get-ChildItem -Path $tmp -Filter "$tool.exe" -Recurse | Select-Object -First 1
    if (-not $exe) { Die "Could not find $tool.exe in the downloaded archive." }

    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    $dest = Join-Path $dir "$tool.exe"
    Move-Item -Path $exe.FullName -Destination $dest -Force
    Remove-Item $tmp -Recurse -Force

    Write-Success "Installed $tool  ->  $dest"
}

# ── PATH update ───────────────────────────────────────────────────────────────
function Update-UserPath([string]$dir) {
    $current = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    if ($current -split ";" | Where-Object { $_ -eq $dir }) {
        return   # already present
    }

    Write-Host ""
    Write-Warn "$dir is not in your PATH."
    Write-Host ""

    if (-not [System.Environment]::GetCommandLineArgs().Contains("-NonInteractive")) {
        Write-Host "  Add it to your user PATH automatically? [Y/n] " -NoNewline
        $ans = Read-Host
        if ($ans -eq "" -or $ans -match '^[Yy]') {
            $new = "$dir;$current".TrimEnd(";")
            [System.Environment]::SetEnvironmentVariable("PATH", $new, "User")
            $env:PATH = "$dir;$env:PATH"
            Write-Success "PATH updated. Changes take effect in new terminals."
            return
        }
    }

    Write-Host ""
    Write-Host "  Add manually — run in PowerShell:" -ForegroundColor DarkGray
    Write-Host "    `$env:PATH = `"$dir;`$env:PATH`"" -ForegroundColor White
    Write-Host "  Or permanently via System Properties > Environment Variables." -ForegroundColor DarkGray
}

# ── diff warning ──────────────────────────────────────────────────────────────
function Show-DiffWarning([string]$dir) {
    # Windows has no system 'diff', so just note the name.
    Write-Host ""
    Write-Warn "Note: 'diff.exe' is glyph's AI diff explainer, not a Unix patch tool."
}

# ── main ──────────────────────────────────────────────────────────────────────
Show-Banner

$platform   = Get-Platform
$ver        = if ($Version)    { $Version }    else { Get-LatestVersion }
$installDir = if ($InstallDir) { $InstallDir } else { Join-Path $env:LOCALAPPDATA "glyph\bin" }

Write-Host "  Version  : " -NoNewline; Write-Host $ver        -ForegroundColor White
Write-Host "  Platform : " -NoNewline; Write-Host $platform   -ForegroundColor White
Write-Host "  Install  : " -NoNewline; Write-Host $installDir -ForegroundColor White
Write-Host ""

# Resolve tool selection.
$selected = @()
if ($All) {
    $selected = $ALL_TOOLS
} elseif ($Tools -and $Tools.Count -gt 0) {
    foreach ($t in $Tools) {
        if ($ALL_TOOLS -contains $t) { $selected += $t }
        else { Write-Warn "Unknown tool '$t', skipping." }
    }
    if ($selected.Count -eq 0) { Die "No valid tools specified. Available: $($ALL_TOOLS -join ', ')" }
} else {
    # Check if we're running piped (no console) — install all silently.
    if ([System.Console]::IsInputRedirected) {
        Write-Warn "Non-interactive mode: installing all tools."
        $selected = $ALL_TOOLS
    } else {
        $selected = Select-Tools
    }
}

Write-Host ""
foreach ($tool in $selected) {
    Install-Tool $tool $ver $platform $installDir
}

if ($selected -contains "diff") { Show-DiffWarning $installDir }
Update-UserPath $installDir

Write-Host ""
Write-Success "Done! Run any tool with --help to get started."
Write-Host ""
