package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	TargetURL      = "http://localhost:8080/jobs"
	WorkerPoolSize = 10
)

func main() {
	log.Println("CHAOS HARNESS RUNNING: Injecting heavy transactional load...")

	var wg sync.WaitGroup
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 1; i <= WorkerPoolSize; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 1; j <= 5; j++ {
				jobName := "video.transcode"
				if r.Float32() < 0.3 {
					jobName = "fail.me"
				}

				payload := map[string]interface{}{
					"name": jobName,
					"payload": map[string]interface{}{
						"load_id":   fmt.Sprintf("w%d_j%d", workerID, j),
						"timestamp": time.Now().Unix(),
					},
				}

				bodyBytes, _ := json.Marshal(payload)

				resp, err := http.Post(TargetURL, "application/json", bytes.NewBuffer(bodyBytes))
				if err != nil {
					log.Printf("[WORKER %d] Ingestion gateway unreachable: %v", workerID, err)
					time.Sleep(500 * time.Millisecond)
					continue
				}

				var respMap map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&respMap)
				resp.Body.Close()

				log.Printf("[WORKER %d] Dispatched task target: %s -> Response HTTP Code %d. Track ID: %v",
					workerID, jobName, resp.StatusCode, respMap["job_id"])

				time.Sleep(time.Duration(1500+r.Intn(2000)) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	log.Println("CHAOS SEED COMPLETE: All concurrent traffic payloads injected.")
}
