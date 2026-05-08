package device

import "person-dumper/internal/domain"

type Client interface {
	// Вернёт пользователей, считанных с устройства по IP:port
	ListUsers(ip string, port int) ([]domain.DeviceUser, error)
	
}
