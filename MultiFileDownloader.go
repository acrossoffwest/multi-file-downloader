package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type LoadData struct {
	index    int
	Url      string `json:"url"`
	Filepath string `json:"filepath"`
	Status   string `json:"status"`
}
type Result struct {
	Results      []LoadData `json:"results"`
	Time         string     `json:"time"`
	ErrorMessage string     `json:"error_message"`
}

func main() {
	start := time.Now()
	args := os.Args[1:]

	if len(args) != 2 {
		result := Result{
			Results:      []LoadData{},
			Time:         "",
			ErrorMessage: "Please provide only one argument: [timeout in seconds] [app name] [Url]:[output file path];...;[Url-n]:[output file-n path]",
		}
		r, e := json.Marshal(result)
		if e != nil {
			panic(e)
		}

		fmt.Println(string(r))
		return
	}
	timeout, _ := strconv.Atoi(args[0])

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(timeout) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: netTransport,
	}

	loadDatas := parseLoadingData(args[1])
	loadDatasLen := len(loadDatas)
	loadDataChan := make(chan LoadData, loadDatasLen)
	for _, data := range loadDatas {
		go func(data LoadData, out chan LoadData, httpClient *http.Client) {
			err := DownloadFile(httpClient, data)
			if err != nil {
				data.Status = fmt.Sprintf("%v", err)
				out <- data
				return
			}
			data.Status = "success"
			out <- data
		}(data, loadDataChan, httpClient)
	}

	i := 0
	for data := range loadDataChan {
		i++
		loadDatas[data.index] = data
		if i >= loadDatasLen {
			close(loadDataChan)
		}
	}

	elapsed := time.Since(start)
	result := Result{
		Results: loadDatas,
		Time:    fmt.Sprintf("%v", elapsed),
	}
	j, e := json.Marshal(result)
	if e != nil {
		panic(e)
	}
	fmt.Println(string(j))
}

// Loader data string template: `[Url]:[output file path];...;[Url-n]:[output file-n path]`
func parseLoadingData(loadingDataString string) []LoadData {
	explode := strings.Split(loadingDataString, ";;;")

	result := make([]LoadData, len(explode))
	for i, item := range explode {
		explodeItem := strings.Split(item, ":::")
		result[i] = LoadData{
			index:    i,
			Url:      explodeItem[0],
			Filepath: explodeItem[1],
		}
	}

	return result
}

func DownloadFile(httpClient *http.Client, data LoadData) error {
	resp, err := httpClient.Get(data.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(data.Filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
