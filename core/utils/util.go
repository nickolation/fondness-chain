package utils

import "log"

//	Hanlder with logger based on getting description
func Handle(des string, err error) {
	if err != nil {
		log.Printf("%s - [%v]", des, err)
	}
}
