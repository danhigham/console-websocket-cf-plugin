package main

import (
  "bufio"
  "code.google.com/p/go.net/websocket"
  "io"
  "flag"
  "fmt"
  "log"
  "net/http"
  "net/http/httputil"
  "net/url"
  "os"
  "os/exec"
  "regexp"
  "strings"
  "time"
)

func printToLog(s string) {
  t := time.Now()
  ts := fmt.Sprintf("[%04d-%02d-%02d %02d:%02d:%02d]", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
  os.Stdout.Write([]byte(ts + " " + s))
}

func reader(ws *websocket.Conn, r io.Reader) {
  
  for {
    buf := make([]byte, 1024)
    n, err := r.Read(buf[:])
    
    if err != nil {
      return
    }

    ws.Write(buf[0:n])
  }
}

func logPipe(r io.Reader) {
  
  for {
    buf := make([]byte, 1024)
    n, err := r.Read(buf[:])
    
    if err != nil {
      return
    }

    os.Stdout.Write(buf[0:n])
  }
}

func runMainWebProc(mainProcCommand string) {
  bin := ""

  a := strings.Split(mainProcCommand, " ")
  bin, a = a[0], a[1:len(a)]
  cmd := exec.Command(bin, a...)

  stdout, err := cmd.StdoutPipe()
  if err != nil {
    log.Print(err)
  }

  stderr, err := cmd.StderrPipe()
  if err != nil {
    log.Print(err)
  }

  errbr := bufio.NewReader(stderr)
  outbr := bufio.NewReader(stdout)
   
  go logPipe(outbr)
  go logPipe(errbr)

  cmd.Run()
}

func main() {

  var consoleProcess = flag.String("console-process", "bash", "The process to be started and connected to by the console command line tool")
  var mainProcess = flag.String("main-process", "", "The main application to be run (e.g rails s)")
   
  wsHandler := websocket.Handler(func (ws *websocket.Conn) {
    
    bin := ""

    a := strings.Split(*consoleProcess, " ")
    bin, a = a[0], a[1:len(a)]
    cmd := exec.Command(bin, a...)

    stdout, err := cmd.StdoutPipe()
    if err != nil {
      log.Print(err)
    }

    stderr, err := cmd.StderrPipe()
    if err != nil {
      log.Print(err)
    }

    stdin, err := cmd.StdinPipe()
    if err != nil {
      log.Print(err)
    }

    if err := cmd.Start(); err != nil {
      log.Print(err)
    }

    if (*consoleProcess == "rails c") {
      io.WriteString(stdin, "conf.prompt_mode=:DEFAULT\n")
    }

    errbr := bufio.NewReader(stderr)
    outbr := bufio.NewReader(stdout)
   
    go reader(ws, outbr)
    go reader(ws, errbr)

    go func() {
      for {
        msg := make([]byte, 1024)
        n, err := ws.Read(msg)
          
        if err != nil {
          log.Print(err)
          ws.Close()
          cmd.Process.Kill()
          return
        }

        io.WriteString(stdin, string(msg[:n]) + "\n")
      }
    }()

    cmd.Wait()

  })

  flag.Parse()

  appVersionRegex, _ := regexp.Compile(`,\"application_version\":\"([^\"]+)`)
  instanceIndexRegex, _ := regexp.Compile(`,\"instance_index\"\:(\d{1,2})`)

  vcap_data := os.Getenv("VCAP_APPLICATION")
  appVersion := appVersionRegex.FindStringSubmatch(vcap_data)[1]
  instanceIndex := instanceIndexRegex.FindStringSubmatch(vcap_data)[1]

  if (*mainProcess != "") {
    printToLog("Running main process :- " + *mainProcess + "\n")
    go runMainWebProc(*mainProcess)
  } 
  
  serverUrl, _ := url.Parse("http://127.0.0.1:8080")
  reverseProxy := httputil.NewSingleHostReverseProxy(serverUrl)
  printToLog("Mounting reverse proxy on " + serverUrl.String() + "\n")
  http.Handle("/", reverseProxy)
 
  mount := "/" + appVersion + "/" + instanceIndex

  printToLog("Mounting console on " + mount + "\n")
  http.Handle(mount, wsHandler) 
  err := http.ListenAndServe(":" + os.Getenv("PORT"), nil)

  if err != nil {
    panic("ListenAndServe: " + err.Error())
  }
}