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
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff"
)

var (
	// Version (set by compiler) is the version of program
	Version = "undefined"
	// BuildTime (set by compiler) is the program build time in '+%Y-%m-%dT%H:%M:%SZ' format
	BuildTime = "undefined"
	// GitHash (set by compiler) is the git commit hash of source tree
	GitHash = "undefined"
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
	defaultUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\tBuildTime: %v\tGitHash: %s\n", Version, BuildTime, GitHash)
		defaultUsage()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return errors.New("provide an url from where you need to download files")
	}

	uri := args[0]
	tracks, err := getTracks(uri)
	if err != nil {
		return err
	}

	fileNameFormat := fmt.Sprintf("%%0%dd-%%s%%s", len(strconv.Itoa(len(tracks))))
	idx := uint64(0)
	for _, tr := range tracks {
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
		fileName := limitFileName(fmt.Sprintf(fileNameFormat, idx, caption, ext), 255)
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

func limitFileName(fileName string, limit int) string {
	if utf8.RuneCountInString(fileName) <= limit {
		return fileName
	}

	runes := []rune(fileName)
	idx := limit / 2
	runes[idx] = rune('…')
	copy(runes[idx+1:], runes[len(runes)+idx+1-limit:])

	return string(runes[:limit])
}

func getTracks(uri string) ([]track, error) {
	doc, err := newDocumentFromURL(uri)
	if err != nil {
		return nil, err
	}

	sel := doc.Find("script.wp-playlist-script")
	if len(sel.Nodes) == 0 {
		return nil, errors.New("playlist is not found on the page")
	}
	var sc wpPlaylistScript
	if err := json.Unmarshal([]byte(sel.Text()), &sc); err != nil {
		return nil, fmt.Errorf("JSON parsing error: %v", err)
	}

	return sc.Tracks, nil
}

type track struct {
	Src     string `json:"src"`
	Title   string `json:"title"`
	Caption string `json:"caption"`
}

type wpPlaylistScript struct {
	Type         string  `json:"type"`
	Tracklist    bool    `json:"tracklist"`
	Tracknumbers bool    `json:"tracknumbers"`
	Images       bool    `json:"images"`
	Artists      bool    `json:"artists"`
	Tracks       []track `json:"tracks"`
}

func newDocumentFromURL(url string) (*goquery.Document, error) {
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

var errHTTPStatusServiceUnavailable = errors.New("503 Service Temporarily Unavailable")

func downloadFile(filepath string, url string) error {
	const permanentOn = 10
	i := 0

	return backoff.Retry(func() error {
		i++
		if i > permanentOn {
			return backoff.Permanent(fmt.Errorf("failed to download file in %d tries", permanentOn))
		}

		err := downloadFileAux(filepath, url)
		if err == nil || err == errHTTPStatusServiceUnavailable {
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
		return errHTTPStatusServiceUnavailable
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
