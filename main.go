package main

import (
	"io"
	"math/rand"
	"net/http"
	"sync"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLShorter struct {
	urls map[string]string
	mu   sync.RWMutex
}

func main() {
	s := NewURLShorter()
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HandleRequest)
	err := http.ListenAndServe(":8082", mux)
	if err != nil {
		panic(err)
	}
}

func NewURLShorter() *URLShorter {
	return &URLShorter{
		urls: make(map[string]string),
	}
}

func (s *URLShorter) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.HandlePOST(w, r)
	} else if r.Method == http.MethodGet {
		s.HandleGET(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *URLShorter) HandlePOST(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	longURL := string(body)
	id := s.GenerateShortID(longURL)
	s.StoreURL(id, longURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8082/" + id))
}

func (s *URLShorter) HandleGET(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	longURL := s.GetLongURL(id)
	if longURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *URLShorter) GetLongURL(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[id]
}

func (s *URLShorter) GenerateShortID(longURL string) string {
	shortURL := make([]byte, 8)
	for i := 0; i < 8; i++ {
		shortURL[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortURL)
}

func (s *URLShorter) StoreURL(id string, longURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[id] = longURL
}
