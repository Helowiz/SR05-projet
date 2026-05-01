#!/bin/bash

echo "==> Compilation de l'application et du contrôleur..."
go build -o ../bin/app ../web/server.go
go build -o ../bin/ctl ../cmd/controler

if [ $? -ne 0 ]; then
    echo "Erreur de compilation. Arrêt du script."
    exit 1
fi