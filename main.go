package main

import (
	"context"
	"encoding/json"
	"github.com/Incubeta/proj-ebay-streaming-data-cache-updater/cache"
	"github.com/apsystole/log"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

// DispatchData is the message sent by the dispatcher to trigger the task handler cloud functions
type DispatchData struct {
	ProjectId string `json:"project_id"`
	DatasetId string `json:"dataset_id"`
	TableId   string `json:"table_id"`
	Market    string `json:"market"`
}

func getStats(w http.ResponseWriter, _ *http.Request) {
	//ctx := context.Background()
	stats, err := cache.GetStats()
	if err != nil {
		log.Errorf("Could not obtain stats: %s", err.Error())
	}
	_, err = w.Write([]byte(stats))
	if err != nil {
		return
	}
}

func PingPong(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("Pong"))
	if err != nil {
		return
	}
}

func UpdateCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Warningf("Method not allowed: %s", r.Method)
		w.WriteHeader(405)
		_, err := w.Write([]byte("Method not allowed"))
		if err != nil {
			return
		}
		return
	}

	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	var dispatchData DispatchData
	if err := json.NewDecoder(r.Body).Decode(&dispatchData); err != nil {
		log.Errorf("Could not decode request, please check if the request body is correct: %s", err.Error())
		w.WriteHeader(400)
		w.Write([]byte("Could not decode request, please check if the request body is correct"))
		return
	}
	log.Infoj("Request received for cache update", dispatchData)
	err := cache.RefreshRedisCache(ctx, dispatchData.Market)
	if err != nil {
		log.Errorf("Could not refresh cache: %s", err.Error())
		w.WriteHeader(500)
		w.Write([]byte("Could not refresh cache"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Cache updated"))
	return
}

func main() {
	log.Info("Starting cache updater web server")
	if err := godotenv.Load(".env"); err != nil {
		log.Info("No .env file found, using environment variables")
	}

	http.HandleFunc("/refresh", UpdateCache)
	http.HandleFunc("/ping", PingPong)
	http.HandleFunc("/stats", getStats)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
