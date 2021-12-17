package utils

import "log"

//	Hanlder with logger based on getting description.
//	Fatal the process.
func FHandle(des string, err error) {
	if err != nil {
		FLog(des, err)
	}
} 

//	Handler with the basic logic.
func Handle(des string, err error) {
	if err != nil {
		Log(des, err)
	}
} 

//	Logger with description and error
func FLog(des string, err error) {
	log.Fatalf("%s - [%v]", des, err)
}

//	The simple logger
func Log(des string, err error) {
	log.Printf("%s - [%v]", des, err)
}

//	Combintd
func HandleLog(des string, err error) error {
	Handle(des, err)
	return err
}
