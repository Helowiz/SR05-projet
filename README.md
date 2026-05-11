# SR05-projet

## Contexte
Ce projet a été réalisé au semestre de printemps 2026 à l'Université de Technologie de Compiègne, dans le cadre de l'UV SR05 (Systèmes et Algorithmes Répartis), par :
- LAUNAY Eloise
- COMBLAT Maryne
- ARTUS Caio
- CHEVALERIAS Evan

## Sujet
Le projet consiste à implémenter un tableau blanc collaboratif en temps réel. Plusieurs utilisateurs, répartis sur différents sites, peuvent dessiner simultanément sur un même canvas JavaScript partagé. L'application permet ainsi de travailler à plusieurs sur une même surface de dessin tout en conservant un état cohérent entre les participants.

L'objectif pédagogique est de mettre en œuvre :
- la communication inter-processus,
- la synchronisation distribuée,
- la gestion de la concurrence,
- la capture d'état global (snapshot).

## Architecture du projet

### Vue d'ensemble
Un site est composé de trois briques principales :
- un contrôleur Go,
- une application Go (serveur web),
- une interface HTML/CSS/JavaScript.

Rôles :
- le contrôleur gère les messages distribués, la file d'attente répartie et la cohérence globale,
- l'application maintient l'état local du tableau blanc et relaie les mises à jour vers l'interface,
- l'interface permet les interactions utilisateur, notamment la création, le déplacement et le redimensionnement des formes.

Cette séparation rend le projet plus lisible et facilite le débogage : chaque responsabilité est isolée dans un composant dédié.

### Organisation des répertoires
- `cmd/controler/main.go` : logique du contrôleur (section critique, messages, synchronisation)
- `web/server.go` : serveur applicatif et WebSocket
- `web/client.html`, `web/scripts.js`, `web/style.css` : interface utilisateur
- `protocol/message.go` : format et parsing des messages
- `shape/shape.go` : modèles de formes et opérations de dessin
- `snapshot/snapshot.go` : structures et sauvegarde des snapshots
- `display/display.go` : affichage de logs et traces
- `scripts/*.sh` : scripts de lancement selon la topologie

## Exécution

### Prérequis
- Go installé
- Bash (Linux, macOS ou WSL)

### Lancement standard
Depuis le dossier `scripts` :

```bash
./run.sh
```

Le script :
1. compile `app` et `ctl`,
2. lance la topologie choisie (ligne de script décommentée dans `run.sh`).

La topologie activée par défaut dans `run.sh` est `anneau.sh`.

### Topologies réseau disponibles

#### 1) `anneau.sh` (anneau unidirectionnel, 3 sites)
Arcs de communication contrôleur vers contrôleur :
- C1 -> C2
- C2 -> C3
- C3 -> C1

Cette topologie forme une boucle simple : chaque site transmet les messages au suivant, ce qui permet de tester la propagation circulaire des requêtes.

Schéma :

```text
    +-------> C2 ------> +
    |                    |
    |                    v
    C1 <------- C3 <-----+
```

#### 2) `chaine.sh` (chaîne orientée, 3 sites)
Arcs :
- C1 -> C2
- C2 -> C1
- C2 -> C3
- C3 -> C2

Cette configuration représente une chaîne de sites consécutifs. Elle permet d'observer un comportement intermédiaire entre le réseau linéaire et le réseau complet.

Schéma :

```text
C1 <------> C2 <------> C3
```

#### 3) `complet.sh` (graphe complet, 3 sites)
Chaque contrôleur envoie aux deux autres.

Le réseau complet maximise la connectivité entre les contrôleurs et sert à valider le fonctionnement du protocole lorsque tous les sites sont directement joignables.

Schéma :

```text
     C1
     ^ ^
     /  \
    v    v
   C2 <-> C3

(liens bidirectionnels entre toutes les paires)
```

#### 4) `quelconque.sh` (graphe orienté, 4 sites)
Arcs :
- C1 -> C2 et C3
- C2 -> C4
- C3 -> C2
- C4 -> C1

Cette topologie illustre un réseau plus asymétrique, utile pour tester des chemins de propagation non uniformes.

Schéma :

```text
      +-------> C2 --------> C4
      |          ^            |
      |          |            |
     C1 -------> C3           |
      ^                       |
      +-----------------------+
```

### Exécution sur plusieurs machines
Le script `remote_anneau.sh` permet un lancement distribué sur plusieurs machines, selon une topologie en anneau.

Points importants :
- renseigner l'adresse IP distante dans `run.sh` à l'endroit prévu,
- utiliser le rôle `mac` pour la machine macOS,
- utiliser le rôle `wsl` pour les machines Linux/WSL,
- prévoir une seule machine macOS dans ce scénario.

## Architecture logicielle

### Communication
- Contrôleur <-> contrôleurs : messages horodatés (estampilles + horloge vectorielle)
- Application <-> contrôleur : demandes et fin de section critique, puis diffusion des données
- Interface <-> application : WebSocket pour les actions utilisateur et la synchronisation visuelle

Chaque échange est structuré pour limiter les ambiguïtés de parsing et garantir un traitement cohérent des événements.


## Fonctionnalités
### File d'attente répartie

Nous avons implémenté l’algorithme de file d’attente répartie pour l’exclusion mutuelle distribuée. Cela a été réalisé grâce à une map `map[int]EltMapFile`, où `EltMapFile` est une structure contenant l’horloge ainsi que le type de message reçu (requête, accusé de réception ou libération). L’utilisation d’une map au lieu d’une liste impose de vérifier que tous les sites sont bien présents dans la map avant d’entrer en section critique, même si nous possédons la plus petite estampille.

### Sauvegarde d'état global via snapshots
Nous avons implémenté l'algorithme d'instantané avec reconstitution de configuration vu en cours.
Cela permet d'avoir une sauvegarde globale de notre système répartie. Nous sauvegardons chaque état de chaque site en plus des messages présents dans les canaux.

Nous l'avons légèrement modifié, car nous gérons la synchronisation entre l'application et le contrôler d'un même site à l'aide d'un Buffer.
De plus, nous avons développé le fait d'effectuer plusieurs snapshots.

### Cohérence et synchronisation
#### Estampilles (horloge logique) pour ordonner les requêtes

Nous avons intégré les estampilles afin d’obtenir un ordre total sur les actions de notre système réparti. Elles ont été implémentées avec une structure Go `Estampille`, qui contient l’horloge logique ainsi que l’identifiant du site. L’identifiant de l’émetteur et la valeur de l’horloge étant envoyés dans les messages de contrôle, nous pouvons reconstruire leur estampille.

#### Horloge vectorielle pour la cohérence causale

Implémentation de l'horloge à travers un map[int]int. 
L'horloge est mise à jour à chaque message reçu par le controleur, puisque chaque site envoie sa propre horloge dans les messages. 
Elle est aussi mise à jour à l'emplacement du site qui la possède à chaque action. 
Elle sert pour tester la cohérence lors du snapshot, puisque chaque site n'est pas censé posséder une information plus avancée à propos d'un site dans son horloge que le site en question. 

### Interface graphique
L'interface a été développée en HTML/CSS/JavaScript.

Une première base a été générée avec l'aide de Claude (modèle Sonnet 4.6), puis adaptée et corrigée par l'équipe pour correspondre aux besoins du projet, notamment sur le redimensionnement et le déplacement des formes. Cette version initiale se trouve dans le fichier `web/interface_originale.html`.

L'interface a ensuite été nettoyée pour offrir une expérience plus lisible et plus stable, en particulier lors des modifications concurrentes.

L'interface permet :
- de se connecter à un port localhost,
- de visualiser les logs importants,
- d'afficher un état d'attente (fond rouge) lorsqu'une modification distante est en cours, ce qui bloque temporairement les interactions sur le canvas.

Elle joue ainsi un rôle central dans l'observation du comportement distribué du système.
