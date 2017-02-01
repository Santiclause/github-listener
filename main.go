package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
		JenkinsKey   string
		JenkinsJob   string
		JenkinsUrl   string
		Debug        bool
	}{
		[]byte("ayylmao"),
		"huehuehue",
		"testing",
		"http://jenkins-ui:8080",
		false,
	}
)

func init() {
	if v := os.Getenv("SIGNATURE_KEY"); v != "" {
		config.SignatureKey = []byte(v)
	}
	if v := os.Getenv("JENKINS_KEY"); v != "" {
		config.JenkinsKey = v
	}
	if v := os.Getenv("JENKINS_JOB"); v != "" {
		config.JenkinsJob = v
	}
	if v := os.Getenv("JENKINS_URL"); v != "" {
		config.JenkinsUrl = v
	}
	flag.BoolVar(&config.Debug, "v", config.Debug, "Debug mode")
}

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
	// This is _only_ for pull_request events for now
	if event := r.Header.Get("X-GitHub-Event"); event != "pull_request" {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	signature := []byte(strings.TrimPrefix(r.Header.Get("X-Hub-Signature"), "sha1="))
	var event PullRequestEvent
	err := json.Unmarshal(body, &event)
	if err != nil {
		log.Printf("Error while decoding webhook: %s", err)
		http.Error(w, "JSON decode failed", http.StatusBadRequest)
		return
	}
	if !CompareSignatures(body, signature, config.SignatureKey) {
		http.Error(w, "Unauthorized request", http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	if config.Debug {
		log.Printf("Received event %s on PR %d", event.Action, event.Number)
	}
	if event.Action != "closed" {
		return
	}
	resp, err := http.Get(fmt.Sprintf("%s/buildByToken/buildWithParameters?job=%s&token=%s&PULL_REQUEST_ID=%d", config.JenkinsUrl, config.JenkinsJob, config.JenkinsKey, event.Number))
	if err != nil {
		log.Printf("error attempting to send GET request to Jenkins")
	} else {
		if config.Debug {
			content, _ := ioutil.ReadAll(resp.Body)
			log.Printf("GET %d\n%s", resp.StatusCode, content)
		}
		resp.Body.Close()
	}
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

func CompareSignatures(payload, signature, key []byte) bool {
	mac := hmac.New(sha1.New, config.SignatureKey)
	mac.Write(payload)
	expected := mac.Sum(nil)
	actual := make([]byte, 20)
	hex.Decode(actual, signature)
	return hmac.Equal(actual, expected)
}
