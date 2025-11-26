package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iris_installer/mhttp"
	"net/http"
	"strings"
	"time"
)

type githubLatestInfo struct {
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string    `json:"name"`
		Digest             string    `json:"digest"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	Body string `json:"body"`
}

type LatestInfo struct {
	Name       string
	IrisAPKURL string
	Digest     string
}

func GetLatestInfo() (*LatestInfo, error) {
	httpClient := mhttp.CreateHTTPClient()

	resp, err := httpClient.Get("https://api.github.com/repos/dolidolih/iris/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("getLatestInfo error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch latest release info")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %w", err)
	}

	var v = githubLatestInfo{}
	json.Unmarshal(body, &v)

	for _, asset := range v.Assets {
		if asset.Name == "Iris.apk" {
			return &LatestInfo{
				Name:       v.Name,
				IrisAPKURL: asset.BrowserDownloadURL,
				Digest:     asset.Digest,
			}, nil
		}
	}

	return nil, fmt.Errorf("iris.apk를 찾을 수 없습니다")
}

func CheckDigest(data []byte, digest string) (bool, error) {
	if strings.HasPrefix(digest, "sha256:") {
		digest = strings.TrimPrefix(digest, "sha256:")
		dataDigest := fmt.Sprintf("%x", sha256.Sum256(data))
		return dataDigest == digest, nil
	} else {
		return false, fmt.Errorf("지원하지 않는 다이제스트 형식입니다: %s", digest)
	}
}
