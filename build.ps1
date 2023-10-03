$env:GOPATH = 'D:\DEV\GoWorkspace'

go build

Set-Content -Path "riscosrv_version.txt" -Value "$(git rev-parse --short HEAD)"