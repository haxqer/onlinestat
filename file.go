package main

import (
	"bytes"
	"encoding/gob"
	"github.com/haxqer/gintools/file"
	"io/ioutil"
	"path/filepath"
)

func store(fileName string, data interface{}) error {
	var err error
	err = file.IsNotExistMkDir(filepath.Dir(fileName))
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, buffer.Bytes(), 0600)
}

func load(fileName string, data interface{}) error {
	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(raw)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(data)
}
