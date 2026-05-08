package zkteco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"person-dumper/internal/domain"
)

var _ domain.DeviceClient = (*Client)(nil)

type Client struct{ timeout time.Duration }

func New() *Client                           { return &Client{timeout: 5 * time.Second} }
func NewWithTimeout(d time.Duration) *Client { return &Client{timeout: d} }


// cleanHost: принимает "127.0.0.1/32", "[::1]:4370", "tcp://host:port" и т.п. → возвращает чистый host
func cleanHost(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	} 
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	if strings.Contains(s, "/") {
		if ip, _, err := net.ParseCIDR(s); err == nil {
			s = ip.String()
		} else {
			return "", fmt.Errorf("invalid CIDR: %w", err)
		}
	}
	s = strings.Trim(s, "[]")  
	if net.ParseIP(s) == nil { 
		if _, err := net.LookupHost(s); err != nil {
			return "", fmt.Errorf("invalid host/IP: %q", raw)
		}
	}
	return s, nil
}

type fileUser struct {
	Pin        string `json:"pin"`         
	ExternalID string `json:"external_id"` 
	FullName   string `json:"full_name"`   
	Name       string `json:"name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
}

func loadUsersFromFile(path string) ([]domain.DeviceUser, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
                                 
	var arr []fileUser
	if err := json.Unmarshal(b, &arr); err != nil {
		return nil, err
	}

	out := make([]domain.DeviceUser, 0, len(arr))
	for _, u := range arr {
		pin := strings.TrimSpace(u.Pin)
		if pin == "" {
			pin = strings.TrimSpace(u.ExternalID)
		}
		if pin == "" {
			continue
		}                                  

		full := strings.TrimSpace(u.FullName)
		if full == "" {
			full = strings.TrimSpace(strings.TrimSpace(u.Name) + " " + strings.TrimSpace(u.LastName))
		}
		out = append(out, domain.DeviceUser{
			ExternalID: pin,
			FullName:   full,
			Department: strings.TrimSpace(u.Department),
		})
	}
	return out, nil
}

                                 


func (c *Client) Ping(ip string, port int) error {
	if os.Getenv("ZK_DEMO") == "1" || os.Getenv("ZK_USERS_FILE") != "" {
		return nil
	}
	host, err := cleanHost(ip)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}


func (c *Client) GetUsers(ip string, port int) ([]domain.DeviceUser, error) {
	if os.Getenv("ZK_DEMO") == "1" {
		const n = 15
		users := make([]domain.DeviceUser, 0, n)
		for i := 1; i <= n; i++ {
			id := 1000 + i
			users = append(users, domain.DeviceUser{
				ExternalID: strconv.Itoa(id),
				FullName:   "Demo " + strconv.Itoa(i) + " User",
				Department: "DemoDept",
			})
		}
		return users, nil
	}
	if file := os.Getenv("ZK_USERS_FILE"); file != "" {
		return loadUsersFromFile(file)
	}
	return nil, errors.New("zkteco: real device GetUsers not implemented yet")
}
