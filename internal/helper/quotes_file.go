package helper

/**
 * Read once, use multiple times
 */

import (
	"github.com/goccy/go-json"
	"github.com/kattana-io/tron-blocks-parser/internal/models"
	"io/ioutil"
	"sync"
)

type QuotesFile struct {
	Quotes []models.QuotePair
	lock   *sync.Mutex
}

func NewQuotesFile() *QuotesFile {
	raw, err := ioutil.ReadFile("quotes.json")
	if err != nil {
		return nil
	}
	cFile := models.ConfigFile{}
	err = json.Unmarshal(raw, &cFile)
	if err != nil {
		return nil
	}

	return &QuotesFile{
		Quotes: cFile.Quotes,
		lock:   &sync.Mutex{},
	}
}

func (q *QuotesFile) Get() []models.QuotePair {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.Quotes
}
