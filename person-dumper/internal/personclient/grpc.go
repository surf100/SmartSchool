package personclient

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"person-dumper/internal/domain"
	personpb "person-dumper/third_party/personapi" // импорт на сгенерённый из proto пакет
)

type GRPC struct {
	cli     personpb.PersonServiceClient
	timeout time.Duration
}

func NewGRPC(conn *grpc.ClientConn, timeout time.Duration) *GRPC {
	return &GRPC{cli: personpb.NewPersonServiceClient(conn), timeout: timeout}
}

func splitName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func (c *GRPC) Upsert(ctx context.Context, t domain.Tenant, src string, u domain.DeviceUser) (bool, *domain.Person, error) {

	md := metadata.New(map[string]string{
		"x-tenant-bin":    t.BIN,
		"x-tenant-schema": t.Schema,
		"x-source":        src,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	name, last := splitName(u.FullName)

	req := &personpb.UpsertByPinRequest{
		Pin:       u.ExternalID,
		Name:      name,
		LastName:  last,
		DeptName:  u.Department,
		PhotoPath: "",
	}

	cctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.cli.UpsertByPin(cctx, req)
	if err != nil {
		return false, nil, err
	}

	var out *domain.Person
	if resp.GetChanged() {
		full := name
		if last != "" {
			full = name + " " + last
		}
		out = &domain.Person{
			ExternalID: req.GetPin(),
			FullName:   full,
		}
	}

	return resp.GetCreated(), out, nil
}
