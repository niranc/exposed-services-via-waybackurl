# WaybackServices

Outil en Go pour la chasse aux bugs et à l’exposition d’un domaine via la **Wayback Machine**, en se concentrant sur les principaux services en ligne (Google, Microsoft, GitHub, Slack, etc.).

## Build

```bash
cd /home/nira/Code/GoogleServices
go build -o waybackservices waybackservices.go
```

## Usage de base

- Scanner un seul domaine :

```bash
./waybackservices -d example.com
```

- Scanner une liste de domaines (un domaine par ligne dans `list.txt`) :

```bash
./waybackservices -l list.txt
```

Sans option `-provider`, tous les providers connus sont utilisés.

## Providers

Limiter la recherche à certains services :

```bash
./waybackservices -d example.com -provider google,github,sharepoint
```

Providers disponibles (principaux) :

- `google`      : Sites / Docs / Groups / Drive / Mail / Sheets / spreadsheets0-8
- `sharepoint`  : SharePoint Online + OneDrive perso
- `onedrive`    : liens courts `1drv.ms`
- `dropbox`     : partages / dossiers Dropbox
- `box`         : partages / dossiers Box
- `github`      : repos + `raw.githubusercontent.com`
- `gitlab`      : repos GitLab
- `bitbucket`   : repos Bitbucket
- `atlassian`   : Confluence + Jira (`*.atlassian.net`)
- `notion`      : pages Notion
- `slack`       : `files.slack.com` + `<domaine>.slack.com`
- `trello`      : boards Trello
- `azure`       : `*.blob.core.windows.net` + `*.azurewebsites.net`
- `s3`          : buckets / sites S3
- `gcs`         : buckets Google Cloud Storage
- `firebase`    : `*.firebaseio.com` + Firebase Storage
- `paste`       : Pastebin + Hastebin
- `calendar`    : Google Calendar publics
- `zoom`        : `<domaine>.zoom.us`
- `figma`       : fichiers Figma

## Rapport HTML (`-report`)

Générer automatiquement un rapport HTML par domaine, nommé avec le domaine et un timestamp :

- Un domaine :

```bash
./waybackservices -d example.com -provider google,github -report
```

- Liste de domaines :

```bash
./waybackservices -l list.txt -provider google,github -report
```

Pour chaque domaine, un fichier est créé dans le répertoire courant :

```text
wayback-<domaine>-<timestamp>.html
```

Chaque rapport contient :

- Un en‑tête avec le domaine et la date de génération
- Un tableau par service (`google_docs`, `github_repos`, etc.) listant :
  - la date Wayback
  - l’URL archivée (cliquable)

## Debug et rate limit

- `-debug` : mode verbeux qui affiche :
  - les URLs d’API Wayback appelées
  - les erreurs HTTP / lecture / JSON

```bash
./waybackservices -d example.com -provider google -debug
```

- Les requêtes vers Wayback sont **séquentielles** avec une petite pause entre chaque pattern pour limiter le rate limit.
- Une progression simple est affichée sur `stderr` :

```text
[1/15] google_sites
[2/15] google_docs
...
```


# exposed-services-via-waybackurl
