package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	AccessLevelGeneralID = "402888479817e873019817ed95ca0471"
	AccessLevelSUSNID    = "8a7c81979851e1dc0198562805e4003f"
	accessToken          = "6E3B736F74072CE24169304C5582F0A57F49F646B0595BDA737B050EFED75C8A"
)

func AssignAccessLevelsToPerson(ctx context.Context, pin string, susn bool) error {
    baseURL := "https://10.252.1.23:8098/api/accLevel/syncPerson"

    levelIds := AccessLevelGeneralID
    if susn {
        levelIds += "," + AccessLevelSUSNID
    }

    url := fmt.Sprintf("%s?levelIds=%s&pin=%s&access_token=%s", baseURL, levelIds, pin, accessToken)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Accept", "application/json")

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: transport}

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    log.Printf("Ответ от ZKTeco [%d]: %s", resp.StatusCode, string(body))

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("ошибка syncPerson: %s", string(body))
    }

    log.Printf("AccessLevels назначены пользователю PIN=%s (susn=%v)", pin, susn)
    return nil
}
