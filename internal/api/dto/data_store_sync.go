// Package dto defines request/response payloads for data store sync endpoints.
package dto

import (
	"time"
)

// DataStoreSyncStatus represents the sync status of a data store.
type DataStoreSyncStatus struct {
	LatestJobID     *string   `json:"latest_job_id,omitempty"`
	LatestSyncStatus *string   `json:"latest_sync_status,omitempty"`
	LatestSyncAt    *time.Time `json:"latest_sync_at,omitempty"`
	CanRequeue      bool      `json:"can_requeue"`
	Logs            []string  `json:"logs"`
}

// DataStoreResponse represents the response for a data store with sync status.
type DataStoreResponse struct {
	ID         string              `json:"id"`
	EngineType string              `json:"engine_type"`
	IsActive   bool                `json:"is_active"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	SyncStatus DataStoreSyncStatus `json:"sync_status"`
}

// DataStoreListResponse represents the response for listing data stores.
type DataStoreListResponse struct {
	DataStores []DataStoreResponse `json:"data_stores"`
	Total      int                 `json:"total"`
}