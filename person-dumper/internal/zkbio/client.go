package zkbio

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"person-dumper/internal/domain"
)

type Client struct {
	baseURL  string
	token    string
	insecure bool
	http     *http.Client
	pageSize int
}

func New(baseURL, token string, insecure bool, pageSize int) *Client {
	if pageSize <= 0 {
		pageSize = 200
	}

	tlsCfg := &tls.Config{InsecureSkipVerify: insecure}

	tr := &http.Transport{
		TLSClientConfig:     tlsCfg,
		ForceAttemptHTTP2:   true,
		DisableCompression:  false,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	cli := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	slog.Info("zkbio http client", "tls_insecure", insecure)

	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		token:    token,
		insecure: insecure,
		http:     cli,
		pageSize: pageSize,
	}
}

type personDTO struct {
	PIN         string `json:"pin"`
	Name        string `json:"name"`
	LastName    string `json:"lastName"`
	DeptName    string `json:"deptName"`
	MobilePhone string `json:"mobilePhone"`
	DeptCode    string `json:"deptCode"`
}

type listResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Page     int         `json:"page"`
		Size     int         `json:"size"`
		Total    int         `json:"total"`
		Data     []personDTO `json:"data"`
		Offset   int         `json:"offset"`
		LastPage bool        `json:"lastPage"`
	} `json:"data"`
}

// ListUsers — постранично тянет всех пользователей по deptCodes и возвращает []domain.DeviceUser
func (c *Client) ListUsers(ctx context.Context, deptCodes []string) ([]domain.DeviceUser, error) {
	var out []domain.DeviceUser
	pageNo := 1

	for {
		u, _ := url.Parse(c.baseURL + "/api/v2/person/getPersonList")
		q := u.Query()
		q.Set("access_token", c.token)
		q.Set("pageNo", fmt.Sprintf("%d", pageNo))
		q.Set("pageSize", fmt.Sprintf("%d", c.pageSize))
		if len(deptCodes) > 0 {
			q.Set("deptCodes", strings.Join(deptCodes, ","))
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader([]byte(`{}`)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		slog.Info("zkbio req", "url", u.String(), "deptCodes", strings.Join(deptCodes, ","))

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if !utf8.Valid(raw) {
			raw = bytes.ToValidUTF8(raw, []byte{})
		}
		runes := []rune(string(raw))
		preview := string(runes[:min(len(runes), 200)])

		if resp.StatusCode == http.StatusOK {
			slog.Info("zkbio http", "status", resp.StatusCode, "preview", preview, "len", len(raw))
		} else {
			slog.Error("zkbio http", "status", resp.StatusCode, "preview", preview, "len", len(raw))
		}

		var lr listResp
		if err := json.Unmarshal(raw, &lr); err != nil {
			return nil, fmt.Errorf("zkbio decode: %w", err)
		}
		if lr.Code != 0 && lr.Code != 200 {
			return nil, fmt.Errorf("zkbio code=%d message=%s", lr.Code, lr.Message)
		}

		for _, p := range lr.Data.Data {
			full := strings.TrimSpace(strings.TrimSpace(p.Name) + " " + strings.TrimSpace(p.LastName))
			out = append(out, domain.DeviceUser{
				ExternalID: p.PIN,
				FullName:   full,
				Department: p.DeptName,
				DeptCode:   p.DeptCode,
			})
		}

		if lr.Data.LastPage || len(out) >= lr.Data.Total || len(lr.Data.Data) == 0 {
			break
		}
		pageNo++
	}

	return out, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
