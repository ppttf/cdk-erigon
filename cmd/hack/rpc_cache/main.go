package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// Rate limiter - semaphore to limit concurrent upstream requests
var (
	maxConcurrent = 5
	semaphore     chan struct{}
	rateMu        sync.Mutex
	lastRequest   time.Time
	minInterval   = 50 * time.Millisecond // ~20 req/sec max
)

var db *bolt.DB

// HTTP client with timeout for upstream requests
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

const (
	bucketName   = "Cache"
	expiryBucket = "Expiry"
)

// methods we don't cache
var methodsToIgnore = map[string]struct{}{}

// methods we configure expiry for
var methodsToExpire = map[string]time.Duration{
	"eth_getBlockByNumber": 1 * time.Minute,
}

// params that trigger expiration
var paramsToExpire = map[string]struct{}{
	"latest":    {},
	"finalized": {},
}

func initDB(file string) {
	var err error
	db, err = bolt.Open(file, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(expiryBucket))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func fetchFromCache(key string) ([]byte, bool) {
	var data []byte
	var expiry []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		e := tx.Bucket([]byte(expiryBucket))
		data = b.Get([]byte(key))
		expiry = e.Get([]byte(key))
		return nil
	})
	if err != nil || data == nil {
		return nil, false
	}

	if expiry != nil {
		expiryTime, err := time.Parse(time.RFC3339, string(expiry))
		if err == nil && time.Now().After(expiryTime) {
			// Cache entry has expired
			return nil, false
		}
	}

	// Check if the cached response contains an error
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(data, &jsonResponse); err == nil {
		if _, hasError := jsonResponse["error"]; hasError {
			// Cache entry is an error, evict it
			evictFromCache(key)
			return nil, false
		}
	}

	return data, true
}

func evictFromCache(key string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		e := tx.Bucket([]byte(expiryBucket))
		if err := b.Delete([]byte(key)); err != nil {
			return err
		}
		return e.Delete([]byte(key))
	})
	if err != nil {
		log.Println("Failed to evict from cache:", err)
	}
}

func saveToCache(key string, response []byte, duration time.Duration) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		e := tx.Bucket([]byte(expiryBucket))
		err := b.Put([]byte(key), response)
		if err != nil {
			return err
		}
		// Only set expiry if duration is not zero (indicating that it should expire)
		if duration > 0 {
			expiryTime := time.Now().Add(duration).Format(time.RFC3339)
			return e.Put([]byte(key), []byte(expiryTime))
		}
		return nil
	})
	if err != nil {
		log.Println("Failed to save to cache:", err)
	}
}

func generateCacheKey(chainID string, body []byte) (string, error) {
	var request map[string]interface{}
	err := json.Unmarshal(body, &request)
	if err != nil {
		return "", err
	}
	delete(request, "id")
	modifiedBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", chainID, modifiedBody), nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	reqID := time.Now().UnixNano()
	log.Printf("[%d] Incoming request: %s %s", reqID, r.Method, r.URL.String())

	endpoint := r.URL.Query().Get("endpoint")
	chainID := r.URL.Query().Get("chainid")
	if endpoint == "" || chainID == "" {
		log.Printf("[%d] ERROR: Missing endpoint or chainid", reqID)
		http.Error(w, "Missing endpoint or chainid parameter", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[%d] ERROR: Failed to read body: %v", reqID, err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Printf("[%d] Request body (%d bytes): %s", reqID, len(body), string(body))

	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("[%d] ERROR: Invalid JSON: %v", reqID, err)
		http.Error(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	method, ok := request["method"].(string)
	if !ok {
		log.Printf("[%d] ERROR: Invalid method", reqID)
		http.Error(w, "Invalid JSON-RPC method", http.StatusBadRequest)
		return
	}

	log.Printf("[%d] Method: %s", reqID, method)

	cacheKey, err := generateCacheKey(chainID, body)
	if err != nil {
		log.Printf("[%d] ERROR: Failed to generate cache key: %v", reqID, err)
		http.Error(w, "Failed to generate cache key", http.StatusInternalServerError)
		return
	}

	if _, ignore := methodsToIgnore[method]; !ignore {
		if cachedResponse, found := fetchFromCache(cacheKey); found {
			log.Printf("[%d] CACHE HIT: %d bytes", reqID, len(cachedResponse))
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache-Status", "HIT")
			n, err := w.Write(cachedResponse)
			log.Printf("[%d] Wrote %d bytes, err: %v", reqID, n, err)
			return
		}
	}

	url, _ := url2.Parse(endpoint)
	log.Printf("[%d] CACHE MISS: fetching from %s", reqID, url.Host)

	// Rate limiting: acquire semaphore and respect minimum interval
	semaphore <- struct{}{}
	rateMu.Lock()
	elapsed := time.Since(lastRequest)
	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}
	lastRequest = time.Now()
	rateMu.Unlock()

	resp, err := httpClient.Post(endpoint, "application/json", bytes.NewBuffer(body))
	<-semaphore // release semaphore
	if err != nil {
		log.Printf("[%d] ERROR: Upstream fetch failed: %v", reqID, err)
		http.Error(w, "Failed to fetch from upstream", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Printf("[%d] Upstream status: %d", reqID, resp.StatusCode)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%d] ERROR: Failed to read upstream response: %v", reqID, err)
		http.Error(w, "Failed to read upstream response", http.StatusInternalServerError)
		return
	}

	log.Printf("[%d] Upstream response: %d bytes", reqID, len(responseBody))

	// Handle non-200 status codes - return JSON-RPC error
	if resp.StatusCode != http.StatusOK {
		reqIDVal, _ := request["id"]
		errorMsg := fmt.Sprintf("upstream returned status %d", resp.StatusCode)
		if resp.StatusCode == http.StatusTooManyRequests {
			errorMsg = "rate limited by upstream"
		}
		jsonRPCError := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      reqIDVal,
			"error": map[string]interface{}{
				"code":    -32603,
				"message": errorMsg,
			},
		}
		errorResponse, _ := json.Marshal(jsonRPCError)
		log.Printf("[%d] Returning JSON-RPC error for status %d", reqID, resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache-Status", "ERROR")
		w.Write(errorResponse)
		return
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &jsonResponse); err == nil {
		if _, hasError := jsonResponse["error"]; hasError {
			log.Printf("[%d] Upstream returned JSON-RPC error, not caching", reqID)
		} else {
			if _, ignore := methodsToIgnore[method]; !ignore {
				cacheDuration := time.Duration(0)
				if duration, found := methodsToExpire[method]; found {
					if method == "eth_getBlockByNumber" {
						params, ok := request["params"].([]interface{})
						if ok && len(params) > 0 {
							param, ok := params[0].(string)
							if ok {
								if _, shouldExpire := paramsToExpire[param]; shouldExpire {
									cacheDuration = duration
								}
							}
						}
					} else {
						cacheDuration = duration
					}
				}
				saveToCache(cacheKey, responseBody, cacheDuration)
				log.Printf("[%d] Cached response (expiry: %v)", reqID, cacheDuration)
			}
		}
	} else {
		log.Printf("[%d] Failed to parse upstream response as JSON: %v", reqID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache-Status", "MISS")
	n, err := w.Write(responseBody)
	log.Printf("[%d] Wrote %d bytes, err: %v", reqID, n, err)
}

func handleCacheLookup(w http.ResponseWriter, r *http.Request) {
	cacheKey := r.URL.Query().Get("key")
	if cacheKey == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	cachedResponse, found := fetchFromCache(cacheKey)
	if !found {
		http.Error(w, "Cache miss", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(cachedResponse)
}

func main() {
	fileFlag := flag.String("file", "cache.db", "file to read")
	rateFlag := flag.Int("rate", 20, "max requests per second to upstream")
	concurrentFlag := flag.Int("concurrent", 5, "max concurrent requests to upstream")
	flag.Parse()

	// Initialize rate limiter
	maxConcurrent = *concurrentFlag
	semaphore = make(chan struct{}, maxConcurrent)
	minInterval = time.Second / time.Duration(*rateFlag)
	log.Printf("Rate limit: %d req/sec, max %d concurrent", *rateFlag, maxConcurrent)

	initDB(*fileFlag)
	defer db.Close()

	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/lookup", handleCacheLookup)
	server := &http.Server{
		Addr:           ":6969",
		Handler:        nil,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Starting proxy server on port 6969")
	log.Fatal(server.ListenAndServe())
}
