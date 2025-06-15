# GoSearchEngine

Un moteur de recherche simple, robuste et concurrentiel, développé en Go. Il est conçu pour "crawler" le web, indexer le contenu en utilisant des algorithmes comme TF-IDF et PageRank, et fournir une interface de recherche via une API REST.

## Table des matières

- [À propos](#à-propos)
- [Fonctionnalités](#fonctionnalités)
- [Architecture et Flux de Travail](#architecture-et-flux-de-travail)
- [Structure du projet](#structure-du-projet)
- [Configuration requise](#configuration-requise)
- [Installation](#installation)
- [Configuration](#configuration)
- [Exécution du Moteur de Recherche](#exécution-du-moteur-de-recherche)
- [API de Recherche](#api-de-recherche)
- [Contribution](#contribution)
- [Licence](#licence)
- [Auteur](#auteur)

## À propos

Ce projet est une implémentation complète d'un moteur de recherche. Il gère l'ensemble du cycle, de la récupération des pages web à leur indexation en passant par le calcul de scores de pertinence pour la recherche. Il met l'accent sur la concurrence pour des performances optimales et utilise MongoDB pour le stockage persistant des données.

## Fonctionnalités

### Web Crawler Concurrentiel
- Explore le web en utilisant un pool de "workers" (goroutines)
- Implémente une gestion de file d'attente robuste via des canaux Go (urlChanSender, urlChanReceiver)
- Respecte un délai minimum configurable entre les requêtes vers le même domaine pour éviter la surcharge des serveurs
- Respecte les règles robots.txt pour chaque domaine
- Limite configurable du nombre de pages à "crawler" (MaxSizePage)
- Gère la taille de la file d'attente d'URL (MaxSizeQueue)

### Analyse HTML
Extrait le contenu textuel nettoyé et les liens internes/externes des pages web.

### Stockage Persistant
Utilise MongoDB pour stocker les pages "crawlées", les URL en file d'attente, et les données d'indexation (mots-clés, scores).

### Indexation Avancée
- **TF-IDF** (Term Frequency-Inverse Document Frequency) : Calcule la pertinence des mots dans les documents
- **PageRank** : Implémente l'algorithme PageRank pour évaluer l'importance des pages en fonction de leur structure de liens
- **Scoring Combiné** : Combine les scores TF-IDF et PageRank pour obtenir un score de pertinence final pour la recherche

### API RESTful
Fournit un point de terminaison de recherche (/) via le framework Gin Gonic, permettant d'interroger l'index.

## Architecture et Flux de Travail

Le moteur de recherche suit un processus en plusieurs étapes :

1. **Initialisation du Crawler** : Une file d'attente (Queue) est créée et initialisée avec une liste d'URLs de départ (sites francophones pré-définis).

2. **Gestion de la File d'Attente (QueueHandler)** : Deux goroutines gèrent le flux des URLs :
    - Une goroutine récupère les URLs à "crawler" de la file d'attente (`q.GetUrl()`) et les envoie aux "workers" via `urlChanSender`
    - Une autre goroutine reçoit les nouvelles URLs découvertes par les "workers" via `urlChanReceiver` et les ajoute à la file d'attente principale (`q.AddUrl()`)

3. **Processus de "Crawling" (CrawlerProcess)** :
    - Plusieurs "workers" (définis par le nombre de goroutines lancées dans main) consomment les URLs depuis `urlChanSender`
    - Chaque "worker" :
        - Récupère la page web
        - Analyse son contenu et extrait les liens
        - Stocke la page dans la base de données MongoDB
        - Envoie les URLs nouvellement découvertes à `urlChanReceiver` pour être ajoutées à la file d'attente
    - Le "crawling" s'arrête lorsque `MaxSizePage` est atteint

4. **Calcul du PageRank** : Une fois le "crawling" terminé, l'algorithme PageRank est exécuté sur toutes les pages stockées pour évaluer leur importance.

5. **Calcul du TF-IDF** : Le processus TF-IDF est ensuite exécuté pour calculer la pertinence des mots dans chaque document. Les scores sont combinés avec le PageRank et stockés dans la base de données.

6. **Exposition de l'API de Recherche** : Un serveur HTTP est lancé avec Gin, exposant un endpoint pour effectuer des requêtes de recherche sur les données indexées.

## Structure du projet

```
GoSearchEngine/
├── api/                # Contient la logique des endpoints de l'API RESTful
│   └── api.go          # Gère la requête de recherche et interagit avec la base de données
├── crawler/
│   ├── crawler.go      # Logique principale du web crawler, gestion de la file d'attente, des workers et des domaines
│   └── (autres fichiers liés au crawler)
├── db/
│   ├── db.go           # Logique de connexion et d'interaction avec MongoDB (Collections, Index)
│   └── models.go       # Définitions des structures de données pour MongoDB (Page, WordPage)
├── indexation/
│   ├── page_rank.go    # Implémentation de l'algorithme PageRank
│   └── tf_idf.go       # Implémentation du calcul TF-IDF et de la gestion des mots vides
├── parser/
│   ├── parser.go       # Logique d'analyse HTML pour extraire le texte et les liens
│   └── robot_txt.go    # Logique pour analyser et respecter les fichiers robots.txt
├── test/               # Dossier pour les tests unitaires et d'intégration
├── utils/              # Dossier pour les fonctions utilitaires diverses
├── go.mod              # Fichier de module Go pour les dépendances
├── main.go             # Point d'entrée principal de l'application, orchestre les étapes
└── README.md           # Ce fichier
```

## Configuration requise

- **Go** : Version 1.16 ou ultérieure
- **MongoDB** : Une instance de MongoDB doit être en cours d'exécution et accessible (par défaut, localhost:27017)

## Installation

1. Cloner le dépôt :
```bash
git clone https://github.com/votre-utilisateur/GoSearchEngine.git
cd GoSearchEngine
```

2. Télécharger les dépendances Go :
```bash
go mod tidy
```

3. Démarrer MongoDB :
   Assurez-vous que votre serveur MongoDB est en cours d'exécution. Si ce n'est pas le cas, vous pouvez le démarrer via `mongod` ou votre gestionnaire de services préféré.

## Configuration

Les constantes importantes du "crawler" sont définies directement dans `crawler/crawler.go` :

```go
const MinTimeBetweenRequest = 10 * time.Second // Délai minimal entre les requêtes vers le même domaine
const MaxIterationForGetUrl = 100000           // Nombre maximal d'itérations pour GetUrl pour trouver une URL disponible
const MaxSizeQueue = 1000000                   // Taille maximale de la file d'attente d'URL du crawler
const MaxSizePage = 1000                       // Nombre maximal de pages à crawler et à stocker
```

Les paramètres de connexion à MongoDB sont définis dans `db/db.go` :

```go
var Config ConfigStruct = ConfigStruct{
    Host:     "localhost",
    Port:     27017,
    Database: "search_engine",
}
```

Vous pouvez modifier ces valeurs directement dans le code source pour ajuster le comportement du moteur de recherche.

## Exécution du Moteur de Recherche

Pour démarrer le processus complet (crawl, indexation, lancement de l'API), exécutez simplement `main.go` depuis le répertoire racine du projet :

```bash
go run main.go
```

Le programme va d'abord :
1. Initialiser le "crawler" avec une liste d'URLs francophones prédéfinies
2. Lancer 10 "workers" du "crawler" et les gestionnaires de la file d'attente
3. Attendre que le "crawling" soit terminé (soit que toutes les URLs aient été traitées, soit que MaxSizePage soit atteint)
4. Exécuter le calcul du PageRank
5. Exécuter le calcul du TF-IDF et des scores combinés
6. Lancer le serveur Gin sur le port 8080 pour l'API de recherche

Des messages de progression seront affichés dans la console pendant le "crawling".

## API de Recherche

Une fois le serveur démarré (généralement après que le "crawling" et l'indexation initiale soient terminés), vous pouvez effectuer des requêtes de recherche via HTTP GET sur le port 8080.

### Endpoint
**GET** `/`

### Paramètres de requête
- `q` (obligatoire) : La chaîne de caractères à rechercher
- `limit` (optionnel) : Le nombre maximum de résultats à renvoyer

### Exemple de requête avec curl
```bash
curl "http://localhost:8080/?q=intelligence%20artificielle&limit=5"
```

### Exemple de réponse (format JSON)
```json
[
  {
    "Url": "https://fr.wikipedia.org/wiki/Intelligence_artificielle",
    "Content": "...",
    "PageRank": 0.123,
    "Score": 0.567
  },
  {
    "Url": "https://www.futura-sciences.com/tech/intelligence-artificielle/",
    "Content": "...",
    "PageRank": 0.089,
    "Score": 0.451
  }
  // ... jusqu'à 'limit' résultats
]
```

## Auteur

Mazong Dongmo Manuel G. - https://www.linkedin.com/in/manuel-glory-mazong-dongmo-132b0a335/