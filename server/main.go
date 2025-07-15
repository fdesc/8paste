package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"paste/service"
	"paste/util"
)


var pasteStorage = make(map[string]service.Paste)
var expirationStorage = make(map[string]time.Time)
var storageMutex = &sync.RWMutex{}

type PageData struct {
	PageTitle   string
	ContentData string
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.LogError("Got request with a different method instead of POST",nil)
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	var err error
	var p service.Paste

	err = r.ParseMultipartForm(4096 << 20)
	if err != nil {
		if err == http.ErrNotMultipart {
			util.LogError("Invalid Content-Type received",err)
			http.Error(w,err.Error(),http.StatusBadRequest)
			return
		}
		util.LogError("Failed to parse form",err)
		http.Error(w,err.Error(),http.StatusBadRequest)
		return
	}

	var info service.PasteInfo
	infojson := ""
	values := r.MultipartForm.Value["info"]
	if len(values) > 0 {
		infojson = values[0]
	}
	if infojson == "" {
		util.LogError("Got missing PasteInfo",nil)
		http.Error(w,"Missing info field",http.StatusBadRequest)
		return
	}

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
		p = service.CreatePaste(buf.Bytes())
	} else {
		p = service.CreatePaste([]byte(r.FormValue("content")))
	}

	if info.Sealed {
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
	if info.Temporary { p.SetExpirationDate(info.Duration) }
	if p.Info.Title != info.Title && info.Title != "" { p.SetTitle(info.Title) }

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

func servePageContent(w http.ResponseWriter,p service.Paste,target string) {
	tmpl,err := template.ParseFiles(target)
	if err != nil {
		util.LogError("Failed to parse HTML",err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		PageTitle: p.Info.Title,
		ContentData: string(p.Content),
	}

	tmpl.Execute(w,data)
}

func idHandler(w http.ResponseWriter,r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path,"/")
	if id == "" {
		util.LogError("Provided invalid ID for paste",nil)
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	storageMutex.RLock()
	paste,exists := pasteStorage[id]
	storageMutex.RUnlock()

	if !exists {
		util.LogError("Got non-existent Paste ID",nil)
		http.Error(w,"Not found",http.StatusNotFound)
		return
	}

	if paste.Info.IsFile {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s",paste.Info.Title))
		w.Header().Set("Content-Type", "application/octet-stream")
	} else {
		w.Header().Set("Content-Type","text/plain")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if paste.Info.Sealed {
		password := r.FormValue("password")
		if password == "" {
			servePageContent(w,paste,"static/locked.html")
			return
		}
		pieces := strings.Split(string(paste.Info.Secrets),":")
		if len(pieces) != 2 || !service.VerifyPassword(pieces[0],pieces[1],password) {
			util.LogError("Invalid password submitted",nil)
			http.Error(w, "Invalid password", http.StatusUnauthorized)
            return
		} else {
			servePageContent(w,paste,"static/index.html")
			return
		}
	} else {
		servePageContent(w,paste,"static/index.html")
	}

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
	http.HandleFunc("/",idHandler)

	port := "8080"
	util.LogInfo("Starting server at port "+port)
	err := http.ListenAndServe(":"+port,nil)
	if (err != nil) {
		util.LogError("Server threw an error!",err)
	}
}
