package main

import (
	"encoding/json"
	"html/template"
	"unicode/utf8"
	"net/http"
	"strings"
	"bytes"
	"flag"
	"sync"
	"time"
	"fmt"
	"io"

	"github.com/fdesc/8paste/service"
	"github.com/fdesc/8paste/util"
)


type Visitor struct {
	IP           string
	LastSeen     time.Time
	RequestCount int
}

type PageData struct {
	PageTitle   	   string
	ContentData 	   string
}

var pasteStorage = make(map[string]service.Paste)
var expirationStorage = make(map[string]time.Time)
var visitorStorage = make(map[string]*Visitor)
var storageMutex = &sync.RWMutex{}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	if !ensurePOST(w,r) { return }
	if !checkRate(w,r) { return }

	var err error
	var p service.Paste

	if !getMultiPart(w,r,4096) { return }

	var info service.PasteInfo
	var infojson string
	if !getJSON(w,r,&infojson,"info") { return }

	d := json.NewDecoder(strings.NewReader(infojson))
	d.DisallowUnknownFields()
	err = d.Decode(&info)
	if err != nil {
		util.LogError("Failed to unmarshal JSON",err)
		http.Error(w,err.Error(),http.StatusBadRequest)
		return
	}

	if info.IsFile {
		file,_,err := r.FormFile("content")
		defer file.Close()
		if err != nil {
			util.LogError("Failed to read form file",err)
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, file)
		if err != nil {
			util.LogError("Failed to copy content to buffer",err)
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		p = service.CreatePaste(buf.Bytes(),info.Title,info.Temporary,info.IsFile,info.Sealed)
	} else {
		p = service.CreatePaste([]byte(r.FormValue("content")),info.Title,info.Temporary,info.IsFile,info.Sealed)
	}

	if p.Info.Sealed {
		password := r.FormValue("password")
		if password == "" {
			util.LogError("Attempted to submit empty password",nil)
			http.Error(w,"Invalid password",http.StatusInternalServerError)
			return
		}
		hash,salt,err := p.Seal(password)
		if err != nil {
			http.Error(w,"Password processing failed",http.StatusInternalServerError)
			return
		}
		p.Info.Secrets = []byte(fmt.Sprintf("%s:%s",hash,salt))
	}
	if p.Info.Temporary { p.SetExpirationDate(info.Duration) }
	if len(p.Info.Title) > 40 {
		p.Info.Title = p.Info.Title[0:41]
	}

	if err != nil {
		util.LogError("Failed to upload paste content",err)
	}

	storageMutex.Lock()
	defer storageMutex.Unlock()
	pasteStorage[p.Info.ID.String()] = p
	if p.Info.Temporary && p.Info.ExpirationDate != nil {
		expirationStorage[p.Info.ID.String()] = *p.Info.ExpirationDate
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id": p.Info.ID.String(),
	})
	util.LogInfo("Uploaded paste")
}

func downloadHandler(w http.ResponseWriter,r *http.Request) {
	setSecurityHeaders(w);
	if !ensurePOST(w,r) { return }
	if !checkRate(w,r) { return }

	var err error

	if !getMultiPart(w,r,10) { return }

	id := struct {
		ID string `json:"id"`
	}{}

	var idjson string
	if !getJSON(w,r,&idjson,"id") { return }

	d := json.NewDecoder(strings.NewReader(idjson))
	d.DisallowUnknownFields()
	err = d.Decode(&id)
	if err != nil {
		util.LogError("Failed to unmarshal JSON",err)
		http.Error(w,err.Error(),http.StatusBadRequest)
		return
	}

	p,exists := pasteStorage[id.ID]
	if !exists {
		util.LogError("Got non-existent Paste ID",nil)
		http.Error(w,"Not found",http.StatusNotFound)
		return
	}

	if p.Info.Sealed {
		password := r.FormValue("password")
		pieces := strings.Split(string(p.Info.Secrets),":")
		if len(pieces) != 2 || !service.VerifyPassword(pieces[0],pieces[1],password) || password == "" {
			util.LogError("Invalid password submitted",nil)
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s",p.Info.Title))
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w,bytes.NewBuffer(p.Content))
	util.LogInfo("Paste download request success")
}

func infoHandler(w http.ResponseWriter,r *http.Request) {
	setSecurityHeaders(w)
	if !ensurePOST(w,r) { return }
	if !checkRate(w,r) { return }

	var err error
	if !getMultiPart(w,r,10) { return }

	id := struct {
		ID string `json:"id"`
	}{}

	var idjson string
	if !getJSON(w,r,&idjson,"id") { return }

	d := json.NewDecoder(strings.NewReader(idjson))
	d.DisallowUnknownFields()
	err = d.Decode(&id)
	if err != nil {
		util.LogError("Failed to unmarshal JSON",err)
		http.Error(w,err.Error(),http.StatusBadRequest)
		return
	}

	p,exists := pasteStorage[id.ID]
	if !exists {
		util.LogError("Got non-existent Paste ID",nil)
		http.Error(w,"Not found",http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p.Info)
}

func idHandler(w http.ResponseWriter,r *http.Request) {
	setSecurityHeaders(w)
	if !checkRate(w,r) { return }
	id := strings.TrimPrefix(r.URL.Path,"/")
	if id == "" {
		util.LogError("Provided invalid ID for paste",nil)
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	storageMutex.RLock()
	p,exists := pasteStorage[id]
	storageMutex.RUnlock()

	if !exists {
		util.LogError("Got non-existent Paste ID",nil)
		http.Error(w,"Not found",http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if p.Info.Sealed {
		password := r.FormValue("password")
		if password == "" {
			servePageContent(w,&p,"static/locked.html")
			return
		}
		pieces := strings.Split(string(p.Info.Secrets),":")
		if len(pieces) != 2 || !service.VerifyPassword(pieces[0],pieces[1],password) {
			util.LogError("Invalid password submitted",nil)
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		} else {
			servePageContent(w,&p,"static/index.html")
			return
		}
	} else {
		servePageContent(w,&p,"static/index.html")
	}

}

func servePageContent(w http.ResponseWriter,p* service.Paste,target string) {
	tmpl,err := template.ParseFiles(target)
	if err != nil {
		util.LogError("Failed to parse HTML",err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		PageTitle: p.Info.Title,
	}

	if p.Info.IsFile {
		if !utf8.Valid(p.Content) {
			data.ContentData = "Cannot preview file."
		} else {
			data.ContentData = string(p.Content)
		}
	} else {
		data.ContentData = string(p.Content)
	}

	tmpl.Execute(w,data)
}

func getJSON(w http.ResponseWriter,r* http.Request,jsondata* string,field string) bool {
	values := r.MultipartForm.Value[field]
	if len(values) > 0 {
		*jsondata = values[0]
	}
	if *jsondata == "" {
		util.LogError("Got missing "+field,nil)
		http.Error(w,field+" is required",http.StatusBadRequest)
		return false
	}
	return true
}

func getMultiPart(w http.ResponseWriter,r *http.Request,size int64) bool {
	var err error
	err = r.ParseMultipartForm(size << 20)
	if err != nil {
		if err == http.ErrNotMultipart {
			util.LogError("Invalid Content-Type received",err)
			http.Error(w,err.Error(),http.StatusBadRequest)
			return false
		}
		util.LogError("Failed to parse form",err)
		http.Error(w,err.Error(),http.StatusBadRequest)
		return false
	}
	return true
}

func checkRate(w http.ResponseWriter,r *http.Request) bool {
	storageMutex.RLock()
	ip := getIP(r)
	addVisitor(ip)
	storageMutex.RUnlock()

	if time.Now().Sub(visitorStorage[ip].LastSeen).Seconds() < 30.0 && visitorStorage[ip].RequestCount > 6 {
		util.LogWarn("Got too many requests in less than 30 seconds")
		http.Error(w,"Too many requests",http.StatusTooManyRequests)
		return false
	}
	return true
}

func ensurePOST(w http.ResponseWriter,r* http.Request) bool {
	if r.Method != http.MethodPost {
		util.LogError("Got request with a different method instead of POST",nil)
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy",
	"default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' https://fonts.googleapis.com https://iosevka-webfonts.github.io; " +
	"font-src 'self' https://fonts.gstatic.com https://iosevka-webfonts.github.io; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'")
}

func getIP(r *http.Request) string {
	f := r.Header.Get("X-Forwarded-For")
	if f != "" {
		return strings.Split(f,",")[0]
	}
	return r.RemoteAddr
}

func addVisitor(IP string) {
	v,exists := visitorStorage[IP]
	if !exists {
		v = &Visitor{}
		v.IP = IP
		v.LastSeen = time.Now()
		visitorStorage[IP] = v
	}
	v.RequestCount++
}

func expirationLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		storageMutex.Lock()
		for k,v := range expirationStorage {
			if time.Now().After(v) {
				delete(pasteStorage,k)
				delete(expirationStorage,k)
				util.LogInfo("Removed expired paste")
			}
		}
		storageMutex.Unlock()
	}

}

func main() {
	go expirationLoop()
	http.Handle("/static/",http.StripPrefix("/static/",http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/upload",uploadHandler)
	http.HandleFunc("/info",infoHandler)
	http.HandleFunc("/download",downloadHandler)
	http.HandleFunc("/",idHandler)

	var port string
	flag.StringVar(&port,"p","","Start the server at the specified port")
	flag.Parse()
	if port == "" {
		util.LogError("A port must be specified! Exiting...",nil)
		return
	}

	s := &http.Server{
		Addr: ":"+port,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	util.LogInfo("Started server at port "+port)
	err := s.ListenAndServe()
	if (err != nil) {
		util.LogError("Server threw an error!",err)
	}
}
