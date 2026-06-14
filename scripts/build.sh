go clean -cache
echo "==> Compilation des programmes Go..."

if ! go build -o bin/app ../web/server.go; then
    echo "Erreur de compilation pour app"
    exit 1
fi

if ! go build -o bin/ctl ../cmd/controler/main.go; then
    echo "Erreur de compilation pour ctl"
    exit 1
fi

if ! go build -o bin/net ../net/main; then
     echo "Erreur de compilation pour net"
     exit 1
fi

if ! go build -o bin/netsrv ../net/web; then
    echo "Erreur de compilation pour log"
    exit 1
fi