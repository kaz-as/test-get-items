package getitems

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unsafe"

	"github.com/iancoleman/orderedmap"
)

const (
	idName = "id"
)

var (
	idAbsentInDataResponse = []byte("id is absent in data")
)

type Handler struct {
	headers []string
	hMap    map[string]int
	data    map[string][]byte
}

func NewHandler(csvPath string) (*Handler, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("open failed: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("close failed: %s", err)
		}
	}()

	r := csv.NewReader(f)
	line, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read first line: %w", err)
	}

	h := &Handler{
		data: make(map[string][]byte),
		hMap: make(map[string]int),
	}

	for i, k := range line {
		if _, ok := h.hMap[k]; ok {
			return nil, fmt.Errorf("header %s is duplicated", k)
		}
		h.hMap[k] = i
		h.headers = append(h.headers, k)
	}

	if _, ok := h.hMap[idName]; !ok {
		return nil, fmt.Errorf("id is absent in header")
	}

	line, err = r.Read()
	for err == nil {
		if len(line) > 0 {
			if len(line) != len(h.headers) {
				return nil, fmt.Errorf("line number %d has %d fields, but %d needed", len(h.data)+1, len(line), len(h.headers))
			}

			o := orderedmap.New()

			if _, exists := h.data[line[h.hMap[idName]]]; exists {
				return nil, fmt.Errorf("id %s already exists", line[h.hMap[idName]])
			}

			for i := 0; i < len(line); i++ {
				o.Set(h.headers[i], line[i])
			}

			jsn, err := json.Marshal(o)
			if err != nil {
				return nil, fmt.Errorf("marshalling line number %d: %w", len(h.data)+1, err)
			}

			idI, _ := o.Get(idName)
			id := idI.(string)

			h.data[id] = jsn
		}

		line, err = r.Read()
	}

	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("error reading a file: %w", err)
	}

	return h, nil
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ids := request.URL.Query()[idName]
	if len(ids) == 0 {
		write(writer, http.StatusOK, []byte{'[', ']'})
		return
	}

	b := strings.Builder{}
	b.WriteByte('[')

	var entry []byte
	var ok bool
	if entry, ok = h.data[ids[0]]; !ok {
		write(writer, http.StatusNotFound, idAbsentInDataResponse)
		return
	}
	b.Write(entry)

	for i := 1; i < len(ids); i++ {
		id := ids[i]
		b.WriteByte(',')

		if entry, ok = h.data[id]; !ok {
			write(writer, http.StatusNotFound, idAbsentInDataResponse)
			return
		}

		b.Write(entry)
	}

	b.WriteByte(']')

	str := b.String()

	// safe to convert directly to slice, because it was just converted from it
	write(writer, http.StatusOK, *(*[]byte)(unsafe.Pointer(&str)))
}

func write(writer http.ResponseWriter, status int, body []byte) {
	if status == 0 {
		status = 500
	}
	writer.WriteHeader(status)
	_, err := writer.Write(body)
	if err != nil {
		log.Printf("cannot write response")
	}
	return
}
