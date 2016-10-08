package main

import (
	"flag"
	"fmt"
	"github.com/dghubble/sling"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Artifact struct {
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	NodeIndex  string `json:"nodex_index"`
	Url        string `json:"url"`
}

func main() {
	vcs := flag.String("vcs", "", "CircleCI VCS Type")
	user := flag.String("user", "", "CircleCI Organization Name")
	project := flag.String("project", "", "CircleCI Organization Name")
	build := flag.String("build", "latest", "Project Build Number (or 'latest')")
	files := flag.Uint("files", 4, "Maximum number of concurrent downloads")
	downloadFolder := flag.String("target", ".", "Target download folder")
	flag.Parse()

	if *vcs == "" || *user == "" || *project == "" || *build == "" {
		flag.Usage()
		os.Exit(1)
	}

	circleToken := os.Getenv("CIRCLE_CI_TOKEN")

	if circleToken == "" {
		fmt.Print("CIRCLE_CI_TOKEN environment variable must be set")
		os.Exit(2)
	}

	filesToDownload := make(chan Artifact, int(*files))
	go downloadArtifacts(circleToken, *downloadFolder, filesToDownload)

	apiClient := sling.New().Base("https://circleci.com/api/v1.1/")
	apiClient.Set("Accept", "application/json")

	result := []Artifact{}
	q := fmt.Sprintf("project/%s/%s/%s/%s/artifacts?circle-token=%s", *vcs, *user, *project, *build, circleToken)
	s := apiClient.Get(q)
	resp, err := s.ReceiveSuccess(&result)
	fmt.Println(resp.Status)

	fmt.Println("starting download of", len(result), "files...")

	if err != nil {
		fmt.Println(err)
	} else {
		for _, artifact := range result {
			fmt.Println(artifact.Url)
			wg.Add(1)
			filesToDownload <- artifact
		}
	}

	close(filesToDownload)

	wg.Wait()

	fmt.Println("...operation completed")
}

func downloadArtifacts(circleToken string, downloadFolder string, artifacts chan Artifact) {
	for file := range artifacts {
		go func() {
			t := time.Now()

			out, err := os.Create(strings.Replace(file.PrettyPath, "$CIRCLE_ARTIFACTS", downloadFolder, 1))
			if err != nil {
				fmt.Println("download of ", file.PrettyPath, "into", downloadFolder, "failed: ", err.Error())
			} else {
				defer out.Close()
			}

			q := fmt.Sprintf("%s?circle-token=%s", file.Url, circleToken)

			resp, err := http.DefaultClient.Get(q)
			if err != nil {
				fmt.Println("download of ", file.PrettyPath, "into", downloadFolder, "failed: ", err.Error())
			} else {
				defer resp.Body.Close()
			}

			n, err := io.Copy(out, resp.Body)
			if err != nil {
				fmt.Println("download of ", file.PrettyPath, "into", downloadFolder, "failed: ", err.Error())
			} else {
				fmt.Println("file", file.PrettyPath, time.Since(t), n, "bytes")
			}

			wg.Done()
		}()
	}
}
