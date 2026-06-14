nettoyer() {
    kill -- -$$
}

trap 'nettoyer; exit 0' INT QUIT TERM


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

# Caio : je propose qu'on build plus NET dcp on ramène l'executable tout fait
# if ! go build -o bin/net ../net; then
#     echo "Erreur de compilation pour net"
#     exit 1
# fi

# if ! go build -o bin/logweb ../net/web; then
#     echo "Erreur de compilation pour log"
#     exit 1
# fi

echo "Compilation réussie."

./config_etude_easy.sh ./bin/app ./bin/ctl ./bin/net ./bin/logweb
