package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	hostinfo "github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "hf_session"
	defaultAdminUser  = "admin"
	defaultAdminPass  = "admin123"
)

var (
	loginTemplate    = template.Must(template.New("login").Parse(loginHTML))
	monitorTemplate  = template.Must(template.New("monitor").Parse(monitorHTML))
	accountsTemplate = template.Must(template.New("accounts").Parse(accountsHTML))
)

type authUser struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Role           string `json:"role"`
	CanStartRun    bool   `json:"can_start_run"`
	CanViewMonitor bool   `json:"can_view_monitor"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type authSession struct {
	Token     string
	UserID    int64
	ExpiresAt string
}

type systemMetrics struct {
	Timestamp    string  `json:"timestamp"`
	HostName     string  `json:"host_name"`
	OS           string  `json:"os"`
	Platform     string  `json:"platform"`
	UptimeSec    uint64  `json:"uptime_sec"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemTotal     uint64  `json:"mem_total"`
	MemUsed      uint64  `json:"mem_used"`
	MemAvailable uint64  `json:"mem_available"`
	MemPercent   float64 `json:"mem_percent"`
	DiskTotal    uint64  `json:"disk_total"`
	DiskUsed     uint64  `json:"disk_used"`
	DiskFree     uint64  `json:"disk_free"`
	DiskPercent  float64 `json:"disk_percent"`
	GoRoutines   int     `json:"go_routines"`
}

func (s *runStore) ensureSingleAdmin() error {
	var adminCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&adminCount); err != nil {
		return err
	}
	if adminCount > 1 {
		return fmt.Errorf("invalid state: more than one admin account exists")
	}
	if adminCount == 1 {
		return nil
	}

	username := strings.TrimSpace(os.Getenv("HTTPFLOOD_ADMIN_USER"))
	if username == "" {
		username = defaultAdminUser
	}
	password := strings.TrimSpace(os.Getenv("HTTPFLOOD_ADMIN_PASS"))
	if password == "" {
		password = defaultAdminPass
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().Format(time.RFC3339)
	_, err = s.db.Exec(
		`INSERT INTO users (username, password_hash, role, can_start_run, can_view_monitor, is_active, created_at, updated_at)
		 VALUES (?, ?, 'admin', 1, 1, 1, ?, ?)`,
		username, string(hash), now, now,
	)
	if err != nil {
		return err
	}
	fmt.Println("Admin account created.")
	fmt.Println("Admin username:", username)
	fmt.Println("Admin password:", password)
	fmt.Println("Set HTTPFLOOD_ADMIN_USER and HTTPFLOOD_ADMIN_PASS to change bootstrap credentials.")
	return nil
}

func (s *runStore) getUserByUsername(username string) (authUser, string, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, role, can_start_run, can_view_monitor, is_active, created_at, updated_at
		 FROM users
		 WHERE username = ?`,
		username,
	)
	var user authUser
	var hash string
	var canStartInt, canViewInt, activeInt int
	err := row.Scan(
		&user.ID, &user.Username, &hash, &user.Role,
		&canStartInt, &canViewInt, &activeInt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return authUser{}, "", err
	}
	user.CanStartRun = canStartInt == 1
	user.CanViewMonitor = canViewInt == 1
	user.IsActive = activeInt == 1
	return user, hash, nil
}

func (s *runStore) getUserByID(id int64) (authUser, error) {
	row := s.db.QueryRow(
		`SELECT id, username, role, can_start_run, can_view_monitor, is_active, created_at, updated_at
		 FROM users
		 WHERE id = ?`,
		id,
	)
	var user authUser
	var canStartInt, canViewInt, activeInt int
	err := row.Scan(
		&user.ID, &user.Username, &user.Role, &canStartInt, &canViewInt, &activeInt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return authUser{}, err
	}
	user.CanStartRun = canStartInt == 1
	user.CanViewMonitor = canViewInt == 1
	user.IsActive = activeInt == 1
	return user, nil
}

func (s *runStore) listUsers() ([]authUser, error) {
	rows, err := s.db.Query(
		`SELECT id, username, role, can_start_run, can_view_monitor, is_active, created_at, updated_at
		 FROM users
		 ORDER BY id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]authUser, 0)
	for rows.Next() {
		var user authUser
		var canStartInt, canViewInt, activeInt int
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Role, &canStartInt, &canViewInt, &activeInt, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.CanStartRun = canStartInt == 1
		user.CanViewMonitor = canViewInt == 1
		user.IsActive = activeInt == 1
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *runStore) createUser(username, password, role string, canStart, canView bool, createdBy int64) (authUser, error) {
	if role != "member" {
		return authUser{}, fmt.Errorf("only member role can be created")
	}
	if strings.TrimSpace(username) == "" {
		return authUser{}, fmt.Errorf("username is required")
	}
	if len(password) < 6 {
		return authUser{}, fmt.Errorf("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return authUser{}, err
	}
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO users (
			username, password_hash, role, can_start_run, can_view_monitor, is_active, created_at, updated_at, created_by
		) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`,
		strings.TrimSpace(username),
		string(hash),
		role,
		boolInt(canStart),
		boolInt(canView),
		now,
		now,
		createdBy,
	)
	if err != nil {
		return authUser{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return authUser{}, err
	}
	return s.getUserByID(id)
}

func (s *runStore) updateUserPermissions(userID int64, canStart, canView, active bool) error {
	user, err := s.getUserByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		return fmt.Errorf("admin account cannot be modified here")
	}
	_, err = s.db.Exec(
		`UPDATE users
		 SET can_start_run = ?, can_view_monitor = ?, is_active = ?, updated_at = ?
		 WHERE id = ?`,
		boolInt(canStart),
		boolInt(canView),
		boolInt(active),
		time.Now().Format(time.RFC3339),
		userID,
	)
	return err
}

func (s *runStore) createSession(userID int64) (authSession, error) {
	token, err := randomToken(32)
	if err != nil {
		return authSession{}, err
	}
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	_, err = s.db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at, created_at, last_seen_at)
		 VALUES (?, ?, ?, ?, ?)`,
		token, userID, expires.Format(time.RFC3339), now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return authSession{}, err
	}
	return authSession{Token: token, UserID: userID, ExpiresAt: expires.Format(time.RFC3339)}, nil
}

func (s *runStore) getSessionUser(token string) (authUser, error) {
	row := s.db.QueryRow(
		`SELECT u.id, u.username, u.role, u.can_start_run, u.can_view_monitor, u.is_active, u.created_at, u.updated_at
		 FROM sessions sess
		 JOIN users u ON u.id = sess.user_id
		 WHERE sess.id = ? AND sess.expires_at > ?`,
		token, time.Now().Format(time.RFC3339),
	)
	var user authUser
	var canStartInt, canViewInt, activeInt int
	if err := row.Scan(
		&user.ID, &user.Username, &user.Role, &canStartInt, &canViewInt, &activeInt, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return authUser{}, err
	}
	user.CanStartRun = canStartInt == 1
	user.CanViewMonitor = canViewInt == 1
	user.IsActive = activeInt == 1
	if !user.IsActive {
		return authUser{}, fmt.Errorf("account is inactive")
	}
	_, _ = s.db.Exec(`UPDATE sessions SET last_seen_at = ? WHERE id = ?`, time.Now().Format(time.RFC3339), token)
	return user, nil
}

func (s *runStore) deleteSession(token string) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, token)
}

func (s *runStore) purgeExpiredSessions() {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, time.Now().Format(time.RFC3339))
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func currentUser(r *http.Request) (authUser, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return authUser{}, err
	}
	return webStore.getSessionUser(cookie.Value)
}

func requireAuthPage(w http.ResponseWriter, r *http.Request) (authUser, bool) {
	user, err := currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return authUser{}, false
	}
	return user, true
}

func requireAuthAPI(w http.ResponseWriter, r *http.Request) (authUser, bool) {
	user, err := currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return authUser{}, false
	}
	return user, true
}

func isAdmin(user authUser) bool {
	return user.Role == "admin"
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, err := currentUser(r); err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := loginTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleLoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := parseWebForm(r); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	if username == "" || password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	user, hash, err := webStore.getUserByUsername(username)
	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !user.IsActive {
		writeJSONError(w, http.StatusForbidden, "account is inactive")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	session, err := webStore.createSession(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func handleLogoutAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		webStore.deleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func handleMeAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func handleAccountsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthPage(w, r)
	if !ok {
		return
	}
	if !isAdmin(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := accountsTemplate.Execute(w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAccountsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	if !isAdmin(user) {
		writeJSONError(w, http.StatusForbidden, "admin only")
		return
	}
	switch r.Method {
	case http.MethodGet:
		users, err := webStore.listUsers()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"accounts": users})
	case http.MethodPost:
		if err := parseWebForm(r); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		account, err := webStore.createUser(
			r.FormValue("username"),
			r.FormValue("password"),
			"member",
			r.FormValue("can_start_run") == "true",
			r.FormValue("can_view_monitor") == "true",
			user.ID,
		)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"account": account})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleAccountResourceAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	if !isAdmin(user) {
		writeJSONError(w, http.StatusForbidden, "admin only")
		return
	}
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PATCH")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/accounts/"), "/")
	accountID, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || accountID <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	var payload struct {
		CanStartRun    bool `json:"can_start_run"`
		CanViewMonitor bool `json:"can_view_monitor"`
		IsActive       bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := webStore.updateUserPermissions(accountID, payload.CanStartRun, payload.CanViewMonitor, payload.IsActive); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func handleMonitorPage(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthPage(w, r)
	if !ok {
		return
	}
	if !user.CanViewMonitor && !isAdmin(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := monitorTemplate.Execute(w, user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleMonitorMetricsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthAPI(w, r)
	if !ok {
		return
	}
	if !user.CanViewMonitor && !isAdmin(user) {
		writeJSONError(w, http.StatusForbidden, "monitor access denied")
		return
	}
	metrics, err := collectSystemMetrics()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": metrics})
}

func collectSystemMetrics() (systemMetrics, error) {
	info, err := hostinfo.Info()
	if err != nil {
		return systemMetrics{}, err
	}
	cpuValues, err := cpu.Percent(0, false)
	if err != nil {
		return systemMetrics{}, err
	}
	memValues, err := mem.VirtualMemory()
	if err != nil {
		return systemMetrics{}, err
	}
	diskValues, err := disk.Usage("/")
	if err != nil {
		diskValues = &disk.UsageStat{}
	}

	cpuPercent := 0.0
	if len(cpuValues) > 0 {
		cpuPercent = cpuValues[0]
	}

	return systemMetrics{
		Timestamp:    time.Now().Format(time.RFC3339),
		HostName:     info.Hostname,
		OS:           info.OS,
		Platform:     info.Platform + " " + info.PlatformVersion,
		UptimeSec:    info.Uptime,
		CPUPercent:   cpuPercent,
		MemTotal:     memValues.Total,
		MemUsed:      memValues.Used,
		MemAvailable: memValues.Available,
		MemPercent:   memValues.UsedPercent,
		DiskTotal:    diskValues.Total,
		DiskUsed:     diskValues.Used,
		DiskFree:     diskValues.Free,
		DiskPercent:  diskValues.UsedPercent,
		GoRoutines:   runtime.NumGoroutine(),
	}, nil
}

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>httpflood login</title>
  <style>
    :root { color-scheme: dark; --line:#1b6131; --text:#8eff9a; --muted:#53bb68; --accent:#0ea12c; --danger:#b14a4a; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; display:grid; place-items:center; background:#000; color:var(--text); font-family:Consolas, "Courier New", monospace; padding:12px; }
    .card { width:min(420px,100%); background:rgba(2,10,5,.95); border:1px solid var(--line); border-radius:4px; padding:16px; display:grid; gap:10px; }
    h1 { margin:0; font-size:1.2rem; text-transform:uppercase; }
    p { margin:0; color:var(--muted); font-size:.86rem; }
    form { display:grid; gap:8px; }
    input { width:100%; border:1px solid var(--line); border-radius:2px; background:#020b04; color:var(--text); padding:9px; font:inherit; }
    button { border:0; border-radius:2px; background:var(--accent); color:#041307; padding:10px; font:inherit; font-weight:700; text-transform:uppercase; cursor:pointer; }
    .msg { border:1px solid var(--danger); color:#ff9b9b; background:#120707; padding:8px; display:none; white-space:pre-wrap; }
  </style>
</head>
<body>
  <main class="card">
    <h1>httpflood login</h1>
    <p>session required</p>
    <div id="msg" class="msg"></div>
    <form id="login-form">
      <input name="username" type="text" placeholder="username" autocomplete="username" required>
      <input name="password" type="password" placeholder="password" autocomplete="current-password" required>
      <button type="submit">Login</button>
    </form>
  </main>
  <script>
    const form = document.getElementById('login-form');
    const msg = document.getElementById('msg');
    form.addEventListener('submit', async function (event) {
      event.preventDefault();
      msg.style.display = 'none';
      const response = await fetch('/api/login', { method: 'POST', body: new URLSearchParams(new FormData(form)) });
      const payload = await response.json().catch(function () { return {}; });
      if (!response.ok) {
        msg.textContent = payload.error || response.statusText;
        msg.style.display = 'block';
        return;
      }
      window.location.href = '/';
    });
  </script>
</body>
</html>`

const monitorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>httpflood monitor</title>
  <style>
    :root { color-scheme: dark; --line:#1b6131; --text:#8eff9a; --muted:#53bb68; --accent:#0ea12c; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; background:#000; color:var(--text); font-family:Consolas, "Courier New", monospace; padding:12px; }
    .wrap { max-width:980px; margin:0 auto; display:grid; gap:12px; }
    .card { background:rgba(2,10,5,.95); border:1px solid var(--line); border-radius:4px; padding:14px; }
    .top { display:flex; justify-content:space-between; gap:10px; align-items:center; flex-wrap:wrap; }
    a, button { border:1px solid var(--line); background:#021107; color:var(--text); text-decoration:none; padding:8px 10px; border-radius:2px; font:inherit; cursor:pointer; }
    .grid { display:grid; gap:8px; grid-template-columns:repeat(2,minmax(0,1fr)); }
    .tile { border:1px solid var(--line); border-radius:3px; padding:10px; background:#021007; }
    .label { color:var(--muted); font-size:.82rem; }
    .value { font-size:1.1rem; margin-top:4px; }
    @media (max-width:720px){ .grid{ grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <main class="wrap">
    <section class="card top">
      <strong>Monitor realtime</strong>
      <div>
        <a href="/">Flood</a>
        {{if eq .Role "admin"}}<a href="/accounts">Accounts</a>{{end}}
        <button id="logout" type="button">Logout</button>
      </div>
    </section>
    <section class="card">
      <div class="label" id="stamp">Loading...</div>
      <div class="grid">
        <div class="tile"><div class="label">CPU</div><div class="value" id="cpu">-</div></div>
        <div class="tile"><div class="label">RAM</div><div class="value" id="ram">-</div></div>
        <div class="tile"><div class="label">Disk</div><div class="value" id="disk">-</div></div>
        <div class="tile"><div class="label">Uptime</div><div class="value" id="uptime">-</div></div>
        <div class="tile"><div class="label">Host</div><div class="value" id="host">-</div></div>
        <div class="tile"><div class="label">Go Routines</div><div class="value" id="goroutines">-</div></div>
      </div>
    </section>
  </main>
  <script>
    function bytes(v) {
      const units = ['B','KB','MB','GB','TB'];
      let i = 0;
      let n = Number(v || 0);
      while (n >= 1024 && i < units.length - 1) { n /= 1024; i++; }
      return n.toFixed(1) + ' ' + units[i];
    }
    function uptime(sec) {
      const s = Number(sec || 0);
      const d = Math.floor(s / 86400);
      const h = Math.floor((s % 86400) / 3600);
      const m = Math.floor((s % 3600) / 60);
      const r = [];
      if (d) r.push(d + 'd');
      if (h || d) r.push(h + 'h');
      r.push(m + 'm');
      return r.join(' ');
    }
    async function refresh() {
      const response = await fetch('/api/system/metrics');
      const payload = await response.json().catch(function () { return {}; });
      if (!response.ok) return;
      const m = payload.metrics;
      document.getElementById('stamp').textContent = 'Updated ' + new Date(m.timestamp).toLocaleTimeString();
      document.getElementById('cpu').textContent = m.cpu_percent.toFixed(1) + '%';
      document.getElementById('ram').textContent = bytes(m.mem_used) + ' / ' + bytes(m.mem_total) + ' (' + m.mem_percent.toFixed(1) + '%)';
      document.getElementById('disk').textContent = bytes(m.disk_used) + ' / ' + bytes(m.disk_total) + ' (' + m.disk_percent.toFixed(1) + '%)';
      document.getElementById('uptime').textContent = uptime(m.uptime_sec);
      document.getElementById('host').textContent = m.host_name + ' - ' + m.platform;
      document.getElementById('goroutines').textContent = String(m.go_routines);
    }
    document.getElementById('logout').addEventListener('click', async function () {
      await fetch('/api/logout', { method: 'POST' });
      window.location.href = '/login';
    });
    refresh();
    window.setInterval(refresh, 1000);
  </script>
</body>
</html>`

const accountsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>httpflood accounts</title>
  <style>
    :root { color-scheme: dark; --line:#1b6131; --text:#8eff9a; --muted:#53bb68; --accent:#0ea12c; --danger:#b14a4a; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; background:#000; color:var(--text); font-family:Consolas, "Courier New", monospace; padding:12px; }
    .wrap { max-width:1080px; margin:0 auto; display:grid; gap:12px; }
    .card { background:rgba(2,10,5,.95); border:1px solid var(--line); border-radius:4px; padding:14px; }
    .top { display:flex; justify-content:space-between; gap:10px; align-items:center; flex-wrap:wrap; }
    a, button { border:1px solid var(--line); background:#021107; color:var(--text); text-decoration:none; padding:8px 10px; border-radius:2px; font:inherit; cursor:pointer; }
    form { display:grid; gap:8px; grid-template-columns:repeat(2,minmax(0,1fr)); }
    form input, form select { width:100%; border:1px solid var(--line); border-radius:2px; background:#020b04; color:var(--text); padding:8px; font:inherit; }
    .full { grid-column:1/-1; }
    table { width:100%; border-collapse:collapse; font-size:.9rem; }
    th, td { border:1px solid var(--line); padding:8px; text-align:left; vertical-align:middle; }
    th { color:var(--muted); }
    .msg { border:1px solid var(--line); background:#021107; padding:8px; white-space:pre-wrap; }
    .err { border-color:var(--danger); color:#ff9b9b; }
    @media (max-width:720px){ form{ grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <main class="wrap">
    <section class="card top">
      <strong>Account management</strong>
      <div>
        <a href="/">Flood</a>
        <a href="/monitor">Monitor</a>
        <button id="logout" type="button">Logout</button>
      </div>
    </section>
    <section class="card">
      <div id="msg" class="msg" style="display:none"></div>
      <form id="create-form">
        <input name="username" type="text" placeholder="username" required>
        <input name="password" type="password" placeholder="password (min 6)" required>
        <label><input name="can_start_run" type="checkbox" checked> can start run</label>
        <label><input name="can_view_monitor" type="checkbox" checked> can view monitor</label>
        <button class="full" type="submit">Create account</button>
      </form>
    </section>
    <section class="card">
      <table>
        <thead>
          <tr><th>ID</th><th>Username</th><th>Role</th><th>Start</th><th>Monitor</th><th>Active</th><th>Action</th></tr>
        </thead>
        <tbody id="accounts-body"></tbody>
      </table>
    </section>
  </main>
  <script>
    const msg = document.getElementById('msg');
    const body = document.getElementById('accounts-body');
    const form = document.getElementById('create-form');
    let accounts = [];

    function showMessage(text, isError) {
      msg.textContent = text;
      msg.style.display = text ? 'block' : 'none';
      msg.className = isError ? 'msg err' : 'msg';
    }

    async function api(path, options) {
      const response = await fetch(path, options);
      const payload = await response.json().catch(function () { return {}; });
      if (!response.ok) throw new Error(payload.error || response.statusText);
      return payload;
    }

    function renderRows() {
      body.replaceChildren();
      for (const account of accounts) {
        const tr = document.createElement('tr');
        const disabled = account.role === 'admin';
        tr.innerHTML = '<td>' + account.id + '</td>' +
          '<td>' + account.username + '</td>' +
          '<td>' + account.role + '</td>' +
          '<td><input type="checkbox" data-key="can_start_run"' + (account.can_start_run ? ' checked' : '') + (disabled ? ' disabled' : '') + '></td>' +
          '<td><input type="checkbox" data-key="can_view_monitor"' + (account.can_view_monitor ? ' checked' : '') + (disabled ? ' disabled' : '') + '></td>' +
          '<td><input type="checkbox" data-key="is_active"' + (account.is_active ? ' checked' : '') + (disabled ? ' disabled' : '') + '></td>' +
          '<td>' + (disabled ? '' : '<button type="button">Save</button>') + '</td>';
        if (!disabled) {
          tr.querySelector('button').addEventListener('click', async function () {
            try {
              await api('/api/accounts/' + account.id, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                  can_start_run: tr.querySelector('input[data-key="can_start_run"]').checked,
                  can_view_monitor: tr.querySelector('input[data-key="can_view_monitor"]').checked,
                  is_active: tr.querySelector('input[data-key="is_active"]').checked
                })
              });
              showMessage('Updated account #' + account.id, false);
              await loadAccounts();
            } catch (error) {
              showMessage(error.message, true);
            }
          });
        }
        body.appendChild(tr);
      }
    }

    async function loadAccounts() {
      const payload = await api('/api/accounts');
      accounts = payload.accounts || [];
      renderRows();
    }

    form.addEventListener('submit', async function (event) {
      event.preventDefault();
      try {
        const fd = new FormData(form);
        fd.set('can_start_run', fd.get('can_start_run') ? 'true' : 'false');
        fd.set('can_view_monitor', fd.get('can_view_monitor') ? 'true' : 'false');
        await api('/api/accounts', { method: 'POST', body: new URLSearchParams(fd) });
        form.reset();
        showMessage('Account created', false);
        await loadAccounts();
      } catch (error) {
        showMessage(error.message, true);
      }
    });

    document.getElementById('logout').addEventListener('click', async function () {
      await fetch('/api/logout', { method: 'POST' });
      window.location.href = '/login';
    });

    loadAccounts();
  </script>
</body>
</html>`
