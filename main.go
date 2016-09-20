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

func main() {
	vcs := flag.String("vcs", "bitbucket", "CircleCI VCS Type")
	user := flag.String("user", "systemlogic", "CircleCI Organization Name")
	project := flag.String("project", "produseal-www", "CircleCI Organization Name")
	build := flag.String("build", "latest", "Project Build Number (or 'latest')")
	flag.Parse()

	circleToken := os.Getenv("CIRCLE_CI_TOKEN")

	filesToDownload := make(chan ArtifactJson, 4)
	go downloadArtifacts(circleToken, filesToDownload)

	apiClient := sling.New().Base("https://circleci.com/api/v1.1/")
	apiClient.Set("Accept", "application/json")

	result := []ArtifactJson{}
	q := fmt.Sprintf("project/%s/%s/%s/%s/artifacts?circle-token=%s", *vcs, *user, *project, *build, circleToken)
	s := apiClient.Get(q)
	resp, err := s.ReceiveSuccess(&result)
	fmt.Println(resp.Status)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, artifact := range result {
			fmt.Println(artifact.Url)
			wg.Add(1)
			filesToDownload <- artifact
		}
	}

	wg.Wait()
}
func downloadArtifacts(circleToken string, artifacts chan ArtifactJson) {
	for file := range artifacts {
		t := time.Now()

		out, _ := os.Create(strings.Replace(file.PrettyPath, "$CIRCLE_ARTIFACTS", ".", 1))
		defer out.Close()

		q := fmt.Sprintf("%s?circle-token=%s", file.Url, circleToken)

		resp, _ := http.DefaultClient.Get(q)
		defer resp.Body.Close()

		n, _ := io.Copy(out, resp.Body)
		fmt.Println("file", file.PrettyPath, time.Since(t), n, "bytes")

		wg.Done()
	}
}

type ArtifactJson struct {
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	NodeIndex  string `json:"nodex_index"`
	Url        string `json:"url"`
}
