package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/service"
)

type TopologyHandler struct {
	topologyService *service.TopologyService
}

func NewTopologyHandler(topologyService *service.TopologyService) *TopologyHandler {
	return &TopologyHandler{
		topologyService: topologyService,
	}
}

func (h *TopologyHandler) Register(api huma.API) {
	// デバイス関連
	huma.Register(api, huma.Operation{
		OperationID: "add-device",
		Method:      http.MethodPost,
		Path:        "/api/devices",
		Summary:     "Add a new device",
		Tags:        []string{"devices"},
	}, h.AddDevice)

	huma.Register(api, huma.Operation{
		OperationID: "get-device",
		Method:      http.MethodGet,
		Path:        "/api/devices/{deviceId}",
		Summary:     "Get device information",
		Tags:        []string{"devices"},
	}, h.GetDevice)

	huma.Register(api, huma.Operation{
		OperationID: "get-device-with-neighbors",
		Method:      http.MethodGet,
		Path:        "/api/devices/{deviceId}/neighbors",
		Summary:     "Get device with neighbors",
		Tags:        []string{"devices"},
	}, h.GetDeviceWithNeighbors)

	huma.Register(api, huma.Operation{
		OperationID: "update-device",
		Method:      http.MethodPut,
		Path:        "/api/devices/{deviceId}",
		Summary:     "Update device",
		Tags:        []string{"devices"},
	}, h.UpdateDevice)

	huma.Register(api, huma.Operation{
		OperationID: "delete-device",
		Method:      http.MethodDelete,
		Path:        "/api/devices/{deviceId}",
		Summary:     "Delete device",
		Tags:        []string{"devices"},
	}, h.DeleteDevice)

	huma.Register(api, huma.Operation{
		OperationID: "list-devices",
		Method:      http.MethodGet,
		Path:        "/api/devices",
		Summary:     "List devices with pagination and filtering",
		Description: "Get paginated list of devices with optional filtering by type, hardware, or instance",
		Tags:        []string{"devices"},
	}, h.SearchDevices)

	huma.Register(api, huma.Operation{
		OperationID: "bulk-add-devices",
		Method:      http.MethodPost,
		Path:        "/api/devices/bulk",
		Summary:     "Bulk add devices",
		Tags:        []string{"devices"},
	}, h.BulkAddDevices)

	// リンク関連
	huma.Register(api, huma.Operation{
		OperationID: "add-link",
		Method:      http.MethodPost,
		Path:        "/api/links",
		Summary:     "Add a new link",
		Tags:        []string{"links"},
	}, h.AddLink)

	huma.Register(api, huma.Operation{
		OperationID: "get-link",
		Method:      http.MethodGet,
		Path:        "/api/links/{linkId}",
		Summary:     "Get link information",
		Tags:        []string{"links"},
	}, h.GetLink)

	huma.Register(api, huma.Operation{
		OperationID: "update-link",
		Method:      http.MethodPut,
		Path:        "/api/links/{linkId}",
		Summary:     "Update link",
		Tags:        []string{"links"},
	}, h.UpdateLink)

	huma.Register(api, huma.Operation{
		OperationID: "delete-link",
		Method:      http.MethodDelete,
		Path:        "/api/links/{linkId}",
		Summary:     "Delete link",
		Tags:        []string{"links"},
	}, h.DeleteLink)

	huma.Register(api, huma.Operation{
		OperationID: "bulk-add-links",
		Method:      http.MethodPost,
		Path:        "/api/links/bulk",
		Summary:     "Bulk add links",
		Tags:        []string{"links"},
	}, h.BulkAddLinks)
}

// デバイス関連ハンドラー
func (h *TopologyHandler) AddDevice(ctx context.Context, input *struct {
	Body topology.Device
}) (*struct {
	Body topology.Device
}, error) {
	device := input.Body
	device.LastSeen = time.Now()

	if device.Status == "" {
		device.Status = "unknown"
	}
	if device.Metadata == nil {
		device.Metadata = make(map[string]string)
	}

	if err := h.topologyService.AddDevice(ctx, device); err != nil {
		return nil, huma.Error400BadRequest("Failed to add device", err)
	}

	addedDevice, err := h.topologyService.GetDevice(ctx, device.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve added device", err)
	}

	return &struct {
		Body topology.Device
	}{
		Body: *addedDevice,
	}, nil
}

func (h *TopologyHandler) GetDevice(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
}) (*struct {
	Body topology.Device
}, error) {
	device, err := h.topologyService.GetDevice(ctx, input.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get device", err)
	}
	if device == nil {
		return nil, huma.Error404NotFound("Device not found")
	}

	return &struct {
		Body topology.Device
	}{
		Body: *device,
	}, nil
}

func (h *TopologyHandler) GetDeviceWithNeighbors(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
}) (*struct {
	Body service.DeviceWithNeighbors
}, error) {
	deviceWithNeighbors, err := h.topologyService.GetDeviceWithNeighbors(ctx, input.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get device with neighbors", err)
	}

	return &struct {
		Body service.DeviceWithNeighbors
	}{
		Body: *deviceWithNeighbors,
	}, nil
}

func (h *TopologyHandler) UpdateDevice(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
	Body     topology.Device
}) (*struct {
	Body topology.Device
}, error) {
	existingDevice, err := h.topologyService.GetDevice(ctx, input.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get device", err)
	}
	if existingDevice == nil {
		return nil, huma.Error404NotFound("Device not found")
	}

	updatedDevice := input.Body
	updatedDevice.ID = input.DeviceID

	if err := h.topologyService.UpdateDevice(ctx, updatedDevice); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update device", err)
	}

	device, err := h.topologyService.GetDevice(ctx, input.DeviceID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve updated device", err)
	}

	return &struct {
		Body topology.Device
	}{
		Body: *device,
	}, nil
}

func (h *TopologyHandler) DeleteDevice(ctx context.Context, input *struct {
	DeviceID string `path:"deviceId"`
}) (*struct{}, error) {
	if err := h.topologyService.RemoveDevice(ctx, input.DeviceID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete device", err)
	}

	return &struct{}{}, nil
}

type DeviceListResponse struct {
	Devices    []topology.Device           `json:"devices"`
	Pagination *topology.PaginationResult `json:"pagination"`
}

func (h *TopologyHandler) SearchDevices(ctx context.Context, input *struct {
	Page     int    `query:"page" default:"1" minimum:"1"`
	PageSize int    `query:"page_size" default:"20" minimum:"1" maximum:"100"`
	OrderBy  string `query:"order_by" default:"name"`
	SortDir  string `query:"sort_dir" default:"asc"`
	Type     string `query:"type"`
	Hardware string `query:"hardware"`
	Instance string `query:"instance"`
}) (*struct {
	Body DeviceListResponse
}, error) {
	opts := topology.PaginationOptions{
		Page:     input.Page,
		PageSize: input.PageSize,
		OrderBy:  input.OrderBy,
		SortDir:  input.SortDir,
		Type:     input.Type,
		Hardware: input.Hardware,
		Instance: input.Instance,
	}

	devices, pagination, err := h.topologyService.GetDevices(ctx, opts)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get devices", err)
	}

	return &struct {
		Body DeviceListResponse
	}{
		Body: DeviceListResponse{
			Devices:    devices,
			Pagination: pagination,
		},
	}, nil
}

func (h *TopologyHandler) BulkAddDevices(ctx context.Context, input *struct {
	Body struct {
		Devices []topology.Device `json:"devices"`
	}
}) (*struct {
	Body []topology.Device
}, error) {
	devices := input.Body.Devices
	for i := range devices {
		devices[i].LastSeen = time.Now()
		if devices[i].Status == "" {
			devices[i].Status = "unknown"
		}
		if devices[i].Metadata == nil {
			devices[i].Metadata = make(map[string]string)
		}
	}

	if err := h.topologyService.BulkAddDevices(ctx, devices); err != nil {
		return nil, huma.Error400BadRequest("Failed to bulk add devices", err)
	}

	return &struct {
		Body []topology.Device
	}{
		Body: devices,
	}, nil
}

// リンク関連ハンドラー
func (h *TopologyHandler) AddLink(ctx context.Context, input *struct {
	Body topology.Link
}) (*struct {
	Body topology.Link
}, error) {
	link := input.Body
	link.LastSeen = time.Now()

	if link.Weight == 0 {
		link.Weight = 1.0
	}
	if link.Status == "" {
		link.Status = "unknown"
	}
	if link.Metadata == nil {
		link.Metadata = make(map[string]string)
	}

	if err := h.topologyService.AddLink(ctx, link); err != nil {
		return nil, huma.Error400BadRequest("Failed to add link", err)
	}

	addedLink, err := h.topologyService.GetLink(ctx, link.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve added link", err)
	}

	return &struct {
		Body topology.Link
	}{
		Body: *addedLink,
	}, nil
}

func (h *TopologyHandler) GetLink(ctx context.Context, input *struct {
	LinkID string `path:"linkId"`
}) (*struct {
	Body topology.Link
}, error) {
	link, err := h.topologyService.GetLink(ctx, input.LinkID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get link", err)
	}
	if link == nil {
		return nil, huma.Error404NotFound("Link not found")
	}

	return &struct {
		Body topology.Link
	}{
		Body: *link,
	}, nil
}

func (h *TopologyHandler) UpdateLink(ctx context.Context, input *struct {
	LinkID string `path:"linkId"`
	Body   topology.Link
}) (*struct {
	Body topology.Link
}, error) {
	existingLink, err := h.topologyService.GetLink(ctx, input.LinkID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get link", err)
	}
	if existingLink == nil {
		return nil, huma.Error404NotFound("Link not found")
	}

	updatedLink := input.Body
	updatedLink.ID = input.LinkID

	if err := h.topologyService.UpdateLink(ctx, updatedLink); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update link", err)
	}

	link, err := h.topologyService.GetLink(ctx, input.LinkID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve updated link", err)
	}

	return &struct {
		Body topology.Link
	}{
		Body: *link,
	}, nil
}

func (h *TopologyHandler) DeleteLink(ctx context.Context, input *struct {
	LinkID string `path:"linkId"`
}) (*struct{}, error) {
	if err := h.topologyService.RemoveLink(ctx, input.LinkID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete link", err)
	}

	return &struct{}{}, nil
}

func (h *TopologyHandler) BulkAddLinks(ctx context.Context, input *struct {
	Body struct {
		Links []topology.Link `json:"links"`
	}
}) (*struct {
	Body []topology.Link
}, error) {
	links := input.Body.Links
	for i := range links {
		links[i].LastSeen = time.Now()
		if links[i].Weight == 0 {
			links[i].Weight = 1.0
		}
		if links[i].Status == "" {
			links[i].Status = "unknown"
		}
		if links[i].Metadata == nil {
			links[i].Metadata = make(map[string]string)
		}
	}

	if err := h.topologyService.BulkAddLinks(ctx, links); err != nil {
		return nil, huma.Error400BadRequest("Failed to bulk add links", err)
	}

	return &struct {
		Body []topology.Link
	}{
		Body: links,
	}, nil
}