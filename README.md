# SR05-projet

## File d'attente répartie

L'implémentation de la file d'attente répartie dans ce projet permet aux différents sites de modifier les données tout en s'assurant que les sites ont toujours la bonne version des données.

L'algorithme que nous avons implémenté suit celui du cours cependant avec quelques différences.

### Communication control a control

Protocole de base avec /=cle=valeur
un message de control contieng TOUJOURS un champ hlg, un champ id et un champ msg
Si le message est a destinataire precis un champ target_id est egalement present
Dans le cas de liberation, un champ data avec les nouvelles donnees est ajoute

### Communication control a app
Utilise le même protocole mais que un message a la fois sous la forme 
/=type=(type du message)/=valeur=(valeur du message)

### Communication app a control
Utilise le protocole de base mais avec
/=type=(fromapp_debut_sc ou fromapp_fin_sc)
et dans le cas de fromapp_fin_sc on a en plus le champ /data=(donnees a jour)

### Communication app a interface et vice versa
On utilise just des pairs cle=valeur dans la websocket
c'est plus simple vu qu'on a pas a gerer autant de choses
A voir si on change ça