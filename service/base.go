package service

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"paste/util"

	"github.com/google/uuid"
	"golang.org/x/crypto/blake2b"
)

type PasteInfo struct {
	Title          string     `json:"title,omitempty"`
	ID        	   uuid.UUID  `json:"-"`
	Temporary 	   bool		  `json:"temporary,omitempty"`
	Duration       string     `json:"duration,omitempty"`
	ExpirationDate *time.Time `json:"-"`
	Sealed    	   bool       `json:"sealed,omitempty"`
	IsFile         bool       `json:"isfile"`
	CreationDate   time.Time  `json:"-"`
	Secrets        []byte     `json:"-"`
}

type Paste struct {
	Info    PasteInfo `json:"info"`
	Content []byte     `json:"-"`
}

func CreatePaste(data []byte) Paste {
	var p Paste = Paste{}
	var err error
	info := PasteInfo{}
	info.CreationDate = time.Now()
	utime := time.UnixMilli(info.CreationDate.UnixMilli()).String()
	info.Title = "paste-"+utime
	info.ID,err = uuid.NewRandom()
	if err != nil {
		util.LogError("Failed to generate UUID for PasteInfo",err)
	}
	p.Info = info
	p.Content = data
	return p
}

func (p* Paste) SetTitle(title string) {
	p.Info.Title = title
}

func (p* Paste) Seal(password string) (string,string,error) {
	salt := make([]byte,16)
	_,err := rand.Read(salt)
	if err != nil {
		return "","",err
	}
	hasher,err := blake2b.New512(append([]byte(password),salt...))
	if err != nil {
		return "","",err
	}
	hash := hasher.Sum(nil)
	p.Info.Sealed = true
	return hex.EncodeToString(hash),hex.EncodeToString(salt),err
}

func (p* Paste) SetExpirationDate(duration string) {
	p.Info.Duration = duration
	var err error
	var hour int
	var minute int
	var second int
	s := strings.TrimSpace(p.Info.Duration)
	for i := 0; i < len(s); i++ {
		if s[i] == 'h' {
			hour,err = strconv.Atoi(string(s[i-1]))
		} else if s[i] == 'm' {
			minute,err = strconv.Atoi(string(s[i-1]))
		} else if s[i] == 's' {
			second,err = strconv.Atoi(string(s[i-1]))
		}
	}
	if err != nil {
		util.LogError("Failed to convert string to int",err)
		return
	}
	p.Info.Temporary = true
	totalSeconds := 3600*hour + 60*minute + second
	t := p.Info.CreationDate.Add(time.Second * time.Duration(totalSeconds))
	p.Info.ExpirationDate = &t

}

func VerifyPassword(hash,salt,input string) bool {
	nsalt,err := hex.DecodeString(salt)
	if err != nil { return false }

	hasher,err := blake2b.New512(append([]byte(input),nsalt...))
	if err != nil { return false }

	finalHash := hex.EncodeToString(hasher.Sum(nil))
	return finalHash == hash
}
