package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// Endpoints backing the interactive previews. They are part of the site, not of
// any shipped component - a consumer writes their own equivalents.

func (s *server) routeDemos(mux *http.ServeMux) {
	mux.HandleFunc("POST /demo/echo", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond) // long enough to see the spinner
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /demo/input/validate", handleValidate)
	mux.HandleFunc("GET /demo/tabs/{tab}", handleTabs)
	mux.HandleFunc("GET /demo/dialog", handleDialogBody)
	mux.HandleFunc("POST /demo/toast", handleToast)
	mux.HandleFunc("GET /demo/progress", handleProgress)
}

func writeHTML(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, s)
}

// handleValidate returns the whole field back, with the verdict applied.
func handleValidate(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("username"))

	var problem string
	switch {
	case name == "":
		problem = ""
	case len(name) < 3:
		problem = "Needs at least 3 characters."
	case name != strings.ToLower(name):
		problem = "Lowercase only."
	case name == "claude":
		problem = "That one's taken."
	}

	border, message := "border-border", ""
	if problem != "" {
		border = "border-destructive"
		message = fmt.Sprintf(`<p class="text-caption text-destructive">%s</p>`, template.HTMLEscapeString(problem))
	} else if name != "" {
		message = `<p class="text-caption text-[var(--success-fg)]">Available.</p>`
	}

	writeHTML(w, fmt.Sprintf(`<div id="username-field" class="flex flex-col gap-1.5">
  <label class="text-caption font-500 text-foreground">Username </label>
  <input id="username" name="username" value="%s" placeholder="lowercase, 3+ characters"
         class="rounded-md border %s bg-background px-control-x py-control-y text-body-sm leading-none text-foreground placeholder:text-muted-foreground/70"
         hx-post="/demo/input/validate"
         hx-trigger="input changed delay:400ms"
         hx-target="#username-field"
         hx-swap="outerHTML">
  %s
</div>`, template.HTMLEscapeString(name), border, message))
}

var tabPanels = map[string]struct{ Label, Body string }{
	"overview": {"Overview", "A short summary of the thing this panel describes."},
	"pricing":  {"Pricing", "One price per seat, billed monthly or yearly."},
	"limits":   {"Limits", "Fair limits that scale with the size of the team."},
}

var tabOrder = []string{"overview", "pricing", "limits"}

// handleTabs returns the whole widget, so the active tab is decided here rather
// than tracked in the client.
func handleTabs(w http.ResponseWriter, r *http.Request) {
	active := r.PathValue("tab")
	panel, ok := tabPanels[active]
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Only the panel body. The strip is client state now - a radio group that
	// htmx never replaces - which is what lets the indicator slide between tabs
	// and what stops the "Panel surface" option being wiped on every change.
	writeHTML(w, template.HTMLEscapeString(panel.Body))
}

func handleDialogBody(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, `<h2 class="mb-2 text-title-md font-500 text-foreground">Delete this component?</h2>
<p class="text-body-md text-foreground">
  The directory and its three source files are removed. This cannot be undone.
</p>`)
}

// handleToast honours the tone the preview asked for. The tone is one class on
// the root; tokens.css maps it to surface, text, border and which icon shows.
func handleToast(w http.ResponseWriter, r *http.Request) {
	icons := map[string]string{
		"success": `<circle cx="12" cy="12" r="9"/><path d="m8 12 2.5 2.5L16 9"/>`,
		"error":   `<circle cx="12" cy="12" r="9"/><path d="m9 9 6 6M15 9l-6 6"/>`,
		"info":    `<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><path d="M12 7.5v.5"/>`,
	}
	messages := map[string]string{
		"success": "Changes saved",
		"error":   "That did not save. Try again.",
		"info":    "Three components were updated.",
	}

	tone := r.FormValue("tone")
	if _, ok := icons[tone]; !ok {
		tone = "success"
	}

	writeHTML(w, fmt.Sprintf(`<div class="toast-life tone-%s flex items-start gap-2 rounded-lg border px-control-x py-control-y shadow-e2"
     onanimationend="this.remove()">
  <svg class="mt-px shrink-0" width="18" height="18" viewBox="0 0 24 24" fill="none"
       stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">%s</svg>
  <p class="min-w-0 flex-1 text-body-sm leading-5">%s</p>
  <button type="button" aria-label="Dismiss"
          onclick="this.closest('.toast-life').remove()"
          class="mt-px grid shrink-0 place-items-center rounded-sm opacity-70 hover:opacity-100">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
         stroke-linecap="round" aria-hidden="true"><path d="M18 6 6 18M6 6l12 12"/></svg>
  </button>
</div>`, tone, icons[tone], template.HTMLEscapeString(messages[tone])))
}

// handleProgress walks a value forward, so the polling example in the Progress
// component has something honest to poll.
func handleProgress(w http.ResponseWriter, r *http.Request) {
	v := (time.Now().Unix() % 10) * 10
	writeHTML(w, fmt.Sprintf(`<progress class="progress" value="%d" max="100"
          hx-get="/demo/progress" hx-trigger="every 900ms" hx-swap="outerHTML">%d%%</progress>`, v, v))
}
