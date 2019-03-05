package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff"
)

func init() {
	// otherwise on some files get an error "stream error: stream ID 1; PROTOCOL_ERROR"
	_ = os.Setenv("GODEBUG", "http2client=0")
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dirPath        string
		skipFilesCount uint64
	)
	flag.StringVar(&dirPath, "dir", ".", "directory name where to save the files")
	flag.Uint64Var(&skipFilesCount, "skip", 0, "number of files to skip")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return errors.New("provide an url from where you need to download files")
	}

	uri := args[0]
	doc, err := newDocumentFromUrl(uri)
	if err != nil {
		return err
	}

	sel := doc.Find("script.wp-playlist-script")
	if len(sel.Nodes) == 0 {
		return errors.New("playlist is not found on the page")
	}
	var sc wpPlaylistScript
	if err := json.Unmarshal([]byte(sel.Text()), &sc); err != nil {
		return fmt.Errorf("JSON parsing error: %v", err)
	}

	fileNameFormat := fmt.Sprintf("%%0%dd-%%s%%s", len(strconv.Itoa(len(sc.Tracks))))
	idx := uint64(0)
	for _, tr := range sc.Tracks {
		idx++

		uri := tr.Src
		trackCaption := tr.Caption
		if trackCaption == "" {
			trackCaption = tr.Title
		}
		caption, err := htmlText(trackCaption)
		if err != nil {
			return err
		}

		uriParsed, err := url.Parse(uri)
		if err != nil {
			return err
		}
		ext := filepath.Ext(path.Base(uriParsed.Path))
		fileName := fmt.Sprintf(fileNameFormat, idx, caption, ext)
		filePath := path.Join(dirPath, fileName)

		if idx > skipFilesCount {
			fmt.Printf("Downloading %s to %s...\n", uri, filePath)
			if err := downloadFile(filePath, uri); err != nil {
				return err
			}
		} else {
			fmt.Printf("Skipping %s to %s...\n", uri, filePath)
		}
	}

	fmt.Println("Done.")
	return nil
}

type wpPlaylistScript struct {
	Type         string `json:"type"`
	Tracklist    bool   `json:"tracklist"`
	Tracknumbers bool   `json:"tracknumbers"`
	Images       bool   `json:"images"`
	Artists      bool   `json:"artists"`
	Tracks       []struct {
		Src         string `json:"src"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		Caption     string `json:"caption"`
		Description string `json:"description"`
		Meta        struct {
			LengthFormatted string `json:"length_formatted"`
			Artist          string `json:"artist"`
			Album           string `json:"album"`
		} `json:"meta"`
	} `json:"tracks"`
}

func newDocumentFromUrl(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %v", err)
	}

	return doc, nil
}

var httpStatusServiceUnavailableErr = errors.New("503 Service Temporarily Unavailable")

func downloadFile(filepath string, url string) error {
	const permanentOn = 10
	i := 0

	return backoff.Retry(func() error {
		i++
		if i > permanentOn {
			return backoff.Permanent(fmt.Errorf("failed to download file in %d tries", permanentOn))
		}

		err := downloadFileAux(filepath, url)
		if err == nil || err == httpStatusServiceUnavailableErr {
			return err
		}
		return backoff.Permanent(err)
	}, backoff.NewExponentialBackOff())
}

func downloadFileAux(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable:
		return httpStatusServiceUnavailableErr
	default:
		return fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func htmlText(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("htmlText: %v", err)
	}
	return doc.Text(), nil
}
