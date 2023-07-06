package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func loadData(appName string) (string, error) {
	fileName := "data/" + appName + ".txt"
	body, err := os.ReadFile(fileName)
	if err != nil {
		return "error", err
	}
	return string(body), nil
}

func saveData(appName string, data string) bool {
	fileName := "data/" + appName + ".txt"
	err := os.WriteFile(fileName, []byte(data), 0666)
	if err != nil {
		return false
	}
	return true
}

func readImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ioutil.Readall -> %v", err)
	}
	res.Body.Close()
	return data, nil
}

func writeResponse(responseCode int, w http.ResponseWriter) {
	var imgData = []byte{}
	var err error
	switch responseCode {
	case http.StatusAccepted:
		imgData, err = readImage("https://httpcats.com/202.jpg")
	case http.StatusBadRequest:
		imgData, err = readImage("https://httpcats.com/400.jpg")
	case http.StatusUnauthorized:
		imgData, err = readImage("https://httpcats.com/401.jpg")
	case http.StatusInternalServerError:
		imgData, err = readImage("https://httpcats.com/500.jpg")
	case http.StatusServiceUnavailable:
		imgData, err = readImage("https://httpcats.com/503.jpg")
	default:
		imgData, err = readImage("https://httpcats.com/500.jpg")
	}
	w.Header().Add("Cache-Control", "no-store")
	w.WriteHeader(responseCode)
	if err == nil {
		w.Header().Set("Content-Type", "image/jpg")
		w.Write(imgData)
	}
}

func main() {
	http.HandleFunc("/update", viewHandler)
	log.Fatal(http.ListenAndServe("127.0.0.1:12345", nil))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		appName := query.Get("name")
		data, err := loadData(appName)
		if err != nil {
			writeResponse(http.StatusBadRequest, w)
		} else {
			w.Header().Add("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(data))
		}
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			writeResponse(http.StatusBadRequest, w)
			return
		}
		v := r.Form
		appName := v.Get("app_name")
		appVersion := v.Get("app_version")
		downloadUrl := v.Get("download_url")
		password := v.Get("password")
		if password != "qw12QW!@" {
			writeResponse(http.StatusUnauthorized, w)
			return
		}
		data := map[string]interface{}{
			"version": appVersion,
			"url":     downloadUrl,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			writeResponse(http.StatusInternalServerError, w)
			return
		}
		result := saveData(appName, string(jsonData))

		if result {
			writeResponse(http.StatusAccepted, w)
			return
		}

		writeResponse(http.StatusServiceUnavailable, w)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
