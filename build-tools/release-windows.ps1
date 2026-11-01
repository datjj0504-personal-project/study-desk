[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [Parameter(Mandatory = $true)][string[]]$Arguments
    )

    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $Command $($Arguments -join ' ')"
    }
}

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$innoCandidates = @(
    (Join-Path $env:ProgramFiles 'Inno Setup 7\ISCC.exe'),
    (Join-Path ${env:ProgramFiles(x86)} 'Inno Setup 7\ISCC.exe'),
    (Join-Path $env:ProgramFiles 'Inno Setup 6\ISCC.exe'),
    (Join-Path ${env:ProgramFiles(x86)} 'Inno Setup 6\ISCC.exe')
)
$innoCommand = Get-Command 'ISCC.exe' -ErrorAction SilentlyContinue
$innoPath = $null
if ($innoCommand) {
    $innoPath = $innoCommand.Source
}
if (-not $innoCommand) {
    foreach ($candidate in $innoCandidates) {
        if (Test-Path -LiteralPath $candidate -PathType Leaf) {
            $innoPath = $candidate
            break
        }
    }
}
if (-not $innoPath) {
    throw 'Inno Setup 6 or 7 is required. Install it from https://jrsoftware.org/isinfo.php and ensure ISCC.exe is on PATH or in its default installation folder.'
}

$goCommand = Get-Command 'go.exe' -ErrorAction Stop
$wailsCommand = Get-Command 'wails.exe' -ErrorAction Stop
$npmCommand = Get-Command 'npm.cmd' -ErrorAction Stop

$appSource = Get-Content -LiteralPath (Join-Path $projectRoot 'app.go') -Raw
$versionMatches = [regex]::Matches($appSource, '(?m)^const appVersion = "(\d+\.\d+\.\d+)"\s*$')
if ($versionMatches.Count -ne 1) {
    throw 'Expected exactly one three-part appVersion constant in app.go.'
}
$appVersion = $versionMatches[0].Groups[1].Value

$appExe = Join-Path $projectRoot 'build\bin\studydesk.exe'
$installerScript = Join-Path $projectRoot 'build\windows\installer.iss'
$releaseDirectory = Join-Path $projectRoot 'release'
$installerName = "StudyDesk-Setup-$appVersion.exe"
$archiveName = "StudyDesk-Windows-x64-$appVersion.zip"
$webView2BootstrapperUrl = 'https://go.microsoft.com/fwlink/p/?LinkId=2124703'

$stageDirectory = Join-Path $env:TEMP ("StudyDesk-release-" + [Guid]::NewGuid().ToString('N'))
$bootstrapperPath = Join-Path $stageDirectory 'MicrosoftEdgeWebView2Setup.exe'
$stagedInstaller = Join-Path $stageDirectory $installerName
$stagedArchive = Join-Path $stageDirectory $archiveName

try {
    New-Item -ItemType Directory -Path $stageDirectory | Out-Null

    Push-Location $projectRoot
    try {
        Invoke-CheckedCommand -Command $goCommand.Source -Arguments @('test', './...')
        Invoke-CheckedCommand -Command $npmCommand.Source -Arguments @('--prefix', 'frontend', 'ci')
        Invoke-CheckedCommand -Command $npmCommand.Source -Arguments @('--prefix', 'frontend', 'run', 'test')
        Invoke-CheckedCommand -Command $npmCommand.Source -Arguments @('--prefix', 'frontend', 'run', 'build')
        Invoke-CheckedCommand -Command $wailsCommand.Source -Arguments @('build', '-platform', 'windows/amd64')
    } finally {
        Pop-Location
    }

    if (-not (Test-Path -LiteralPath $appExe -PathType Leaf)) {
        throw "Wails build did not produce the expected executable: $appExe"
    }

    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $webView2BootstrapperUrl -OutFile $bootstrapperPath -UseBasicParsing
    $signature = Get-AuthenticodeSignature -FilePath $bootstrapperPath
    if ($signature.Status -ne 'Valid' -or $signature.SignerCertificate.Subject -notmatch 'CN=Microsoft Corporation') {
        throw "Downloaded WebView2 bootstrapper did not have a valid Microsoft Corporation signature (status: $($signature.Status))."
    }

    $compilerArguments = @(
        "/DAppVersion=$appVersion",
        "/DAppExePath=$appExe",
        "/DWebView2BootstrapperPath=$bootstrapperPath",
        "/DOutputDirectory=$stageDirectory",
        $installerScript
    )
    Invoke-CheckedCommand -Command $innoPath -Arguments $compilerArguments

    if (-not (Test-Path -LiteralPath $stagedInstaller -PathType Leaf)) {
        throw "Inno Setup did not produce the expected installer: $stagedInstaller"
    }

    Compress-Archive -LiteralPath $stagedInstaller -DestinationPath $stagedArchive -CompressionLevel Optimal
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive = [System.IO.Compression.ZipFile]::OpenRead($stagedArchive)
    try {
        $installerEntry = @($archive.Entries | Where-Object { $_.FullName -eq $installerName })
        if ($archive.Entries.Count -ne 1 -or $installerEntry.Count -ne 1 -or
            $installerEntry[0].Length -ne (Get-Item -LiteralPath $stagedInstaller).Length) {
            throw 'Release ZIP verification failed: it must contain exactly the generated installer with its full length.'
        }
    } finally {
        $archive.Dispose()
    }

    New-Item -ItemType Directory -Path $releaseDirectory -Force | Out-Null
    Copy-Item -LiteralPath $stagedInstaller -Destination (Join-Path $releaseDirectory $installerName) -Force
    Copy-Item -LiteralPath $stagedArchive -Destination (Join-Path $releaseDirectory $archiveName) -Force

    Write-Output "Release artifacts created in $releaseDirectory"
    Get-Item -LiteralPath (Join-Path $releaseDirectory $installerName), (Join-Path $releaseDirectory $archiveName) |
        Select-Object Name, Length
} finally {
    if (Test-Path -LiteralPath $stageDirectory -PathType Container) {
        Remove-Item -LiteralPath $stageDirectory -Recurse -Force
    }
}
