# WaybackServices

`waybackservices` is a Go tool to hunt for exposed information about a target domain using the **Wayback Machine**, focusing on major online services (Google, Microsoft, GitHub, Slack, etc.).

---

## Build

go build -o waybackservices waybackservices.go---

## Basic usage

- **Single domain**:

./waybackservices -d example.com- **List of domains** (one per line in `list.txt`):

./waybackservices -l list.txtIf you don’t specify `-provider`, all known providers are used.

---

## Providers

You can restrict the scan to one or more providers (comma‑separated):

./waybackservices -d example.com -provider google,github,sharepointSupported providers (main ones):

- `google`      – Sites / Docs / Groups / Drive / Mail / Sheets / spreadsheets0–8
- `sharepoint`  – SharePoint Online + personal OneDrive
- `onedrive`    – short links on `1drv.ms`
- `dropbox`     – shared links / folders
- `box`         – Box shared links / folders
- `github`      – repos + `raw.githubusercontent.com`
- `gitlab`      – GitLab repos
- `bitbucket`   – Bitbucket repos
- `atlassian`   – Confluence + Jira (`*.atlassian.net`)
- `notion`      – Notion pages
- `slack`       – `files.slack.com` + `<domain>.slack.com`
- `trello`      – Trello boards
- `azure`       – `*.blob.core.windows.net` + `*.azurewebsites.net`
- `s3`          – S3 buckets / website endpoints
- `gcs`         – Google Cloud Storage buckets
- `firebase`    – `*.firebaseio.com` + Firebase Storage
- `paste`       – Pastebin + Hastebin
- `calendar`    – public Google Calendars
- `zoom`        – `<domain>.zoom.us`
- `figma`       – Figma files

---

## HTML report (`-report`)

Enable automatic HTML report generation per domain:

- **Single domain**:

./waybackservices -d example.com -provider google,github -report- **List of domains**:

./waybackservices -l list.txt -provider google,github -reportFor each domain, an HTML file is created in the current directory:

wayback-<domain>-<timestamp>.htmlEach report includes:

- A header with the domain and generation date.
- One table per service (e.g. `google_docs`, `github_repos`) listing:
  - the Wayback capture date,
  - the archived URL (clickable).

---

## Debug & rate limiting

- **Debug mode**:

./waybackservices -d example.com -provider google -debugThis prints:

- The Wayback API URLs being requested.
- HTTP / read / JSON errors (with `[WARN]` messages).

- To be nicer to Wayback and reduce rate limiting:
  - Requests are sent **sequentially**, not in parallel.
  - A small delay is added between each pattern query.

During a run, a simple progress indicator is printed to `stderr`, for example:

[1/15] google_sites
[2/15] google_docs
...
