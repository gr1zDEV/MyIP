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
    // If behind a proxy or CDN
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return forwarded
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}

func fetchIPInfo(ip string) (*IPInfo, error) {
    url := fmt.Sprintf("https://ipwho.is/%s", ip)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    var info IPInfo
    if err := json.Unmarshal(body, &info); err != nil {
        return nil, err
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
        <html><head><title>My IP Info</title></head>
        <body style="font-family:sans-serif;max-width:600px;margin:20px auto">
        <h1>📡 Your IP Information</h1>
        <p><strong>IP Address:</strong> %s (%s)</p>
        <p><strong>Location:</strong> %s, %s, %s</p>
        <p><strong>Latitude / Longitude:</strong> %.4f, %.4f</p>
        <p><strong>ISP:</strong> %s</p>
        <p><strong>ASN:</strong> %s (#%d)</p>
        <p><strong>User-Agent:</strong> %s</p>
        <hr>
        <p><a href="/json">View as JSON</a></p>
        </body></html>
    `,
        info.IP, getIPVersion(info.IP),
        info.City, info.Region, info.Country,
        info.Latitude, info.Longitude,
        info.Connection.ISP,
        info.ASN.Name, info.ASN.ASN,
        r.UserAgent())
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
