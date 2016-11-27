package backend

type Item struct {
	Id   string            `json:"id"`
	Tags map[string]string `json:"tags"`
}

type DataItem struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type StatItem struct {
	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Empty   bool    `json:"empty"`
	Samples int64   `json:"samples"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	Median  float64 `json:"median"`
	Sum     float64 `json:"sum"`
}

type Backend interface {
	GetItemList(tags map[string]string) []Item
	GetRawData(id string, end int64, start int64, limit int64, order string) []DataItem
	GetStatData(id string, end int64, start int64, limit int64, order string, bucketDuration string) []StatItem
}
