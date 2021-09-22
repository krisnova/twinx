// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ────────────────────────────────────────────────────────────────────────────
//
//  ████████╗██╗    ██╗██╗███╗   ██╗██╗  ██╗
//  ╚══██╔══╝██║    ██║██║████╗  ██║╚██╗██╔╝
//     ██║   ██║ █╗ ██║██║██╔██╗ ██║ ╚███╔╝
//     ██║   ██║███╗██║██║██║╚██╗██║ ██╔██╗
//     ██║   ╚███╔███╔╝██║██║ ╚████║██╔╝ ██╗
//     ╚═╝    ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
//
// ────────────────────────────────────────────────────────────────────────────

//go:build linux
// +build linux

package twinx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/kris-nova/logger"
)

const (
	// OpenCommand is linux specific
	OpenCommand = "xdg-open"

	// LocalhostListenPort will listen onport 1717 for requests
	LocalhostListenPort = "1717"
)

type CallbackText struct {
	Title       string
	MainHeading string
	SubHeading1 string
	SubHeading2 string
}

func (d *CallbackText) render() string {
	return fmt.Sprintf(NivenlyDefaultPageTemplate, d.Title, d.MainHeading, d.SubHeading1, d.SubHeading2)
}

// LocalhostServerGetParameters will run an HTTP server on localhost
// and listen for parameters passed back over the GET request to populate
// the parameters into JSON returned over the []byte channel
func LocalhostServerGetParameters(d *CallbackText) (chan []byte, chan error) {
	errCh := make(chan error)
	iCh := make(chan []byte)
	go func() {
		http.HandleFunc("/", SharedValuesDynamic(d))
		go func() {
			http.ListenAndServe(fmt.Sprintf(":%s", LocalhostListenPort), nil)
			// TODO we need a way to close this server after we have our parameters
		}()
		for {
			localhostServerValuesMutex.Lock()
			if len(localhostServerValuesQueue) > 0 {
				requestParams := localhostServerValuesQueue[0]
				localhostServerValuesQueue = []map[string][]string{} // Reset the queue to zero
				data, err := json.Marshal(requestParams)
				if err != nil {
					errCh <- fmt.Errorf("unable to parse callback response on localhost GET server: %v", err)
				}
				localhostServerValuesMutex.Unlock()
				// Success with no error...
				errCh <- nil
				iCh <- data
				return
			}
			localhostServerValuesMutex.Unlock()
			time.Sleep(time.Second * 1)
		}
	}()
	return iCh, errCh
}

var (
	localhostServerValuesQueue []map[string][]string
	localhostServerValuesMutex sync.Mutex
)

var dynamicCallbackText = &CallbackText{
	Title:       "Twinx",
	MainHeading: "Live Streaming Application",
	SubHeading1: "Built for Linux",
	SubHeading2: "Very impressive",
}

func SharedValuesDynamic(d *CallbackText) func(w http.ResponseWriter, r *http.Request) {
	dynamicCallbackText = d
	return SharedValues
}

// SharedValues will plumb shared values back over the queue
func SharedValues(w http.ResponseWriter, r *http.Request) {
	localhostServerValuesMutex.Lock()
	defer localhostServerValuesMutex.Unlock()
	localhostServerValuesQueue = append(localhostServerValuesQueue, r.URL.Query())
	//logger.Debug(r.URL.Query().Encode())
	w.Write([]byte(dynamicCallbackText.render()))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
}

func Open(args []string) error {
	r, err := ExecCommand(OpenCommand, args)
	if err != nil {
		return err
	}
	if r.Stderr.Len() > 0 {
		return fmt.Errorf("error from open: %s", r.Stderr.String())
	}
	logger.Info(r.Stdout.String())
	return nil
}

// Browsers are default browser executables (in order) to try and use on Linux.
// taken from my personal Archlinux setup.
//
// Please feel free to add these, but do not change the order as they are important to me!
var Browsers = []string{"/usr/bin/brave", "/usr/bin/firefox", "/usr/bin/google-chrome-stable"}

// OpenInBrowserDefault will try to open a URL in one of the default browsers
func OpenInBrowserDefault(url string) error {
	success := false
	for _, browser := range Browsers {
		if Exists(browser) {
			logger.Info("Trying browser %s...", browser)
			_, err := ExecCommand(browser, []string{url})
			if err != nil {
				logger.Warning("Unable to open authentication URL: %s", err)
			}
			// Note: We ignore browser logs because
			// we are doing .Start() instead of .Run()
			// logger.Info(r.Stdout.String())
			success = true
			break
		}
	}
	if !success {
		return fmt.Errorf("unable to open authentication URL")
	}
	return nil
}

// Exists will check if a file exists
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// ExecResult is the response from executing a command
type ExecResult struct {
	Command *exec.Cmd
	Stdout  *bytes.Buffer
	Stderr  *bytes.Buffer
}

// ExecCommand is a wrapper for exec.Command but with a dedicated
// result{} struct. This works better for my brain.
func ExecCommand(cmd string, args []string) (*ExecResult, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c := exec.Command(cmd, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	err := c.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to execute command: %v", err)
	}
	return &ExecResult{
		Command: c,
		Stdout:  stdout,
		Stderr:  stderr,
	}, nil
}

// fonts/TerminusModern.ttf

const (
	NivenlyDefaultPageHeader string = "Access-Control-Allow-Origin: *"

	// NivenlyDefaultPageTemplate
	// 4 string substitutions
	//   1. Title
	//   2. Main text
	//   3. Sub 1
	//   4. Sub 2
	NivenlyDefaultPageTemplate string = `

<!DOCTYPE html>

<html lang="en">
    <head>
	<meta name="generator" content="Hugo 0.88.1" />
    
        <title>
        
            %s
        
        </title>
    
    <meta http-equiv="content-type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />


    <meta name="description" content="professional computer adult :: famous transgender queer cyber hacker :: high alpine enthusiast"/>
    <meta name="keywords" content="kris, kris nova, nova, nóva, programming, linux, engineer, blog, archive, official, sexy, nivenly, homepage, wepage, important, serious, iron clad, professional, adult, business, computer, tech, books, author, author, super nóva, supernova, github, freebsd, mountains, speaker, public, public speaker, twitch, stream, stream stream, live, nova"/>




<meta name="robots" content="noodp" />
<link rel="canonical" href="/" />


<link rel="stylesheet" href="https://nivenly.com/assets/style.css" />
<link rel="stylesheet" href="https://nivenly.com/assets/nova.css" />


<script async src="https://www.googletagmanager.com/gtag/js?id=G-C8QMK1BGKN"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', 'G-C8QMK1BGKN');
</script>


<link rel="shortcut icon" href="https://nivenly.com/favicon.ico" />





<meta name="twitter:image" content="https://nivenly.com/assets/logo/Nivenly_2.png"/>
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="nivenly.com"/>
<meta name="twitter:creator" content="@krisnova">
<meta name="twitter:site" content="@krisnova">

<meta property="og:image" content="https://nivenly.com/assets/logo/Nivenly_2.png" />
<meta property="og:title" content="nivenly.com" />
<meta property="og:type" content="website" />
<meta property="og:url" content="/" /><meta property="og:site_name" content="nivenly.com" />


<meta name="twitter:title" content="official @krisnóva website [genuine, approved, legitimate]"/>
<meta name="twitter:description" content="1' and 1=1;- - professional computer adult :: famous transgender queer cyber hacker :: high alpine enthusiast" />
<meta property="og:description" content="1' and 1=1;- -professional computer adult :: famous transgender queer cyber hacker :: high alpine enthusiast" />
<meta property="og:title" content="official @krisnóva website [genuine, approved, legitimate]" />





<link rel="alternate" type="application/rss+xml" href="https://nivenly.com/index.xml" title="nivenly.com" />

    </head>


    <body>
    <header class="header glow">

  <span class="header__inner">
    <span class="logo glow">
    <a href="/" class="logo" style="text-decoration: none">
    <span class="logo-punc">[</span>
    <span class="logo-user">nova</span>
    <span class="logo-punc">@</span>
    <span class="logo-host">nivenly</span>
    <span class="logo-punc">]: </span>
    </a>
    
    <span class="logo-host">  ~</span>
    
    <span class="logo-punc">></span><span class="logo-user">$</span>
  <span class="logo-cursor"></span>
</span>

    <span class="header__right">
      
        <nav class="menu content-desktop">
  <ul class="menu__inner menu__inner--desktop">
    <li>./</li>
    
    
      
        
          <li><a href="https://nivenly.com/activestreamer">activestreamer</a></li>
        
      
        
          <li><a href="https://www.youtube.com/channel/UCRvH2UexTzcbZRwCS6OxJ3w">archive</a></li>
        
      
        
          <li><a href="https://nivenly.com/author">author</a></li>
        
      
        
          <li><a href="https://nivenly.com/contrib">contrib</a></li>
        
      
        
          <li><a href="https://nivenly.com/lib">lib</a></li>
        
      
        
          <li><a href="https://nivenly.com/live">live</a></li>
        
      
        
          <li><a href="https://nivenly.com/questions">questions</a></li>
        
      
        
          <li><a href="https://nivenly.com/readme">readme</a></li>
        
      
        
          <li><a href="https://github.com/kris-nova/nivenly.com">src</a></li>
        
      
    
    
  </ul>

</nav>

<nav class="menu-alt content-mobile" id="mobile-menu">
  <ul>
  
  
  <li><a href="https://nivenly.com/activestreamer">activestreamer</a></li>
  
  
  
  <li><a href="https://www.youtube.com/channel/UCRvH2UexTzcbZRwCS6OxJ3w">archive</a></li>
  
  
  
  <li><a href="https://nivenly.com/author">author</a></li>
  
  
  
  <li><a href="https://nivenly.com/contrib">contrib</a></li>
  
  
  
  <li><a href="https://nivenly.com/lib">lib</a></li>
  
  
  
  <li><a href="https://nivenly.com/live">live</a></li>
  
  
  
  <li><a href="https://nivenly.com/questions">questions</a></li>
  
  
  
  <li><a href="https://nivenly.com/readme">readme</a></li>
  
  
  
  <li><a href="https://github.com/kris-nova/nivenly.com">src</a></li>
  
  
  </ul>
</nav>

      

      
      <span class="toggle-sidebar menu-icon">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512"><path fill="currentColor" d="M216 288h-48c-8.84 0-16 7.16-16 16v192c0 8.84 7.16 16 16 16h48c8.84 0 16-7.16 16-16V304c0-8.84-7.16-16-16-16zM88 384H40c-8.84 0-16 7.16-16 16v96c0 8.84 7.16 16 16 16h48c8.84 0 16-7.16 16-16v-96c0-8.84-7.16-16-16-16zm256-192h-48c-8.84 0-16 7.16-16 16v288c0 8.84 7.16 16 16 16h48c8.84 0 16-7.16 16-16V208c0-8.84-7.16-16-16-16zm128-96h-48c-8.84 0-16 7.16-16 16v384c0 8.84 7.16 16 16 16h48c8.84 0 16-7.16 16-16V112c0-8.84-7.16-16-16-16zM600 0h-48c-8.84 0-16 7.16-16 16v480c0 8.84 7.16 16 16 16h48c8.84 0 16-7.16 16-16V16c0-8.84-7.16-16-16-16z"/></svg>
</span>

      </span>

      
      <span class="toggle-contrast menu-icon menu-desktop">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path fill="currentColor" d="M8 256c0 136.966 111.033 248 248 248s248-111.034 248-248S392.966 8 256 8 8 119.033 8 256zm248 184V72c101.705 0 184 82.311 184 184 0 101.705-82.311 184-184 184z"/></svg>
</span>

      </span>

      
      <span class="toggle-mobile-menu menu-icon">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path fill="currentColor" d="M16 132h416c8.837 0 16-7.163 16-16V76c0-8.837-7.163-16-16-16H16C7.163 60 0 67.163 0 76v40c0 8.837 7.163 16 16 16zm0 160h416c8.837 0 16-7.163 16-16v-40c0-8.837-7.163-16-16-16H16c-8.837 0-16 7.163-16 16v40c0 8.837 7.163 16 16 16zm0 160h416c8.837 0 16-7.163 16-16v-40c0-8.837-7.163-16-16-16H16c-8.837 0-16 7.163-16 16v40c0 8.837 7.163 16 16 16z"/></svg>
</span>

      </span>

      
      <a href="https://discord.gg/pMtuhDJ54a" target="_blank" rel="noopener noreferrer" class="menu-icon">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path fill="currentColor" d="M297.216 243.2c0 15.616-11.52 28.416-26.112 28.416-14.336 0-26.112-12.8-26.112-28.416s11.52-28.416 26.112-28.416c14.592 0 26.112 12.8 26.112 28.416zm-119.552-28.416c-14.592 0-26.112 12.8-26.112 28.416s11.776 28.416 26.112 28.416c14.592 0 26.112-12.8 26.112-28.416.256-15.616-11.52-28.416-26.112-28.416zM448 52.736V512c-64.494-56.994-43.868-38.128-118.784-107.776l13.568 47.36H52.48C23.552 451.584 0 428.032 0 398.848V52.736C0 23.552 23.552 0 52.48 0h343.04C424.448 0 448 23.552 448 52.736zm-72.96 242.688c0-82.432-36.864-149.248-36.864-149.248-36.864-27.648-71.936-26.88-71.936-26.88l-3.584 4.096c43.52 13.312 63.744 32.512 63.744 32.512-60.811-33.329-132.244-33.335-191.232-7.424-9.472 4.352-15.104 7.424-15.104 7.424s21.248-20.224 67.328-33.536l-2.56-3.072s-35.072-.768-71.936 26.88c0 0-36.864 66.816-36.864 149.248 0 0 21.504 37.12 78.08 38.912 0 0 9.472-11.52 17.152-21.248-32.512-9.728-44.8-30.208-44.8-30.208 3.766 2.636 9.976 6.053 10.496 6.4 43.21 24.198 104.588 32.126 159.744 8.96 8.96-3.328 18.944-8.192 29.44-15.104 0 0-12.8 20.992-46.336 30.464 7.68 9.728 16.896 20.736 16.896 20.736 56.576-1.792 78.336-38.912 78.336-38.912z"/></svg>
</span>

      </a>

      
      <a href="https://twitter.com/krisnova" target="_blank" rel="noopener noreferrer" class="menu-icon">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path fill="currentColor" d="M459.37 151.716c.325 4.548.325 9.097.325 13.645 0 138.72-105.583 298.558-298.558 298.558-59.452 0-114.68-17.219-161.137-47.106 8.447.974 16.568 1.299 25.34 1.299 49.055 0 94.213-16.568 130.274-44.832-46.132-.975-84.792-31.188-98.112-72.772 6.498.974 12.995 1.624 19.818 1.624 9.421 0 18.843-1.3 27.614-3.573-48.081-9.747-84.143-51.98-84.143-102.985v-1.299c13.969 7.797 30.214 12.67 47.431 13.319-28.264-18.843-46.781-51.005-46.781-87.391 0-19.492 5.197-37.36 14.294-52.954 51.655 63.675 129.3 105.258 216.365 109.807-1.624-7.797-2.599-15.918-2.599-24.04 0-57.828 46.782-104.934 104.934-104.934 30.213 0 57.502 12.67 76.67 33.137 23.715-4.548 46.456-13.32 66.599-25.34-7.798 24.366-24.366 44.833-46.132 57.827 21.117-2.273 41.584-8.122 60.426-16.243-14.292 20.791-32.161 39.308-52.628 54.253z"/></svg>
</span>

      </a>

      
      <a href="https://github.com/kris-nova/nivenly.com" target="_blank" rel="noopener noreferrer" class="menu-icon">
        
<span class="inline-svg"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 496 512"><path fill="currentColor" d="M165.9 397.4c0 2-2.3 3.6-5.2 3.6-3.3.3-5.6-1.3-5.6-3.6 0-2 2.3-3.6 5.2-3.6 3-.3 5.6 1.3 5.6 3.6zm-31.1-4.5c-.7 2 1.3 4.3 4.3 4.9 2.6 1 5.6 0 6.2-2s-1.3-4.3-4.3-5.2c-2.6-.7-5.5.3-6.2 2.3zm44.2-1.7c-2.9.7-4.9 2.6-4.6 4.9.3 2 2.9 3.3 5.9 2.6 2.9-.7 4.9-2.6 4.6-4.6-.3-1.9-3-3.2-5.9-2.9zM244.8 8C106.1 8 0 113.3 0 252c0 110.9 69.8 205.8 169.5 239.2 12.8 2.3 17.3-5.6 17.3-12.1 0-6.2-.3-40.4-.3-61.4 0 0-70 15-84.7-29.8 0 0-11.4-29.1-27.8-36.6 0 0-22.9-15.7 1.6-15.4 0 0 24.9 2 38.6 25.8 21.9 38.6 58.6 27.5 72.9 20.9 2.3-16 8.8-27.1 16-33.7-55.9-6.2-112.3-14.3-112.3-110.5 0-27.5 7.6-41.3 23.6-58.9-2.6-6.5-11.1-33.3 2.6-67.9 20.9-6.5 69 27 69 27 20-5.6 41.5-8.5 62.8-8.5s42.8 2.9 62.8 8.5c0 0 48.1-33.6 69-27 13.7 34.7 5.2 61.4 2.6 67.9 16 17.7 25.8 31.5 25.8 58.9 0 96.5-58.9 104.2-114.8 110.5 9.2 7.9 17 22.9 17 46.4 0 33.7-.3 75.4-.3 83.6 0 6.5 4.6 14.4 17.3 12.1C428.2 457.8 496 362.9 496 252 496 113.3 383.5 8 244.8 8zM97.2 352.9c-1.3 1-1 3.3.7 5.2 1.6 1.6 3.9 2.3 5.2 1 1.3-1 1-3.3-.7-5.2-1.6-1.6-3.9-2.3-5.2-1zm-10.8-8.1c-.7 1.3.3 2.9 2.3 3.9 1.6 1 3.6.7 4.3-.7.7-1.3-.3-2.9-2.3-3.9-2-.6-3.6-.3-4.3.7zm32.4 35.6c-1.6 1.3-1 4.3 1.3 6.2 2.3 2.3 5.2 2.6 6.5 1 1.3-1.3.7-4.3-1.3-6.2-2.2-2.3-5.2-2.6-6.5-1zm-11.4-14.7c-1.6 1-1.6 3.6 0 5.9 1.6 2.3 4.3 3.3 5.6 2.3 1.6-1.3 1.6-3.9 0-6.2-1.4-2.3-4-3.3-5.6-2z"/></svg>
</span>

      </a>

    </span>
  </span>
</header>



<div class="content">
<center>
<div class="main">

    <div class="main-logo img">
    <a href="https://nivenly.com/bio">
        <img src="https://nivenly.com/assets/logo/Nivenly_2.png" width="450" height="auto" style="padding-bottom: 20px">
        <span class="logo-host glow" style="font-size: 1.3rem">[</span>
        <span class="logo-user glow" style="font-size: 1.3rem">%s</span>
        <span class="logo-host glow" style="font-size: 1.3rem">]</span>
    </a>
</div>
</center>
<center>
<div class="glow">

    <div class="logo-host glow" style="font-size: 1.1rem; padding-top: 20px">
		%s
	</div>
    <div class="logo-user glow" style="font-size: 1.1rem">
		%s
    </div>

</div>
</center>

</div>

            </div>

        
        <footer class="footer">
    

  <div class="content-paragraph" style="text-align: center">
      <div style="padding-bottom: 10px">© 2021 Kris Nóva</div>
      <div class="color-footer-body" style="text-decoration: none">
        Nivenly.com and its related assets, tools, and API endpoints are provided for
        convenience and educational purposes only.
      </div>
      <div class="color-footer-body" style="text-decoration: none">
          The website's source code and its Apache2 license is hosted
        online at <a style="text-decoration: none" href="https://github.com/kris-nova/nivenly.com">github.com/kris-nova/nivenly.com</a>.
      </div>
      <div class="color-footer-body" style="text-decoration: none">
        By using this website you agree to never use these tools for any illegal or malicious activity, and to adhere to ethical and responsible disclosure and practice.
        The owner and operator of this website accepts no responsibility for the usage of these tools in any way.
      </div>
      <div class="color-footer-body" style="text-decoration: none">
          These tools should not be used by anyone.
      </div>
  </div>
</footer>


        
    </div>
    
    </body>
</html>
`
)
