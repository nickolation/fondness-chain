package assets

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"os"

	"github.com/nickolation/fondness-chain/core/utils"
)

const (
	//	Path to the storage memory
	memPath = "./source/memory/memory.data"
)


//	Instance of memory.
//	Storage map provides the acces to the [addr] -> [hearts] data.
type Memory struct {
	Storage map[string]*Heart
}


//	Writes heart to memory file
func (mem *Memory) WriteMemory() {
	var buff bytes.Buffer

	gob.Register(elliptic.P256())

	enc := gob.NewEncoder(&buff)
	utils.Handle(
		"heart gob encode",
		enc.Encode(mem),
	)

	utils.Handle(
		"write heart to the file",
		ioutil.WriteFile(memPath, buff.Bytes(), 0644),
	)
}


//	Read heart data from memory file
func (mem *Memory) ReadMemory() error {
	if _, err := os.Stat(memPath); os.IsNotExist(err) {
		utils.Log(
			"file mem isn't exist",
			err,
		)

		return err
	}

	var m Memory

	info, err := ioutil.ReadFile(memPath)
	if err != nil {
		utils.Log(
			"read memfile",
			err,
		)

		return err
	}

	gob.Register(elliptic.P256())
	dec := gob.NewDecoder(bytes.NewReader(info))
	err = dec.Decode(&m)

	if err != nil {
		utils.Log(
			"decode to memory",
			err,
		)

		return err
	}

	mem.Storage = m.Storage

	return nil
}


//	Memory generator with wallets.
func AccesMemory() (*Memory, error) {
	mem := Memory{}
	mem.Storage = make(map[string]*Heart)

	err := mem.ReadMemory()
	return &mem, err
}


//	Add new heart to the memory
func (mem *Memory) LinkHeart() string {
	hrt := FeelHeart()
	addr := string(hrt.Addr())

	mem.Storage[string(addr)] = hrt
	return addr
}


//	Get heart from the map with addr
func (mem *Memory) GetHeart(addr string) *Heart {
	return mem.Storage[addr]
}


//	Get all hearts from the map with addr
func (mem *Memory) GetAddrs() []string {
	var addrs []string

	for addr := range mem.Storage {
		addrs = append(addrs, addr)
	}

	return addrs
}
