package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
)

type IPInfo struct {
    Query       string `json:"query"`
    Country     string `json:"country"`
    RegionName  string `json:"regionName"`
    City        string `json:"city"`
    ISP         string `json:"isp"`
    Org         string `json:"org"`
    Timezone    string `json:"timezone"`
    AS          string `json:"as"`
    Lat         float64 `json:"lat"`
    Lon         float64 `json:"lon"`
}

func getIP(r *http.Request) string {
    // Check if behind proxy
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return forwarded
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}

func fetchIPInfo(ip string) (*IPInfo, error) {
    url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
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

func htmlHandler(w http.ResponseWriter, r *http.Request) {
    ip := getIP(r)
    info, err := fetchIPInfo(ip)
    if err != nil {
        http.Error(w, "Failed to fetch IP info", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprintf(w, `
        <html><head><title>My IP Info</title></head><body style="font-family:sans-serif;max-width:600px;margin:20px auto">
        <h1>📡 Your IP Information</h1>
        <p><strong>IP Address:</strong> %s</p>
        <p><strong>Location:</strong> %s, %s, %s</p>
        <p><strong>ISP:</strong> %s</p>
        <p><strong>Org:</strong> %s</p>
        <p><strong>Timezone:</strong> %s</p>
        <p><strong>User-Agent:</strong> %s</p>
        <hr><p><a href="/json">View as JSON</a></p>
        </body></html>
    `, info.Query, info.City, info.RegionName, info.Country, info.ISP, info.Org, info.Timezone, r.UserAgent())
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
