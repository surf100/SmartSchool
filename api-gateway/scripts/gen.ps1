# scripts/gen.ps1
$RootDir    = Get-Location
$ProtoDir   = "$RootDir\proto"
$ThirdParty = "$RootDir\third_party"

# ===== terminal_event =====
$OutDirTerminal = "$ProtoDir\gen\terminalpb"
New-Item -ItemType Directory -Force -Path $OutDirTerminal | Out-Null

protoc -I="$ProtoDir" -I="$ThirdParty" `
  --go_out="$OutDirTerminal" --go_opt=paths=source_relative `
  --go-grpc_out="$OutDirTerminal" --go-grpc_opt=paths=source_relative `
  --grpc-gateway_out="$OutDirTerminal" --grpc-gateway_opt=paths=source_relative `
  "$ProtoDir\terminal_event.proto"

# ===== personsdumper =====
$OutDirPD = "$ProtoDir\gen\persondumperpb"
New-Item -ItemType Directory -Force -Path $OutDirPD | Out-Null

protoc -I="$ProtoDir" -I="$ThirdParty" `
  --go_out="$OutDirPD" --go_opt=paths=source_relative `
  --go-grpc_out="$OutDirPD" --go-grpc_opt=paths=source_relative `
  --grpc-gateway_out="$OutDirPD" --grpc-gateway_opt=paths=source_relative `
  "$ProtoDir\persondumper.proto"

Write-Host "✅ Generated terminalpb → $OutDirTerminal"
Write-Host "✅ Generated personsdumperpb → $OutDirPD"
