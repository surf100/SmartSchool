# scripts/gen.ps1
$RootDir = Get-Location
$ProtoDir = "$RootDir\proto"
$OutDir = "$RootDir\proto\gen"

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

protoc -I="$ProtoDir" `
  --go_out="$OutDir" --go_opt=paths=source_relative `
  --go-grpc_out="$OutDir" --go-grpc_opt=paths=source_relative `
  "personsdumper.proto"

Write-Host "Сгенерировано в $OutDir"
