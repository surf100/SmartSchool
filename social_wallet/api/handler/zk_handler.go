package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "social-wallet/api/proto/gen"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ZKEventRequest struct {
	Pin       string `json:"pin"`
	DevSn     string `json:"devSn"`
	Reader    string `json:"readerName"`
	DoorName  string `json:"doorName"`
	Timestamp string `json:"timestamp"`
}
type ZKContent struct {
	Pin        string `json:"pin"`
	DoorName   string `json:"doorName"`
	ReaderName string `json:"readerName"`
	EventTime  int64  `json:"eventTime"`
	DevSn      string `json:"devSn"`
}

func (h *TerminalHandler) HandleZKBioAttendance(ctx context.Context, req *pb.ZKEventRequest) (*emptypb.Empty, error) {
	var content ZKContent
	if err := json.Unmarshal([]byte(req.Content), &content); err != nil {
		log.Printf("Ошибка парсинга content: %v", err)
		return &emptypb.Empty{}, nil
	}

	log.Printf("ZKBio Push Event: PIN=%s, Door=%s, Reader=%s", content.Pin, content.DoorName, content.ReaderName)

	person, err := h.repo.GetPersonByPin(ctx, content.Pin)
	if err != nil || person == nil {
		log.Printf("Пользователь не найден: %s", content.Pin)
		return &emptypb.Empty{}, nil
	}

	if !person.Susn {
		log.Printf("Пользователь без SUSN: %s", content.Pin)
		return &emptypb.Empty{}, nil
	}

	eventTime := time.UnixMilli(content.EventTime).UTC()

	err = sendToSocialWallet(person.IIN, person.SchoolBin, eventTime)
	if err != nil {
		log.Printf("Ошибка отправки в Social Wallet: %v", err)
	}

	return &emptypb.Empty{}, nil
}


func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var swHTTP = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: func() time.Duration {
		if s := os.Getenv("SW_TIMEOUT"); s != "" {
			if d, err := time.ParseDuration(s); err == nil {
				return d
			}
		}
		return 10 * time.Second
	}(),
}


func sendToSocialWallet(iin, schoolBin string, eventTime time.Time) error {
	endpoint := os.Getenv("SW_ENDPOINT")
	user := os.Getenv("SW_USER")
	pass := os.Getenv("SW_PASS")
	timeoutStr := os.Getenv("SW_TIMEOUT")

	if endpoint == "" || user == "" || pass == "" {
		return fmt.Errorf("missing SW_USER/SW_PASS env")
	}

	timeout := 10 * time.Second
	if timeoutStr != "" {
		if t, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = t
		}
	}

	// 🔹 Приводим время к UTC и форматируем под ТЗ
	utcDate := eventTime.UTC().Format("2006-01-02 15:04:05")

	payload := map[string]string{
		"iin":        iin,
		"date":       utcDate,
		"school_bin": schoolBin,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("SW response: %s", body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
