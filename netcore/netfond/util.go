package netfond

import (
	"bytes"
	"encoding/gob"

	"github.com/nickolation/fondness-chain/netcore/utils"
)

func EncodeCmd(cmd string) []byte {
	var buff [cmdClaim]byte

	for i, c := range cmd {
		buff[i] = byte(c)
	}

	return buff[:]
}

func DecodeCmd(buff []byte) string {
	var cmd []byte 

	for _, b := range buff {
		if b != 0x0 {
			cmd = append(cmd, b)
		} 
	}

	return string(cmd)
}


func NodeExistance(addr string) bool {
	for _, a := range listNodes {
		if a == addr {
			return true
		}
	}

	return false
}


func Encode(object interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	utils.FHandle(
		"object encoding to the bytes",
		enc.Encode(object),
	) 

	return buff.Bytes()
}

