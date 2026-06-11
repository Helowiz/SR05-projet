#!/bin/bash

nettoyer() {
    killall app 2> /dev/null
    killall ctl 2> /dev/null
    killall cat 2> /dev/null
    killall tee 2> /dev/null
    rm -f /tmp/fifo_*
}

trap 'nettoyer; exit 0' INT QUIT TERM

nettoyer

mkdir -p bin

echo "==> Compilation des programmes Go..."

if ! go build -o bin/app ../web/server.go; then
    echo "Erreur de compilation pour app"
    exit 1
fi

if ! go build -o bin/ctl ../cmd/controler/main.go; then
    echo "Erreur de compilation pour ctl"
    exit 1
fi

echo "Compilation réussie."

echo "==> Démarrage de l'anneau..."

# ./remote_anneau.sh ./bin/app ./bin/ctl <role> <ip>
./anneau.sh ./bin/app ./bin/ctl 
#./complet.sh ./bin/app ./bin/ctl
#./chaine.sh ./bin/app ./bin/ctl
#./quelconque.sh ./bin/app ./bin/ctl