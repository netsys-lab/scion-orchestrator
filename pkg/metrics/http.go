package metrics

import (
	"log"
	"net/http"
)

func RunStatusHTTPServer(url string) error {
	// Start the status server
	h := http.NewServeMux()
	h.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		// Return the status of the services
		w.Header().Set("Content-Type", "application/json")
		statusBytes, err := Status.Json()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(statusBytes)
	})

	listenUrl := ":8141"
	if url != "" {
		listenUrl = url
	}

	log.Println("[Metrics] Starting HTTP status server on", listenUrl)

	err := http.ListenAndServe(listenUrl, h)
	return err
}
