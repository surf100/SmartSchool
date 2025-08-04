$RootDir = Get-Location
$ProtoDir = "$RootDir\proto"
$OutDir = "$ProtoDir\gen"
$ThirdParty = "$RootDir\third_party"

New-Item -ItemType Directory -Force -Path $OutDir

protoc -I="$ProtoDir" -I="$ThirdParty" `
  --go_out="$OutDir" --go_opt=paths=source_relative `
  --go-grpc_out="$OutDir" --go-grpc_opt=paths=source_relative `
  --grpc-gateway_out="$OutDir" --grpc-gateway_opt=paths=source_relative `
  "$ProtoDir\terminal_event.proto"
