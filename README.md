# WaybackServices

`waybackservices` is a Go tool to hunt for exposed information about a target domain via the **Wayback Machine**, focusing on major online services (Google, Microsoft, GitHub, Slack, etc.).

---

## Build

```sh
go build -o waybackservices waybackservices.go
```

---

## Usage

- **Single domain:**

```sh
./waybackservices -d example.com
```

- **List of domains** (one domain per line in `list.txt`):

```sh
./waybackservices -l list.txt
```

If you do not specify the `-provider` option, all supported providers will be scanned.

---

## Providers

You can restrict the scan to one or more providers (comma-separated):

```sh
./waybackservices -d example.com -provider google,github,sharepoint
```

**Main supported providers:**

- `google`      – Sites, Docs, Groups, Drive, Mail, Sheets, spreadsheets0–8
- `sharepoint`  – SharePoint Online + personal OneDrive
- `onedrive`    – short links on `1drv.ms`
- `dropbox`     – shared Dropbox links and folders
- `box`         – Box shared links and folders
- `github`      – repositories + `raw.githubusercontent.com`
- `gitlab`      – GitLab repositories
- `bitbucket`   – Bitbucket repositories
- `atlassian`   – Confluence + Jira (`*.atlassian.net`)
- `notion`      – Notion public pages
- `slack`       – `files.slack.com` + `<domain>.slack.com`
- `trello`      – Trello boards
- `azure`       – `*.blob.core.windows.net` + `*.azurewebsites.net`
- `s3`          – S3 buckets / website endpoints
- `gcs`         – Google Cloud Storage buckets
- `firebase`    – `*.firebaseio.com` + Firebase Storage files
- `paste`       – Pastebin and Hastebin
- `calendar`    – public Google Calendars
- `zoom`        – `<domain>.zoom.us`
- `figma`       – Figma files

---

## HTML report (`-report`)

Enable automatic HTML report generation per domain:

- **Single domain:**

```sh
./waybackservices -d example.com -provider google,github -report
```

- **List of domains:**

```sh
./waybackservices -l list.txt -provider google,github -report
```

For each domain, a HTML file will be created in the current directory:

```
wayback-<domain>-<timestamp>.html
```

Each report includes:

- A header with the domain and generation date.
- One table per service (e.g., `google_docs`, `github_repos`) listing:
  - the Wayback capture date,
  - the archived URL (clickable).

---

## Debug & Rate Limiting

- **Debug mode:**

```sh
./waybackservices -d example.com -provider google -debug
```

This will print:

- The queried Wayback API URLs.
- HTTP / read / JSON errors (with `[WARN]` messages).

- To be polite to Wayback and reduce rate limiting:
  - Requests are sent **sequentially**, not in parallel.
  - A small delay is added between each pattern query.

During a run, a simple progress indicator is printed to `stderr`, for example:

```
[1/15] google_sites
[2/15] google_docs
...
```

