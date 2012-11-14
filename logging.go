package main

import(
	"fmt"
	"time"
)

func LogError(ertype string, message string, err error){
	fmt.Printf("%s\t%s: %s: %s\n", time.Now().String(), ertype, message, err)
}

func LogIt(ertype string, message string){
	fmt.Printf("%s\t%s: %s\n", time.Now().String(), ertype, message)
}

func ErrorCheck(err error, message string) (bool){
	if(err != nil){
		LogError("ERROR", message, err)
		return true
	}
	return false
}
