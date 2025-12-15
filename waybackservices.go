package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type WaybackResult struct {
	Service string
	URL     string
	Date    string
}

type WBConfig struct {
	Domain    string
	ListPath  string
	Providers []string
	Debug     bool
	Report    bool
}

func parseWBFlags() WBConfig {
	domain := flag.String("d", "", "Domaine unique à analyser, par ex: example.com")
	list := flag.String("l", "", "Fichier contenant une liste de domaines (un par ligne)")
	provider := flag.String("provider", "", "Liste de providers à inclure (ex: google,sharepoint). Vide = tous.")
	report := flag.Bool("report", false, "Générer un rapport HTML par domaine (nom basé sur le domaine et un timestamp)")
	debug := flag.Bool("debug", false, "Activer le mode debug (affiche les URLs Wayback et les erreurs)")
	flag.Parse()

	var providers []string
	if strings.TrimSpace(*provider) != "" {
		parts := strings.Split(*provider, ",")
		for _, p := range parts {
			p = strings.ToLower(strings.TrimSpace(p))
			if p == "" {
				continue
			}
			providers = append(providers, p)
		}
	}

	return WBConfig{
		Domain:    strings.TrimSpace(*domain),
		ListPath:  strings.TrimSpace(*list),
		Providers: providers,
		Report:    *report,
		Debug:     *debug,
	}
}

func readWBDomainsFromFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var domains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return domains, nil
}

type servicePattern struct {
	Service  string
	Pattern  string
	Provider string
}

func buildServicePatterns(domain string, providers []string) []servicePattern {
	var patterns []servicePattern

	if isProviderEnabled(providers, "google") {
		patterns = append(patterns, servicePattern{
			Service:  "google_sites",
			Pattern:  fmt.Sprintf("https://sites.google.com/a/%s/*", domain),
			Provider: "google",
		})
		patterns = append(patterns, servicePattern{
			Service:  "google_docs",
			Pattern:  fmt.Sprintf("https://docs.google.com/a/%s/*", domain),
			Provider: "google",
		})
		patterns = append(patterns, servicePattern{
			Service:  "google_groups",
			Pattern:  fmt.Sprintf("https://groups.google.com/a/%s/*", domain),
			Provider: "google",
		})
		patterns = append(patterns, servicePattern{
			Service:  "google_drive",
			Pattern:  fmt.Sprintf("https://drive.google.com/a/%s/*", domain),
			Provider: "google",
		})
		patterns = append(patterns, servicePattern{
			Service:  "google_mail",
			Pattern:  fmt.Sprintf("https://mail.google.com/a/%s/*", domain),
			Provider: "google",
		})
		patterns = append(patterns, servicePattern{
			Service:  "google_sheets",
			Pattern:  fmt.Sprintf("https://spreadsheets.google.com/a/%s/*", domain),
			Provider: "google",
		})

		for i := 0; i <= 8; i++ {
			patterns = append(patterns, servicePattern{
				Service:  fmt.Sprintf("google_sheets_legacy_%d", i),
				Pattern:  fmt.Sprintf("https://spreadsheets%d.google.com/a/%s/*", i, domain),
				Provider: "google",
			})
		}
	}

	if isProviderEnabled(providers, "sharepoint") {
		patterns = append(patterns, servicePattern{
			Service:  "sharepoint_sites",
			Pattern:  fmt.Sprintf("https://*.sharepoint.com/*%s*", domain),
			Provider: "sharepoint",
		})
		patterns = append(patterns, servicePattern{
			Service:  "sharepoint_onedrive",
			Pattern:  fmt.Sprintf("https://*.sharepoint.com/personal/*%s*", domain),
			Provider: "sharepoint",
		})
	}

	if isProviderEnabled(providers, "onedrive") {
		patterns = append(patterns, servicePattern{
			Service:  "onedrive_short",
			Pattern:  fmt.Sprintf("https://*.1drv.ms/*%s*", domain),
			Provider: "onedrive",
		})
	}

	if isProviderEnabled(providers, "dropbox") {
		patterns = append(patterns, servicePattern{
			Service:  "dropbox_shares",
			Pattern:  fmt.Sprintf("https://www.dropbox.com/s/*%s*", domain),
			Provider: "dropbox",
		})
		patterns = append(patterns, servicePattern{
			Service:  "dropbox_folders",
			Pattern:  fmt.Sprintf("https://www.dropbox.com/sh/*%s*", domain),
			Provider: "dropbox",
		})
	}

	if isProviderEnabled(providers, "box") {
		patterns = append(patterns, servicePattern{
			Service:  "box_shares",
			Pattern:  fmt.Sprintf("https://*.box.com/s/*%s*", domain),
			Provider: "box",
		})
		patterns = append(patterns, servicePattern{
			Service:  "box_folders",
			Pattern:  fmt.Sprintf("https://*.box.com/folder/*%s*", domain),
			Provider: "box",
		})
	}

	if isProviderEnabled(providers, "github") {
		patterns = append(patterns, servicePattern{
			Service:  "github_repos",
			Pattern:  fmt.Sprintf("https://github.com/*%s*/*", domain),
			Provider: "github",
		})
		patterns = append(patterns, servicePattern{
			Service:  "github_raw",
			Pattern:  fmt.Sprintf("https://raw.githubusercontent.com/*%s*/*", domain),
			Provider: "github",
		})
	}

	if isProviderEnabled(providers, "gitlab") {
		patterns = append(patterns, servicePattern{
			Service:  "gitlab_repos",
			Pattern:  fmt.Sprintf("https://gitlab.com/*%s*/*", domain),
			Provider: "gitlab",
		})
	}

	if isProviderEnabled(providers, "atlassian") {
		patterns = append(patterns, servicePattern{
			Service:  "confluence",
			Pattern:  fmt.Sprintf("https://*.atlassian.net/wiki/*%s*", domain),
			Provider: "atlassian",
		})
		patterns = append(patterns, servicePattern{
			Service:  "jira",
			Pattern:  fmt.Sprintf("https://*.atlassian.net/browse/*%s*", domain),
			Provider: "atlassian",
		})
	}

	if isProviderEnabled(providers, "notion") {
		patterns = append(patterns, servicePattern{
			Service:  "notion",
			Pattern:  fmt.Sprintf("https://www.notion.so/*%s*", domain),
			Provider: "notion",
		})
	}

	if isProviderEnabled(providers, "slack") {
		patterns = append(patterns, servicePattern{
			Service:  "slack_files",
			Pattern:  fmt.Sprintf("https://files.slack.com/*%s*", domain),
			Provider: "slack",
		})
		patterns = append(patterns, servicePattern{
			Service:  "slack_workspace",
			Pattern:  fmt.Sprintf("https://%s.slack.com/*", domain),
			Provider: "slack",
		})
	}

	if isProviderEnabled(providers, "trello") {
		patterns = append(patterns, servicePattern{
			Service:  "trello_boards",
			Pattern:  fmt.Sprintf("https://trello.com/b/*%s*", domain),
			Provider: "trello",
		})
	}

	if isProviderEnabled(providers, "azure") {
		patterns = append(patterns, servicePattern{
			Service:  "azure_blob",
			Pattern:  fmt.Sprintf("https://%s.blob.core.windows.net/*", domain),
			Provider: "azure",
		})
		patterns = append(patterns, servicePattern{
			Service:  "azure_sites",
			Pattern:  fmt.Sprintf("https://%s.azurewebsites.net/*", domain),
			Provider: "azure",
		})
	}

	if isProviderEnabled(providers, "s3") {
		patterns = append(patterns, servicePattern{
			Service:  "aws_s3_bucket",
			Pattern:  fmt.Sprintf("https://%s.s3.amazonaws.com/*", domain),
			Provider: "s3",
		})
		patterns = append(patterns, servicePattern{
			Service:  "aws_s3_website",
			Pattern:  fmt.Sprintf("https://%s.s3-website.*.amazonaws.com/*", domain),
			Provider: "s3",
		})
	}

	if isProviderEnabled(providers, "gcs") {
		patterns = append(patterns, servicePattern{
			Service:  "gcs_bucket",
			Pattern:  fmt.Sprintf("https://storage.googleapis.com/%s/*", domain),
			Provider: "gcs",
		})
	}

	if isProviderEnabled(providers, "bitbucket") {
		patterns = append(patterns, servicePattern{
			Service:  "bitbucket_repos",
			Pattern:  fmt.Sprintf("https://bitbucket.org/*%s*/*", domain),
			Provider: "bitbucket",
		})
	}

	if isProviderEnabled(providers, "paste") {
		patterns = append(patterns, servicePattern{
			Service:  "pastebin",
			Pattern:  fmt.Sprintf("https://pastebin.com/*%s*", domain),
			Provider: "paste",
		})
		patterns = append(patterns, servicePattern{
			Service:  "hastebin",
			Pattern:  fmt.Sprintf("https://hastebin.com/*%s*", domain),
			Provider: "paste",
		})
	}

	if isProviderEnabled(providers, "calendar") {
		patterns = append(patterns, servicePattern{
			Service:  "google_calendar",
			Pattern:  fmt.Sprintf("https://calendar.google.com/calendar/*%s*", domain),
			Provider: "calendar",
		})
	}

	if isProviderEnabled(providers, "zoom") {
		patterns = append(patterns, servicePattern{
			Service:  "zoom_meetings",
			Pattern:  fmt.Sprintf("https://%s.zoom.us/*", domain),
			Provider: "zoom",
		})
	}

	if isProviderEnabled(providers, "figma") {
		patterns = append(patterns, servicePattern{
			Service:  "figma_files",
			Pattern:  fmt.Sprintf("https://www.figma.com/file/*%s*", domain),
			Provider: "figma",
		})
	}

	if isProviderEnabled(providers, "firebase") {
		patterns = append(patterns, servicePattern{
			Service:  "firebaseio_db",
			Pattern:  fmt.Sprintf("https://%s.firebaseio.com/*", domain),
			Provider: "firebase",
		})
		patterns = append(patterns, servicePattern{
			Service:  "firebase_storage",
			Pattern:  fmt.Sprintf("https://firebasestorage.googleapis.com/*%s*", domain),
			Provider: "firebase",
		})
	}

	return patterns
}

func isProviderEnabled(providers []string, provider string) bool {
	if provider == "" {
		return true
	}
	if len(providers) == 0 {
		return true
	}
	for _, p := range providers {
		if p == provider {
			return true
		}
	}
	return false
}

type wbEntry struct {
	Date string
	URL  string
}

func fetchWayback(pattern string, debug bool) ([]wbEntry, error) {
	escaped := url.QueryEscape(pattern)
	endpoint := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&output=json&collapse=urlkey", escaped)

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Wayback request: %s\n", endpoint)
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Wayback HTTP error for pattern %s: %v\n", pattern, err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Wayback read error for pattern %s: %v\n", pattern, err)
		return nil, err
	}

	var raw [][]string
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Wayback JSON error for pattern %s: %v\n", pattern, err)
		return nil, err
	}

	var entries []wbEntry
	skip := true
	for _, row := range raw {
		if skip {
			skip = false
			continue
		}
		if len(row) < 3 {
			continue
		}
		entries = append(entries, wbEntry{
			Date: row[1],
			URL:  row[2],
		})
	}
	return entries, nil
}

func runForDomainWayback(domain string, providers []string, debug bool) ([]WaybackResult, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, fmt.Errorf("domaine vide")
	}

	patterns := buildServicePatterns(domain, providers)

	var results []WaybackResult
	total := len(patterns)
	for i, p := range patterns {
		fmt.Fprintf(os.Stderr, "[%d/%d] %s\n", i+1, total, p.Service)
		entries, err := fetchWayback(p.Pattern, debug)
		if err != nil {
			if !debug {
				fmt.Fprintf(os.Stderr, "[WARN] Wayback error for service %s (pattern %s)\n", p.Service, p.Pattern)
			}
		} else {
			for _, e := range entries {
				results = append(results, WaybackResult{
					Service: p.Service,
					URL:     e.URL,
					Date:    e.Date,
				})
			}
		}
		time.Sleep(1200 * time.Millisecond)
	}

	return results, nil
}

func printWaybackResults(domain string, results []WaybackResult) {
	fmt.Println("====================================================")
	fmt.Println(" Wayback Services Hunting pour le domaine:", domain)
	fmt.Println("====================================================")
	if len(results) == 0 {
		fmt.Println("Aucune URL archivée trouvée pour les patterns configurés.")
		return
	}

	groups := make(map[string][]WaybackResult)
	for _, r := range results {
		groups[r.Service] = append(groups[r.Service], r)
	}

	services := make([]string, 0, len(groups))
	for s := range groups {
		services = append(services, s)
	}
	sortStrings(services)

	for _, s := range services {
		fmt.Println()
		fmt.Println("[" + s + "]")
		for _, r := range groups[s] {
			if r.Date != "" {
				fmt.Printf("%s %s\n", formatWBDate(r.Date), r.URL)
			} else {
				fmt.Println(r.URL)
			}
		}
	}
}

func formatWBDate(raw string) string {
	if len(raw) != 14 {
		return raw
	}
	t, err := time.Parse("20060102150405", raw)
	if err != nil {
		return raw
	}
	return t.Format(time.RFC3339)
}

func sortStrings(a []string) {
	if len(a) < 2 {
		return
	}
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

func writeWaybackHTMLReport(path string, domain string, results []WaybackResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	now := time.Now().Format(time.RFC3339)

	b := &strings.Builder{}
	b.WriteString("<!DOCTYPE html><html lang=\"fr\"><head><meta charset=\"UTF-8\"><title>Wayback Services Hunting - ")
	b.WriteString(escapeHTML(domain))
	b.WriteString("</title><style>")
	b.WriteString("body{font-family:Arial,sans-serif;background:#020617;color:#e5e7eb;padding:20px}h1,h2{color:#fbbf24}table{border-collapse:collapse;width:100%;margin-bottom:24px}th,td{border:1px solid #1f2937;padding:6px 8px;font-size:13px}th{background:#0f172a}tr:nth-child(even){background:#020617}tr:nth-child(odd){background:#020617}code{background:#111827;padding:2px 4px;border-radius:4px;color:#93c5fd}")
	b.WriteString("</style></head><body>")

	b.WriteString("<h1>Wayback Services Hunting - ")
	b.WriteString(escapeHTML(domain))
	b.WriteString("</h1>")
	b.WriteString("<p>Généré le ")
	b.WriteString(escapeHTML(now))
	b.WriteString("</p>")

	if len(results) == 0 {
		b.WriteString("<p>Aucune URL archivée trouvée pour les patterns configurés.</p>")
		b.WriteString("</body></html>")
		_, err = f.WriteString(b.String())
		return err
	}

	grouped := make(map[string][]WaybackResult)
	for _, r := range results {
		grouped[r.Service] = append(grouped[r.Service], r)
	}

	services := make([]string, 0, len(grouped))
	for s := range grouped {
		services = append(services, s)
	}
	sortStrings(services)

	for _, s := range services {
		b.WriteString("<h2>")
		b.WriteString(escapeHTML(s))
		b.WriteString("</h2>")
		b.WriteString("<table><thead><tr><th>Date</th><th>URL archivée</th></tr></thead><tbody>")
		for _, r := range grouped[s] {
			b.WriteString("<tr><td>")
			if r.Date != "" {
				b.WriteString(escapeHTML(formatWBDate(r.Date)))
			} else {
				b.WriteString("-")
			}
			b.WriteString("</td><td><a href=\"")
			b.WriteString(escapeHTML(r.URL))
			b.WriteString("\" target=\"_blank\" rel=\"noopener noreferrer\">")
			b.WriteString(escapeHTML(r.URL))
			b.WriteString("</a></td></tr>")
		}
		b.WriteString("</tbody></table>")
	}

	b.WriteString("</body></html>")

	_, err = f.WriteString(b.String())
	return err
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	return s
}

func escapeHTML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}

func main() {
	cfg := parseWBFlags()

	if cfg.Domain == "" && cfg.ListPath == "" {
		fmt.Println("Usage :")
		fmt.Println("  ./waybackservices -d example.com")
		fmt.Println("  ./waybackservices -l list.txt")
		fmt.Println()
		fmt.Println("Options :")
		fmt.Println("  -provider p1,p2,p3   Limiter aux providers spécifiés (par défaut : tous)")
		fmt.Println("  -report              Générer un rapport HTML par domaine (wayback-<domaine>-<timestamp>.html)")
		fmt.Println("  -debug               Mode verbeux, affiche les URLs Wayback et les erreurs")
		fmt.Println()
		fmt.Println("Providers disponibles :")
		fmt.Println("  google      sites/docs/drive/groups/mail/sheets/spreadsheets0-8")
		fmt.Println("  sharepoint  sharepoint online + onedrive perso")
		fmt.Println("  onedrive    liens courts 1drv.ms")
		fmt.Println("  dropbox     liens de partage/folders Dropbox")
		fmt.Println("  box         liens de partage/folders Box")
		fmt.Println("  github      repos + raw.githubusercontent.com")
		fmt.Println("  gitlab      repos GitLab")
		fmt.Println("  bitbucket   repos Bitbucket")
		fmt.Println("  atlassian   Confluence + Jira (*.atlassian.net)")
		fmt.Println("  notion      pages Notion")
		fmt.Println("  slack       fichiers Slack + workspace <domaine>.slack.com")
		fmt.Println("  trello      boards Trello")
		fmt.Println("  azure       blob storage + sites azurewebsites")
		fmt.Println("  s3          buckets/website S3")
		fmt.Println("  gcs         buckets Google Cloud Storage")
		fmt.Println("  firebase    *.firebaseio.com + Firebasestorage")
		fmt.Println("  paste       pastebin + hastebin")
		fmt.Println("  calendar    Google Calendar publics liés au domaine")
		fmt.Println("  zoom        sous-domaine <domaine>.zoom.us")
		fmt.Println("  figma       fichiers Figma")
		os.Exit(1)
	}

	if cfg.Domain != "" && cfg.ListPath != "" {
		fmt.Println("Veuillez utiliser soit -d soit -l, mais pas les deux en même temps")
		os.Exit(1)
	}

	if cfg.Domain != "" {
		results, err := runForDomainWayback(cfg.Domain, cfg.Providers, cfg.Debug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erreur pour le domaine %s: %v\n", cfg.Domain, err)
			os.Exit(1)
		}
		printWaybackResults(cfg.Domain, results)
		if cfg.Report {
			ts := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("wayback-%s-%s.html", sanitizeFilename(cfg.Domain), ts)
			err = writeWaybackHTMLReport(filename, cfg.Domain, results)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erreur lors de l'écriture du rapport HTML pour %s: %v\n", cfg.Domain, err)
			} else {
				fmt.Fprintf(os.Stderr, "Rapport HTML généré: %s\n", filename)
			}
		}
		return
	}

	domains, err := readWBDomainsFromFile(cfg.ListPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Impossible de lire la liste de domaines: %v\n", err)
		os.Exit(1)
	}
	if len(domains) == 0 {
		fmt.Fprintf(os.Stderr, "Aucun domaine valide trouvé dans %s\n", cfg.ListPath)
		os.Exit(1)
	}

	for _, d := range domains {
		results, err := runForDomainWayback(d, cfg.Providers, cfg.Debug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erreur pour le domaine %s: %v\n", d, err)
			continue
		}
		printWaybackResults(d, results)
		if cfg.Report {
			ts := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("wayback-%s-%s.html", sanitizeFilename(d), ts)
			err = writeWaybackHTMLReport(filename, d, results)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erreur lors de l'écriture du rapport HTML pour %s: %v\n", d, err)
			} else {
				fmt.Fprintf(os.Stderr, "Rapport HTML généré: %s\n", filename)
			}
		}
		fmt.Println()
	}
}


