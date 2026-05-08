Write-Host "Generating gRPC and Gateway code from proto/chat.proto..."

protoc `
  -I proto `
  --go_out=proto/gen --go_opt=paths=source_relative `
  --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative `
  --grpc-gateway_out=proto/gen --grpc-gateway_opt=paths=source_relative `
  proto/chat.proto

Write-Host "Generation complete."
