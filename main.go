package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
)

type IPInfo struct {
    IP       string  `json:"ip"`
    Type     string  `json:"type"`
    Country  string  `json:"country"`
    Region   string  `json:"region"`
    City     string  `json:"city"`
    Latitude float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    ASN struct {
        ASN  int    `json:"asn"`
        Name string `json:"name"`
    } `json:"asn"`
    Connection struct {
        ISP string `json:"isp"`
    } `json:"connection"`
}

func getIP(r *http.Request) string {
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        return forwarded
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}

func fetchIPInfo(ip string) (*IPInfo, error) {
    url := fmt.Sprintf("https://ipwho.is/%s", ip)

    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (MyIPApp)")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to reach ipwho.is: %v", err)
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("ipwho.is status %d: %s", resp.StatusCode, string(body))
    }

    var info IPInfo
    if err := json.Unmarshal(body, &info); err != nil {
        return nil, fmt.Errorf("error parsing response: %v\nRaw: %s", err, string(body))
    }

    return &info, nil
}

func getIPVersion(ip string) string {
    parsed := net.ParseIP(ip)
    if parsed == nil {
        return "Unknown"
    }
    if parsed.To4() != nil {
        return "IPv4"
    }
    return "IPv6"
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    info, err := fetchIPInfo(ip)
    if err != nil {
        http.Error(w, "Failed to fetch IP info", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprintf(w, `
        <html>
        <head>
            <title>My IP Info</title>
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <style>
                body {
                    font-family: sans-serif;
                    max-width: 600px;
                    margin: 20px auto;
                    background: #121212;
                    color: #eee;
                    transition: all 0.3s ease;
                }
                .light-mode {
                    background: #fff;
                    color: #000;
                }
                #map { height: 300px; margin-top: 20px; }
                iframe { width: 100%%; border: none; }
                .theme-toggle {
                    margin: 10px 0;
                    padding: 6px 12px;
                    font-size: 14px;
                    cursor: pointer;
                }
            </style>
            <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
            <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
        </head>
        <body>
            <h1>Your IP Information</h1>
            <button class="theme-toggle" onclick="toggleTheme()">Toggle Theme</button>
            <p><strong>IP Address:</strong> %s (%s)</p>
            <p><strong>Location:</strong> %s, %s, %s</p>
            <p><strong>Latitude / Longitude:</strong> %.4f, %.4f</p>
            <p><strong>ISP:</strong> %s</p>`,
        info.IP, getIPVersion(info.IP),
        info.City, info.Region, info.Country,
        info.Latitude, info.Longitude,
        info.Connection.ISP)

    if info.ASN.ASN != 0 && info.ASN.Name != "" {
        fmt.Fprintf(w, `<p><strong>ASN:</strong> %s (#%d)</p>`, info.ASN.Name, info.ASN.ASN)
    }

    fmt.Fprintf(w, `
            <p><strong>User-Agent:</strong> %s</p>
            <div id="map"></div>

            <script>
                var map = L.map('map').setView([%f, %f], 10);
                L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
                    attribution: '&copy; OpenStreetMap contributors'
                }).addTo(map);
                L.marker([%f, %f]).addTo(map)
                    .bindPopup('%s, %s')
                    .openPopup();
            </script>

            <h2>Speed Test</h2>
            <div style="text-align:right;">
              <div style="min-height:360px;">
                <div style="width:100%%;height:0;padding-bottom:50%%;position:relative;">
                  <iframe style="position:absolute;top:0;left:0;width:100%%;height:100%%;min-height:360px;overflow:hidden !important;" src="https://www.metercustom.net/plugin/"></iframe>
                </div>
              </div>
              Provided by <a href="https://www.meter.net">Meter.net</a>
            </div>

            <hr><p><a href="/json">View as JSON</a></p>

            <script>
                function toggleTheme() {
                    document.body.classList.toggle('light-mode');
                    const isLight = document.body.classList.contains('light-mode');
                    localStorage.setItem('theme', isLight ? 'light' : 'dark');
                }
                window.onload = function() {
                    const saved = localStorage.getItem('theme');
                    if (saved === 'light') {
                        document.body.classList.add('light-mode');
                    }
                }
            </script>
        </body>
        </html>
    `,
        r.UserAgent(),
        info.Latitude, info.Longitude,
        info.Latitude, info.Longitude,
        info.City, info.Country)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    info, err := fetchIPInfo(ip)
    if err != nil {
        http.Error(w, "Failed to fetch IP info", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}

func main() {
    http.HandleFunc("/", htmlHandler)
    http.HandleFunc("/json", jsonHandler)

    fmt.Println("Starting server on port 8000...")
    http.ListenAndServe(":8000", nil)
}
