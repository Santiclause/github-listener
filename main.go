package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type PullRequestEvent struct {
	Action string
	Number int
}

var (
	config = struct {
		SignatureKey []byte
	}{
		[]byte("ayylmao"),
	}
)

func main() {
	http.HandleFunc("/webhook", HandleWebhook)
	http.ListenAndServe(":11111", nil)
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	defer HandleRecover(w, r)
	if r.Method != "POST" {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	if event := r.Header.Get("X-GitHub-Event"); event != "pull_request" {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	signature := []byte(strings.TrimPrefix(r.Header.Get("X-Hub-Signature"), "sha1="))
	body, ok := CompareSignature(r.Body, signature, config.SignatureKey)
	if !ok {
		log.Printf("Unauthorized HTTP request: %s", signature)
		http.Error(w, "Unauthorized request", http.StatusForbidden)
		return
	}
	decoder := json.NewDecoder(body)
	var event PullRequestEvent
	err := decoder.Decode(&event)
	if err != nil {
		log.Printf("Error while handling webhook: %s", err)
		http.Error(w, "JSON decode failed", http.StatusBadRequest)
		return
	}
	test, _ := json.Marshal(event)
	log.Printf("OK: %s", test)
	w.WriteHeader(http.StatusOK)
}

func HandleRecover(w http.ResponseWriter, r *http.Request) {
	err := recover()
	if nil != err {
		w.WriteHeader(http.StatusInternalServerError)
		stack := make([]byte, 1<<16)
		runtime.Stack(stack, false)
		log.Printf("HTTP error %s\n%s", err, stack)
	}
}

func CompareSignature(payload io.Reader, signature, key []byte) (io.Reader, bool) {
	mac := hmac.New(sha1.New, key)
	tr := io.TeeReader(payload, mac)
	expected := mac.Sum(nil)
	return tr, hmac.Equal(signature, expected)
}