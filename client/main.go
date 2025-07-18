package main

import (
	"mime/multipart"
	"encoding/json"
	"net/http"
	"strconv"
	"bytes"
	"flag"
	"time"
	"fmt"
	"io"
	"os"

	"github.com/fdesc/8paste/service"
	"github.com/fdesc/8paste/util"
)

func getPasteInfo(id string,source string) service.PasteInfo {
	body := &bytes.Buffer{}
	writer:= multipart.NewWriter(body)
	data := map[string]string{
		"id":id,
	}

	datajson,err := json.Marshal(data)
	if err != nil {
		util.LogError("Failed to marshal JSON",err)
		return service.PasteInfo{}
	}

	writer.WriteField("id",string(datajson))
	writer.Close()

	req,err := http.NewRequest("POST",source+"/info",body)
	if err != nil {
		util.LogError("Failed to create request",err)
		return service.PasteInfo{}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp,err := client.Do(req)
	if err != nil {
		util.LogError("Failed to send request",err)
		return service.PasteInfo{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		status := strconv.Itoa(resp.StatusCode)
		util.LogError("Unknown status code received from request expected 201 got "+status,nil)
		return service.PasteInfo{}
	}

	var info service.PasteInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		util.LogError("Failed to unmarshal JSON",err)
		return service.PasteInfo{}
	}
	return info
}

func getPaste(id string,password string,source string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	info := getPasteInfo(id,source)

	iddata := map[string]string{
		"id":id,
	}
	idjson,err := json.Marshal(iddata)
	if err != nil {
		util.LogError("Failed to marshal JSON",err)
		return
	}

	writer.WriteField("id",string(idjson))
	if info.Sealed {
		if password == "" {
			var p string
			fmt.Println("This paste is password protected, type the password in order to get access")
			fmt.Scan(&p)
			writer.WriteField("password",p)
		}
		writer.WriteField("password",password)
	}

	writer.Close()

	req,err := http.NewRequest("POST",source+"/download",body)
	if err != nil {
		util.LogError("Failed to create request",err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp,err := client.Do(req)
	if err != nil {
		util.LogError("Failed to send request",err)
		return
	}
	defer resp.Body.Close()

	data,err := io.ReadAll(resp.Body)
	if err != nil {
		util.LogError("Failed to read response body",err)
		return
	}
	os.Stdout.Write(data)
}

func uploadPaste(p *service.Paste,password string,source string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	infojson,err := json.Marshal(p.Info)
	if err != nil {
		util.LogError("Failed to marshal JSON",err)
		return
	}

	writer.WriteField("info",string(infojson))
	if p.Info.IsFile {
		part,_ := writer.CreateFormFile("content",p.Info.Title)
		io.Copy(part,bytes.NewReader(p.Content))
	} else {
		writer.WriteField("content",string(p.Content))
	}

	if p.Info.Sealed && password != "" {
		writer.WriteField("password",password)
	}
	writer.Close()

	req,err := http.NewRequest("POST",source+"/upload",body)
	if err != nil {
		util.LogError("Failed to create request",err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp,err := client.Do(req)
	if err != nil {
		util.LogError("Failed to send request",err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		status := strconv.Itoa(resp.StatusCode)
		util.LogError("Unknown status code received from request expected 201 got "+status,nil)
		return
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		util.LogError("Failed to unmarshal JSON",err)
		return
	}
	fmt.Println(source+"/"+result["id"])
}

func main() {
	var id   string
	var path string
	var password string
	var title string
	var duration string
	var data string
	var target string

	flag.Usage = func() {
		fmt.Println("8paste-client: Usage")
		fmt.Println("  -u <content>:      Upload the specified content")
		fmt.Println("  -g <ID>:           Get the content from ID")
		fmt.Println("  -f <path>:         Upload the specified file from path")
		fmt.Println("  -p <password>:     Set a password for uploaded content")
		fmt.Println("  -t <title>:        Set title for the content")
		fmt.Println("  -d x(h) y(m) z(s): Set duration for the uploaded content as in x hours y minutes z seconds")
		fmt.Println("  -source <URL>:     Set the source to get/upload contents from")
	}

	flag.StringVar(&data,"u","","Upload the specified content")
	flag.StringVar(&id,"g","","Get the content from ID")
	flag.StringVar(&path,"f","","Upload a file from specified path")
	flag.StringVar(&password,"p","","Set password for the uploaded content")
	flag.StringVar(&title,"t","","Set title for the uploaded content")
	flag.StringVar(&duration,"d","","Set duration for the uploaded content as in x hours, y minutes and z seconds")
	flag.StringVar(&target,"source","","Set the source to get/upload contents from")
	flag.Parse()

	var isfile bool
	var sealed bool
	var temp   bool
	var p service.Paste
	if (data != "") {
		isfile = false
	} else {
		isfile = true
	}
	if (password != "") {
		sealed = true
	} else {
		sealed = false
	}
	if (duration != "") {
		temp = true
	} else {
		temp = false
	}

	if (id == "") {
		if (isfile) {
			file,err := os.ReadFile(path)
			if err != nil {
				util.LogError("Failed to open file from path",err)
				return
			}
			p = service.CreatePaste(file,title,temp,isfile,sealed)
		} else {
			p = service.CreatePaste([]byte(data),title,temp,isfile,sealed)
		}
		if temp && duration != "" {
			p.SetExpirationDate(duration)
		}
		if sealed {
			uploadPaste(&p,password,target)
		} else {
			uploadPaste(&p,"",target)
		}
	} else {
		if (sealed) {
			getPaste(id,password,target)
		} else {
			getPaste(id,"",target)
		}
	}
}
