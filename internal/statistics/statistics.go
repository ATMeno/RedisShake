package statistics

import (
	"encoding/json"
	"github.com/alibaba/RedisShake/internal/config"
	"github.com/alibaba/RedisShake/internal/log"
	"net/http"
	"time"
)

type metrics struct {
	// info
	Address string `json:"address"`

	// entries
	EntryId              uint64 `json:"entry_id"`
	AllowEntriesCount    uint64 `json:"allow_entries_count"`
	DisallowEntriesCount uint64 `json:"disallow_entries_count"`

	// rdb
	RdbFileSize     uint64 `json:"rdb_file_size"`
	RdbReceivedSize uint64 `json:"rdb_received_size"`
	RdbSendSize     uint64 `json:"rdb_send_size"`

	// aof
	AofReceivedOffset uint64 `json:"aof_received_offset"`
	AofAppliedOffset  uint64 `json:"aof_applied_offset"`

	// for performance debug
	InQueueEntriesCount  uint64 `json:"in_queue_entries_count"`
	UnansweredBytesCount uint64 `json:"unanswered_bytes_count"`
}

var Metrics = &metrics{}

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(Metrics)
	if err != nil {
		log.PanicError(err)
	}
}

func Init() {
	go func() {
		seconds := config.Config.Advanced.LogInterval
		if seconds <= 0 {
			log.Infof("statistics disabled. seconds=[%d]", seconds)
		}

		lastAllowEntriesCount := Metrics.AllowEntriesCount
		lastDisallowEntriesCount := Metrics.DisallowEntriesCount

		for range time.Tick(time.Duration(seconds) * time.Second) {
			if Metrics.RdbFileSize == 0 {
				continue
			}
			if Metrics.RdbSendSize > Metrics.RdbReceivedSize {
				log.Infof("receiving rdb. percent=[%.2f]%%, rdbFileSize=[%.3f]G, rdbReceivedSize=[%.3f]G",
					float64(Metrics.RdbReceivedSize)/float64(Metrics.RdbFileSize)*100,
					float64(Metrics.RdbFileSize)/1024/1024/1024,
					float64(Metrics.RdbReceivedSize)/1024/1024/1024)
			} else if Metrics.RdbFileSize > Metrics.RdbSendSize {
				log.Infof("syncing rdb. percent=[%.2f]%%, allowOps=[%.2f], disallowOps=[%.2f], entryId=[%d], InQueueEntriesCount=[%d], unansweredBytesCount=[%d]bytes, rdbFileSize=[%.3f]G, rdbSendSize=[%.3f]G",
					float64(Metrics.RdbSendSize)*100/float64(Metrics.RdbFileSize),
					float32(Metrics.AllowEntriesCount-lastAllowEntriesCount)/float32(seconds),
					float32(Metrics.DisallowEntriesCount-lastDisallowEntriesCount)/float32(seconds),
					Metrics.EntryId,
					Metrics.InQueueEntriesCount,
					Metrics.UnansweredBytesCount,
					float64(Metrics.RdbFileSize)/1024/1024/1024,
					float64(Metrics.RdbSendSize)/1024/1024/1024)
			} else {
				log.Infof("syncing aof. allowOps=[%.2f], disallowOps=[%.2f], entryId=[%d], InQueueEntriesCount=[%d], unansweredBytesCount=[%d]bytes, diff=[%d], aofReceivedOffset=[%d], aofAppliedOffset=[%d]",
					float32(Metrics.AllowEntriesCount)/float32(seconds),
					float32(Metrics.DisallowEntriesCount)/float32(seconds),
					Metrics.EntryId,
					Metrics.InQueueEntriesCount,
					Metrics.UnansweredBytesCount,
					Metrics.AofReceivedOffset-Metrics.AofAppliedOffset,
					Metrics.AofReceivedOffset,
					Metrics.AofAppliedOffset)
			}

			lastAllowEntriesCount = Metrics.AllowEntriesCount
			lastDisallowEntriesCount = Metrics.DisallowEntriesCount
		}
	}()
}

// entry id

func UpdateEntryId(id uint64) {
	Metrics.EntryId = id
}
func AddAllowEntriesCount() {
	Metrics.AllowEntriesCount++
}
func AddDisallowEntriesCount() {
	Metrics.DisallowEntriesCount++
}

// rdb

func SetRDBFileSize(size uint64) {
	Metrics.RdbFileSize = size
}
func UpdateRDBReceivedSize(size uint64) {
	Metrics.RdbReceivedSize = size
}
func UpdateRDBSentSize(offset uint64) {
	Metrics.RdbSendSize = offset
}

// aof

func UpdateAOFReceivedOffset(offset uint64) {
	Metrics.AofReceivedOffset = offset
}
func UpdateAOFAppliedOffset(offset uint64) {
	Metrics.AofAppliedOffset = offset
}

// for debug

func UpdateInQueueEntriesCount(count uint64) {
	Metrics.InQueueEntriesCount = count
}
func UpdateUnansweredBytesCount(count uint64) {
	Metrics.UnansweredBytesCount = count
}