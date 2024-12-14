package ddns

import (
	"log"
	"net/http"
	"os"
)

func Invoke(w http.ResponseWriter, r *http.Request) {
	createProvider()
	createCloudProvider()
	initialize()
	zoneName := os.Getenv("ZONE_NAME")
	recordName := os.Getenv("RECORD_NAME")

	rec, err := update(zoneName, recordName)
	if err != nil {
		log.Printf("DNS update failed: %v", err)
		http.Error(w, "DNS update failed", http.StatusInternalServerError)
		return
	}

	recordChecker(rec)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("DNS updated successfully"))
}

func startServerless() {
	log.Println("running in serverless mode")
	rec, err := update(zoneName, recordName)
	if err != nil {
		log.Println(err)
	}
	recordChecker(rec)
	http.HandleFunc("/", Invoke)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
