package util

import (
	"bufio"
	"os"
)

type Loglevel struct {
   Header string
   IsError bool
}

func writeLog(msg string,target *os.File) {
   writer := bufio.NewWriter(target)
   defer writer.Flush()
   writer.WriteString(msg+"\n")
}

func LogInfo(msg string) {
   Custom(&Loglevel{Header:"INFO"},msg)(msg)
}

func LogWarn(msg string) {
   Custom(&Loglevel{Header:"WARN"},msg)(msg)
}

func LogError(msg string,err error) {
   if err == nil {
	   Custom(&Loglevel{Header:"ERROR"},msg)(msg)
   } else {
	   Custom(&Loglevel{Header:"ERROR"},msg+" "+err.Error())(msg+" "+err.Error())
   }
}

func Custom(level *Loglevel,msg string) func(msg string) {
   if level.IsError {
      return func(string) {
         writeLog(level.Header+": "+msg,os.Stderr)
      }
   } else {
      return func(string) {
         writeLog(level.Header+": "+msg,os.Stdout)
      }
   }
}
