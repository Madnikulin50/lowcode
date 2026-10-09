package agent

import "context"

type Storage interface {
	EnsureModule(ctx context.Context) (uint64, error)
	FindDevice(ctx context.Context, moduleID uint64, d Device) (uint64, error)
	FindDeviceByIP(ctx context.Context, moduleID uint64, ip string) (uint64, error)
	CreateDevice(ctx context.Context, moduleID uint64, d Device) (uint64, error)
	UpdateDevice(ctx context.Context, moduleID, recordID uint64, d Device) error
	ListDevices(ctx context.Context, moduleID uint64) ([]Device, error)
	GetDevice(ctx context.Context, moduleID, recordID uint64) (*Device, error)
	DeleteDevice(ctx context.Context, moduleID, recordID uint64) error
}
