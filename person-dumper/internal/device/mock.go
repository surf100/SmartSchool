package device

import (
	"fmt"
	"strconv"

	"person-dumper/internal/domain"
)

type Mock struct {
	N int 
}

func NewMock(n int) *Mock { 
	if n <= 0 {
		n = 15
	}
	return &Mock{N: n}
}

func (m *Mock) Ping(ip string, port int) error {

	return nil
}

func (m *Mock) GetUsers(ip string, port int) ([]domain.DeviceUser, error) {
	users := make([]domain.DeviceUser, 0, m.N)
	for i := 1; i <= m.N; i++ {
		id := strconv.Itoa(1000 + i)
		users = append(users, domain.DeviceUser{
			ExternalID: id,
			FullName:   fmt.Sprintf("User %s", id),
			CardNo:     fmt.Sprintf("C%s", id),
			Department: "Default",
			Role:       "user",
		})
	}
	return users, nil
}
