package memory

import (
	"testing"
	"github.com/MohawkTSDB/mohawk/storage"
	"math/rand"
	"github.com/magiconair/properties/assert"
)

var backend Backend

func TestBackend_GetLastDataItems(t *testing.T) {
	// Init backend object.
	backend.Open(nil)
	// Create some fake data.
	data := make([]storage.DataItem, 0)
	for i:=0;i<10;i++ {
		data = append(data, storage.DataItem{
			Timestamp:int64(i*30000),
			Value:rand.Float64(),
		})
	}
	// Post data to backend
	for _, item := range data {
		backend.PostRawData("test_tenant","test_metric",item.Timestamp,item.Value)
	}
	// Retrieve five last values from backend.
	items, _ := backend.GetLastDataItems("test_tenant","test_metric",5)
	expected := data[len(data)-5:]

	assert.Equal(t,items,expected)
}
