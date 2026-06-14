go clean -cache
go build -o ./bin/net ../main
go build -o ./bin/log ../logweb
go build -o ./bin/netsrv ../web