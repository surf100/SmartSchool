param(
  # Можно явно указать путь к person.proto (если знаешь)
  [string]$SrcProto = ""
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()

# 0) Корни и папки
$Root = Get-Location
$VendoredDir = Join-Path $Root "third_party\personapi"
$VendoredProtoDir = Join-Path $VendoredDir "proto"

# 1) Если путь не задан — пытаемся найти person.proto рядом в монорепо
if ([string]::IsNullOrWhiteSpace($SrcProto)) {
  $default = Join-Path $Root "..\person-api\proto\person.proto"
  if (Test-Path $default) {
    $SrcProto = $default
  } else {
    $candidates = Get-ChildItem -Path (Join-Path $Root "..") -Recurse -Filter "person.proto" -ErrorAction SilentlyContinue |
      Where-Object { Select-String -Path $_.FullName -Pattern 'service\s+PersonService' -Quiet } |
      Select-Object -ExpandProperty FullName
    if ($candidates -and $candidates.Count -gt 0) {
      $SrcProto = $candidates[0]
    }
  }
}

if (-not (Test-Path $SrcProto)) {
  throw "Не нашёл person.proto. Укажи путь параметром:  .\scripts\gen_personapi.ps1 -SrcProto 'C:\...\person-api\proto\person.proto'"
}

# 2) Вытащим module path из go.mod текущего сервиса
$goModPath = Join-Path $Root "go.mod"
if (-not (Test-Path $goModPath)) { throw "Не найден go.mod в $Root" }
$moduleLine = (Get-Content $goModPath | Where-Object { $_ -match '^\s*module\s+(.+)\s*$' })[0]
$modulePath = ($moduleLine -replace '^\s*module\s+','').Trim()

# 3) Готовим каталоги
New-Item -ItemType Directory -Force -Path $VendoredDir | Out-Null
New-Item -ItemType Directory -Force -Path $VendoredProtoDir | Out-Null

# 4) Копируем и правим go_package под наш модуль
$proto = Get-Content $SrcProto -Raw
$newGoPkg = "option go_package = `"$modulePath/third_party/personapi;personpb`";"

if ($proto -match 'option\s+go_package\s*=') {
  $proto = [regex]::Replace($proto, 'option\s+go_package\s*=\s*".*?";', $newGoPkg)
} else {
  $proto = [regex]::Replace($proto, '(^\s*package\s+.*?;\s*)', "`$1`r`n$newGoPkg`r`n")
}

$VendoredProto = Join-Path $VendoredProtoDir "person.proto"
Set-Content -LiteralPath $VendoredProto -Value $proto -NoNewline -Encoding UTF8

# 5) Проверим наличие генераторов
function Ensure-Cli($name) {
  $p = (Get-Command $name -ErrorAction SilentlyContinue)
  if (-not $p) { throw "Не найден $name в PATH. Установи: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" }
}
Ensure-Cli "protoc"
Ensure-Cli "protoc-gen-go"
Ensure-Cli "protoc-gen-go-grpc"

# 6) Генерация
$OutDir = $VendoredDir
protoc -I="$VendoredProtoDir" `
  --go_out="$OutDir" --go_opt=paths=source_relative `
  --go-grpc_out="$OutDir" --go-grpc_opt=paths=source_relative `
  "person.proto"

Write-Host "✅ Сгенерировано в: $OutDir"
Write-Host "👉 Импортируй так:  import personpb `"$modulePath/third_party/personapi`""
