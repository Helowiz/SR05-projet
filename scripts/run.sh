#!/bin/bash

echo "==> Compilation des programmes Go..."
go build -o bin/writer ../cmd/writer
go build -o bin/controler ../cmd/controler
go build -o bin/reader ../cmd/reader

if [ $? -ne 0 ]; then
    echo "Erreur de compilation. Arrêt du script."
    exit 1
fi

nettoyer() {

    killall writer 2> /dev/null
    killall controler 2> /dev/null
    killall reader 2> /dev/null

    killall cat 2> /dev/null
    killall tee 2> /dev/null

    rm -f /tmp/fifo_*

    exit 0
}

trap nettoyer INT QUIT TERM

rm -f /tmp/fifo_*

mkfifo /tmp/fifo_W_vers_C
mkfifo /tmp/fifo_C_vers_R


# L'écrivain écrit (>) dans le premier tube
./bin/writer -n Auteur > /tmp/fifo_W_vers_C &

# Le contrôleur lit (<) le premier tube et écrit (>) dans le deuxième tube
./bin/controler -n Controle < /tmp/fifo_W_vers_C > /tmp/fifo_C_vers_R &

# Le lecteur lit (<) depuis le deuxième tube
./bin/reader -n Lecteur < /tmp/fifo_C_vers_R &

echo "=========================================================="
echo "   Réseau actif : Writer -> [fifo] -> Controler -> [fifo] -> Reader"
echo "   Appuyez sur Ctrl+C pour arrêter proprement tous les processus."
echo "=========================================================="

sleep 3600
nettoyer