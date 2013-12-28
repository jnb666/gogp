// gogpweb is an http server for plotting gogp stats data in real time
package main

import (
    "log"
    "strconv"
    "net/http"
    "os"
    "os/exec"
    "flag"
    "github.com/jnb666/gogp/stats"
)

var webBrowser = []string{
    "/etc/alternatives/gnome-www-browser",
    "/etc/alternatives/x-www-browser",
    "google-chrome",
    "firefox",
}

// point a web browser to url - assumes Linux
func startBrowser(url string) {
    for _, name := range webBrowser {
        cmd := exec.Command(name, url)
        if err := cmd.Start(); err == nil { 
            log.Println("started browser", name, url)    
            return
        }
    }
    log.Println("no browser found - go to", url, "to view the data")
}

// wrapper to log request urls
func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Println(r.URL)
	    handler.ServeHTTP(w, r)
    })
}

// main server loop
func main() {
    var err error
    var browser, verbose bool
    var webPort int
    var webRoot string
    cdir, _ := os.Getwd()
    flag.StringVar(&webRoot, "root", cdir+"/docs", "root directory for web docs")
    flag.IntVar(&webPort, "port", 8080, "port number for web server")
    flag.BoolVar(&browser, "browser", false, "start web browser")
    flag.BoolVar(&verbose, "v", false, "log requests")
    flag.Parse()
    http.Handle("/data/", stats.NewHistory().Serve())
    http.Handle("/", http.FileServer(http.Dir(webRoot)))
    port := ":" + strconv.Itoa(webPort)
    if browser { 
        startBrowser("http://localhost" + port)
    }
    if verbose {
        err = http.ListenAndServe(port, Log(http.DefaultServeMux))
    } else {
        err = http.ListenAndServe(port, nil)
    }
    log.Fatal("ListenAndServe: ", err)
}

