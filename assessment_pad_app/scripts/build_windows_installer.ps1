param(
  [ValidateSet("local", "prod", "production", "online")]
  [string]$Target = "prod",

  [ValidateSet("release", "debug", "profile")]
  [string]$BuildMode = "release",

  [string]$ApiHost = "",
  [string]$ProdDomain = $env:PROD_DOMAIN,
  [string]$ProdScheme = $env:PROD_SCHEME,
  [string]$AppName = "Assessment Assistant",
  [string]$Publisher = "YBK",
  [string]$OutputBaseFilename = "AssessmentAssistantSetup",
  [string]$InnoSetupPath = ""
)

$ErrorActionPreference = "Stop"

$isWindowsHost = $env:OS -eq "Windows_NT"
if (-not $isWindowsHost) {
  throw "Windows installer builds must run on Windows. Flutter cannot build Windows executables on macOS or Linux hosts."
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
Set-Location $ProjectDir

if ([string]::IsNullOrWhiteSpace($ProdDomain)) {
  $ProdDomain = "irts-children.cn"
}
if ([string]::IsNullOrWhiteSpace($ProdScheme)) {
  $ProdScheme = "https"
}

function Get-CommandPath {
  param([string]$Name)
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if ($null -eq $cmd) {
    return ""
  }
  return $cmd.Source
}

function Detect-LanIp {
  $addresses = Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
    Where-Object {
      $_.IPAddress -notlike "127.*" -and
      $_.IPAddress -notlike "169.254.*" -and
      $_.PrefixOrigin -ne "WellKnown"
    } |
    Sort-Object InterfaceMetric, InterfaceIndex

  if ($addresses.Count -gt 0) {
    return $addresses[0].IPAddress
  }
  return ""
}

function EnvOrDefault {
  param([string]$Name, [string]$DefaultValue)
  $value = [Environment]::GetEnvironmentVariable($Name, "Process")
  if ([string]::IsNullOrWhiteSpace($value)) {
    return $DefaultValue
  }
  return $value
}

function Find-InnoSetupCompiler {
  param([string]$PreferredPath)

  if (-not [string]::IsNullOrWhiteSpace($PreferredPath) -and (Test-Path $PreferredPath)) {
    return (Resolve-Path $PreferredPath).Path
  }

  $fromPath = Get-CommandPath "ISCC.exe"
  if (-not [string]::IsNullOrWhiteSpace($fromPath)) {
    return $fromPath
  }

  $programFilesX86 = [Environment]::GetFolderPath([Environment+SpecialFolder]::ProgramFilesX86)
  $programFiles = [Environment]::GetFolderPath([Environment+SpecialFolder]::ProgramFiles)
  $candidates = @(
    (Join-Path $programFilesX86 "Inno Setup 6\ISCC.exe"),
    (Join-Path $programFiles "Inno Setup 6\ISCC.exe"),
    (Join-Path $programFilesX86 "Inno Setup 5\ISCC.exe"),
    (Join-Path $programFiles "Inno Setup 5\ISCC.exe")
  )

  foreach ($candidate in $candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return ""
}

$flutter = Get-CommandPath "flutter"
if ([string]::IsNullOrWhiteSpace($flutter)) {
  throw "flutter command not found. Install Flutter and add it to PATH first."
}

if ($Target -in @("prod", "production", "online")) {
  $Target = "prod"
  $LoginApiBaseUrl = EnvOrDefault "LOGIN_API_BASE_URL" "$ProdScheme://$ProdDomain"
  $EducationApiBaseUrl = EnvOrDefault "EDUCATION_API_BASE_URL" "$ProdScheme://$ProdDomain"
  $TrainingGameBaseUrl = EnvOrDefault "TRAINING_GAME_BASE_URL" "$ProdScheme://$ProdDomain/training-games/"
  $LoginTenantDomain = EnvOrDefault "LOGIN_TENANT_DOMAIN" "$ProdDomain"
  $LoginQrUrl = EnvOrDefault "LOGIN_QR_URL" "$ProdScheme://$ProdDomain/institution/"
} else {
  $Target = "local"
  $LoginApiBaseUrl = [Environment]::GetEnvironmentVariable("LOGIN_API_BASE_URL", "Process")
  $EducationApiBaseUrl = [Environment]::GetEnvironmentVariable("EDUCATION_API_BASE_URL", "Process")

  if ([string]::IsNullOrWhiteSpace($LoginApiBaseUrl) -or [string]::IsNullOrWhiteSpace($EducationApiBaseUrl)) {
    if ([string]::IsNullOrWhiteSpace($ApiHost)) {
      $ApiHost = Detect-LanIp
    }
    if ([string]::IsNullOrWhiteSpace($ApiHost)) {
      throw "Could not detect a LAN IP. Pass -ApiHost, for example: .\scripts\build_windows_installer.ps1 -Target local -ApiHost 192.168.1.23"
    }
  }

  if ([string]::IsNullOrWhiteSpace($LoginApiBaseUrl)) {
    $LoginApiBaseUrl = "http://$ApiHost`:8081"
  }
  if ([string]::IsNullOrWhiteSpace($EducationApiBaseUrl)) {
    $EducationApiBaseUrl = "http://$ApiHost`:8083"
  }
  $TrainingGameBaseUrl = EnvOrDefault "TRAINING_GAME_BASE_URL" ""
  $LoginTenantDomain = [Environment]::GetEnvironmentVariable("LOGIN_TENANT_DOMAIN", "Process")
  $LoginQrUrl = [Environment]::GetEnvironmentVariable("LOGIN_QR_URL", "Process")
}

$LoginSource = EnvOrDefault "LOGIN_SOURCE" "assessment-pad"
$DefaultLoginUsername = EnvOrDefault "DEFAULT_LOGIN_USERNAME" "17601241636"
$DefaultLoginPassword = EnvOrDefault "DEFAULT_LOGIN_PASSWORD" "123456"

Write-Host "Using BUILD_TARGET=$Target"
Write-Host "Using LOGIN_API_BASE_URL=$LoginApiBaseUrl"
Write-Host "Using EDUCATION_API_BASE_URL=$EducationApiBaseUrl"
Write-Host "Using TRAINING_GAME_BASE_URL=$TrainingGameBaseUrl"
Write-Host "Using DEFAULT_LOGIN_USERNAME=$DefaultLoginUsername"

& $flutter pub get
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}

$dartDefines = @(
  "--dart-define=LOGIN_API_BASE_URL=$LoginApiBaseUrl",
  "--dart-define=EDUCATION_API_BASE_URL=$EducationApiBaseUrl",
  "--dart-define=TRAINING_GAME_BASE_URL=$TrainingGameBaseUrl",
  "--dart-define=LOGIN_SOURCE=$LoginSource",
  "--dart-define=DEFAULT_LOGIN_USERNAME=$DefaultLoginUsername",
  "--dart-define=DEFAULT_LOGIN_PASSWORD=$DefaultLoginPassword"
)

if (-not [string]::IsNullOrWhiteSpace($LoginTenantDomain)) {
  $dartDefines += "--dart-define=LOGIN_TENANT_DOMAIN=$LoginTenantDomain"
}
if (-not [string]::IsNullOrWhiteSpace($LoginQrUrl)) {
  $dartDefines += "--dart-define=LOGIN_QR_URL=$LoginQrUrl"
}

& $flutter build windows "--$BuildMode" @dartDefines
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}

$configuration = switch ($BuildMode) {
  "release" { "Release" }
  "debug" { "Debug" }
  "profile" { "Profile" }
}

$bundleCandidates = @(
  (Join-Path $ProjectDir "build\windows\x64\runner\$configuration"),
  (Join-Path $ProjectDir "build\windows\runner\$configuration")
)

$bundleDir = ""
foreach ($candidate in $bundleCandidates) {
  if (Test-Path (Join-Path $candidate "assessment_pad_app.exe")) {
    $bundleDir = $candidate
    break
  }
}

if ([string]::IsNullOrWhiteSpace($bundleDir)) {
  throw "Windows bundle was not found. Expected assessment_pad_app.exe under build\windows\x64\runner\$configuration."
}

$iscc = Find-InnoSetupCompiler $InnoSetupPath
if ([string]::IsNullOrWhiteSpace($iscc)) {
  throw "Inno Setup compiler ISCC.exe was not found. Install Inno Setup 6, or pass -InnoSetupPath C:\Path\To\ISCC.exe."
}

$installerDir = Join-Path $ProjectDir "build\windows\installer"
New-Item -ItemType Directory -Force -Path $installerDir | Out-Null

$issPath = Join-Path $installerDir "assessment_assistant.iss"

$issContent = @"
#define MyAppName "$AppName"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "$Publisher"
#define MyAppExeName "assessment_pad_app.exe"

[Setup]
AppId=AssessmentAssistant
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir="$installerDir"
OutputBaseFilename=$OutputBaseFilename
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
UninstallDisplayIcon={app}\{#MyAppExeName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional icons:"; Flags: unchecked

[Files]
Source: "$bundleDir\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent
"@

Set-Content -Path $issPath -Value $issContent -Encoding UTF8

& $iscc $issPath
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}

$installerPath = Join-Path $installerDir "$OutputBaseFilename.exe"
if (-not (Test-Path $installerPath)) {
  throw "Build finished, but installer was not found at expected path: $installerPath"
}

Write-Host "Installer built successfully:"
Write-Host $installerPath
