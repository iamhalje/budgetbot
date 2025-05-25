package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetStargazers(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/stargazers", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []struct {
		Login string `json:"login"`
	}

	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return nil, err
	}

	var logins []string
	for _, u := range users {
		logins = append(logins, u.Login)
	}

	return logins, nil
}

func IsUserStargazer(username string, stargazers []string) bool {
	for _, u := range stargazers {
		if strings.EqualFold(u, username) {
			return true
		}
	}

	return false
}
