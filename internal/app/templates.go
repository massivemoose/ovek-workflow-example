package app

import (
	"html/template"
	"net/http"

	"github.com/massivemoose/ovek-workflow-example/internal/workflow"
)

type HomeData struct {
	WorkflowResults []workflow.WorkflowResult
	ResultsError    string
}

type resultData struct {
	Kind    string
	Title   string
	Message string
}

func renderHome(w http.ResponseWriter, data HomeData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := homeTemplate.Execute(w, data); err != nil {
		http.Error(w, "render page", http.StatusInternalServerError)
	}
}

func renderResult(w http.ResponseWriter, data resultData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := resultTemplate.Execute(w, data); err != nil {
		http.Error(w, "render page", http.StatusInternalServerError)
	}
}

var templateFuncs = template.FuncMap{
	"shortRunID": workflow.ShortRunID,
	"finishedAt": func(result workflow.WorkflowResult) string {
		if result.FinishedAt.IsZero() {
			return ""
		}
		return result.FinishedAt.UTC().Format("2006-01-02 15:04 UTC")
	},
}

var homeTemplate = template.Must(template.New("home").Funcs(templateFuncs).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Ovek Workflow Example</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: oklch(97.2% 0.012 86);
      --ink: oklch(23% 0.018 70);
      --muted: oklch(48% 0.018 72);
      --line: oklch(84% 0.018 76);
      --field: oklch(99% 0.006 82);
      --accent: oklch(41% 0.09 150);
      --accent-ink: oklch(98% 0.01 120);
      --accent-soft: oklch(92% 0.055 145);
      --danger-soft: oklch(93% 0.045 28);
      --shadow: 0 18px 45px oklch(48% 0.03 72 / 13%);
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 32px 18px;
      background:
        linear-gradient(135deg, oklch(99% 0.01 95), var(--bg) 46%, oklch(94% 0.018 118));
      color: var(--ink);
    }
    main {
      width: min(100%, 760px);
      display: grid;
      grid-template-columns: minmax(0, 1fr);
      gap: 18px;
    }
    .panel {
      padding: 28px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: oklch(98.5% 0.008 88 / 86%);
      box-shadow: var(--shadow);
    }
    .eyebrow {
      display: inline-flex;
      align-items: center;
      margin: 0 0 18px;
      padding: 5px 9px;
      border: 1px solid oklch(76% 0.04 145);
      border-radius: 999px;
      background: var(--accent-soft);
      color: oklch(30% 0.075 150);
      font-size: 0.78rem;
      font-weight: 650;
    }
    h1, h2 {
      margin: 0 0 10px;
      line-height: 1.08;
      letter-spacing: 0;
    }
    h1 { font-size: 2.35rem; }
    h2 { font-size: 1.15rem; }
    p {
      margin: 0 0 24px;
      max-width: 64ch;
      color: var(--muted);
      font-size: 1rem;
      line-height: 1.55;
    }
    form {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 10px;
      align-items: stretch;
      margin-top: 6px;
    }
    label {
      position: absolute;
      width: 1px;
      height: 1px;
      overflow: hidden;
      clip: rect(0 0 0 0);
      white-space: nowrap;
    }
    input, button {
      min-height: 48px;
      font: inherit;
      border-radius: 8px;
      border: 1px solid var(--line);
      padding: 12px 14px;
    }
    input {
      width: 100%;
      background: var(--field);
      color: var(--ink);
      transition: border-color 160ms ease-out, background-color 160ms ease-out;
    }
    input::placeholder { color: oklch(62% 0.018 72); }
    input:hover { border-color: oklch(75% 0.024 80); }
    input:focus {
      outline: 3px solid oklch(74% 0.06 150 / 35%);
      border-color: var(--accent);
      background: oklch(99.5% 0.004 88);
    }
    button {
      border-color: var(--accent);
      background: var(--accent);
      color: var(--accent-ink);
      cursor: pointer;
      font-weight: 700;
      transition: background-color 160ms ease-out, transform 160ms ease-out;
    }
    button:hover { background: oklch(36% 0.085 150); }
    button:active { transform: translateY(1px); }
    button:focus-visible {
      outline: 3px solid oklch(74% 0.06 150 / 42%);
      outline-offset: 2px;
    }
    .results-head {
      display: flex;
      align-items: baseline;
      justify-content: space-between;
      gap: 12px;
      margin-bottom: 14px;
    }
    .results-head p {
      margin: 0;
      font-size: 0.92rem;
    }
    .results-list {
      display: grid;
      gap: 10px;
      margin: 0;
      padding: 0;
      list-style: none;
    }
    .result {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 8px 14px;
      padding: 14px 0;
      border-top: 1px solid var(--line);
    }
    .result:first-child { border-top: 0; padding-top: 0; }
    .result-title {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 7px;
      margin: 0 0 5px;
      font-weight: 750;
    }
    .status {
      padding: 2px 7px;
      border-radius: 999px;
      background: var(--accent-soft);
      color: oklch(30% 0.075 150);
      font-size: 0.75rem;
      line-height: 1.35;
    }
    .status.failed {
      background: var(--danger-soft);
      color: oklch(36% 0.09 28);
    }
    .summary {
      margin: 0;
      font-size: 0.92rem;
    }
    .counts {
      display: inline-grid;
      grid-template-columns: repeat(2, auto);
      gap: 6px;
      align-content: start;
      justify-content: end;
      color: var(--muted);
      font-size: 0.86rem;
      white-space: nowrap;
    }
    .empty, .notice {
      margin: 0;
      padding-top: 2px;
      font-size: 0.95rem;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --bg: oklch(19% 0.015 74);
        --ink: oklch(93% 0.012 84);
        --muted: oklch(73% 0.015 78);
        --line: oklch(34% 0.018 76);
        --field: oklch(24% 0.014 74);
        --accent: oklch(74% 0.09 150);
        --accent-ink: oklch(18% 0.02 150);
        --accent-soft: oklch(30% 0.045 150);
        --danger-soft: oklch(31% 0.055 28);
        --shadow: 0 20px 50px oklch(6% 0.015 72 / 45%);
      }
      body {
        background:
          linear-gradient(135deg, oklch(16% 0.015 76), var(--bg) 48%, oklch(22% 0.018 120));
      }
      .panel { background: oklch(21% 0.014 78 / 88%); }
      .eyebrow {
        border-color: oklch(43% 0.06 150);
        color: oklch(86% 0.075 150);
      }
      input::placeholder { color: oklch(58% 0.012 76); }
      input:hover { border-color: oklch(43% 0.02 76); }
      input:focus { background: oklch(26% 0.014 74); }
      button:hover { background: oklch(80% 0.085 150); }
      .status { color: oklch(86% 0.075 150); }
      .status.failed { color: oklch(86% 0.07 30); }
    }
    @media (max-width: 640px) {
      body { align-items: start; }
      .panel { padding: 24px; }
      h1 { font-size: 2rem; }
      form, .result {
        grid-template-columns: 1fr;
      }
      .counts {
        grid-template-columns: repeat(2, auto);
        justify-content: start;
      }
    }
  </style>
</head>
<body>
  <main>
    <section class="panel" aria-labelledby="signup-title">
      <p class="eyebrow">PocketBase Email List Demo</p>
      <h1 id="signup-title">Join the Ovek list</h1>
      <p>Submit an email address and the app will write it to the managed PocketBase sidecar.</p>
      <form method="post" action="/signup">
        <label for="email">Email address</label>
        <input id="email" name="email" type="email" placeholder="you@example.com" autocomplete="email" required>
        <button type="submit">Sign up</button>
      </form>
    </section>

    <section class="panel" aria-labelledby="results-title">
      <div class="results-head">
        <h2 id="results-title">Recent workflow results</h2>
        <p>Latest 5 digest runs</p>
      </div>
      {{if .ResultsError}}
        <p class="notice">{{.ResultsError}}</p>
      {{else if .WorkflowResults}}
        <ol class="results-list">
          {{range .WorkflowResults}}
            <li class="result">
              <div>
                <p class="result-title">
                  <span>{{.Workflow}}</span>
                  <span class="status {{.Status}}">{{.Status}}</span>
                  <span>{{shortRunID .RunID}}</span>
                </p>
                <p class="summary">{{.Summary}}</p>
              </div>
              <div class="counts" aria-label="Signup counts">
                <span>{{.TotalSignups}} total</span>
                <span>{{.NewSignups}} new</span>
                <span>{{finishedAt .}}</span>
              </div>
            </li>
          {{end}}
        </ol>
      {{else}}
        <p class="empty">No workflow results yet.</p>
      {{end}}
    </section>
  </main>
</body>
</html>`))

var resultTemplate = template.Must(template.New("result").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}} | Ovek Workflow Example</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: oklch(97.2% 0.012 86);
      --ink: oklch(23% 0.018 70);
      --muted: oklch(48% 0.018 72);
      --line: oklch(84% 0.018 76);
      --accent: oklch(41% 0.09 150);
      --accent-ink: oklch(98% 0.01 120);
      --success-bg: oklch(92% 0.055 145);
      --success-ink: oklch(30% 0.075 150);
      --error-bg: oklch(91% 0.052 24);
      --error-ink: oklch(34% 0.09 28);
      --shadow: 0 18px 45px oklch(48% 0.03 72 / 13%);
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 32px 18px;
      background:
        linear-gradient(135deg, oklch(99% 0.01 95), var(--bg) 46%, oklch(94% 0.018 118));
      color: var(--ink);
    }
    main {
      width: min(100%, 520px);
      padding: 28px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: oklch(98.5% 0.008 88 / 86%);
      box-shadow: var(--shadow);
    }
    h1 {
      margin: 0 0 10px;
      font-size: 2.35rem;
      line-height: 1.02;
      letter-spacing: 0;
    }
    p {
      margin: 0 0 24px;
      max-width: 58ch;
      color: var(--muted);
      font-size: 1rem;
      line-height: 1.55;
    }
    .message {
      display: inline-block;
      max-width: 100%;
      margin: 0 0 18px;
      padding: 6px 10px;
      border-radius: 999px;
      font-size: 0.85rem;
      font-weight: 700;
      line-height: 1.35;
    }
    .success { background: var(--success-bg); color: var(--success-ink); }
    .error { background: var(--error-bg); color: var(--error-ink); }
    .button {
      display: inline-block;
      min-height: 48px;
      border: 1px solid var(--accent);
      border-radius: 8px;
      padding: 12px 14px;
      background: var(--accent);
      color: var(--accent-ink);
      font-weight: 700;
      text-decoration: none;
      transition: background-color 160ms ease-out, transform 160ms ease-out;
    }
    .button:hover { background: oklch(36% 0.085 150); }
    .button:active { transform: translateY(1px); }
    .button:focus-visible {
      outline: 3px solid oklch(74% 0.06 150 / 42%);
      outline-offset: 2px;
    }
    @media (prefers-color-scheme: dark) {
      :root {
        --bg: oklch(19% 0.015 74);
        --ink: oklch(93% 0.012 84);
        --muted: oklch(73% 0.015 78);
        --line: oklch(34% 0.018 76);
        --accent: oklch(74% 0.09 150);
        --accent-ink: oklch(18% 0.02 150);
        --success-bg: oklch(30% 0.045 150);
        --success-ink: oklch(86% 0.075 150);
        --error-bg: oklch(31% 0.055 28);
        --error-ink: oklch(86% 0.07 30);
        --shadow: 0 20px 50px oklch(6% 0.015 72 / 45%);
      }
      body {
        background:
          linear-gradient(135deg, oklch(16% 0.015 76), var(--bg) 48%, oklch(22% 0.018 120));
      }
      main { background: oklch(21% 0.014 78 / 88%); }
      .button:hover { background: oklch(80% 0.085 150); }
    }
    @media (max-width: 520px) {
      main { padding: 24px; }
      h1 { font-size: 2rem; }
    }
  </style>
</head>
<body>
  <main>
    <div class="message {{.Kind}}">{{.Message}}</div>
    <h1>{{.Title}}</h1>
    <p>Return to the form whenever you want to submit another email address.</p>
    {{if eq .Kind "success"}}
      <a class="button" href="/">Add another email</a>
    {{else}}
      <a class="button" href="/">Try another email</a>
    {{end}}
  </main>
</body>
</html>`))
