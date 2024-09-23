package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

type Format interface {
	Transform(data any, writer io.Writer) error
}

type JSON struct{}

func (JSON) Transform(data any, writer io.Writer) error {
	result, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = writer.Write(result)
	return err
}

type CVS struct{}

func (CVS) Transform(data interface{}, writer io.Writer) error {
	records, err := toCSVRecords(data)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	return csvWriter.WriteAll(records)
}

func toCSVRecords(data interface{}) ([][]string, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %s", v.Kind())
	}

	var records [][]string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		var record []string
		for j := 0; j < elem.NumField(); j++ {
			record = append(record, fmt.Sprintf("%v", elem.Field(j).Interface()))
		}
		records = append(records, record)
	}
	return records, nil
}

func Write() map[string]func(data any, writer io.Writer) {
	json, cvs := &JSON{}, &CVS{}

	return map[string]func(data any, writer io.Writer){
		"json": func(data any, writer io.Writer) {
			json.Transform(data, writer)
		},
		"cvs": func(data any, writer io.Writer) {
			cvs.Transform(data, writer)
		},
	}
}

type Human struct {
	Name     string
	Age      int
	Location string
}

func New(name string, age int, location string) *Human {
	return &Human{
		Name:     name,
		Age:      age,
		Location: location,
	}
}

func main() {
	file, err := os.Create("./test.txt")
	if err != nil {
		log.Println(err)
	}

	defer file.Close()

	formate := Write()

	human := New("John", 21, "NYC")
	formate["json"](human, file)

	buf := bytes.NewBuffer(make([]byte, 4096))
	formate["json"](human, buf)

	fmt.Println(buf)
}
