package storage_test

import (
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMinioStorageProvider_PublicURLEndpointAlignment(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		publicURL    string
		bucket       string
		key          string
		wantURL      string
	}{
		{
			name:      "internal docker hostname rewritten to localhost",
			endpoint:  "minio:9000",
			publicURL: "http://minio:9000",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "http://localhost:9000/catalog/products/shoe.jpg",
		},
		{
			name:      "internal docker hostname without scheme rewritten to localhost",
			endpoint:  "minio:9000",
			publicURL: "minio:9000",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "http://localhost:9000/catalog/products/shoe.jpg",
		},
		{
			name:      "relative public URL preserved without http prefix",
			endpoint:  "minio:9000",
			publicURL: "/media",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "/media/catalog/products/shoe.jpg",
		},
		{
			name:      "empty public URL produces root-relative bucket path",
			endpoint:  "minio:9000",
			publicURL: "",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "/catalog/products/shoe.jpg",
		},
		{
			name:      "standard localhost URL",
			endpoint:  "minio:9000",
			publicURL: "http://localhost:9000",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "http://localhost:9000/catalog/products/shoe.jpg",
		},
		{
			name:      "public CDN URL preserved",
			endpoint:  "minio:9000",
			publicURL: "https://cdn.example.com",
			bucket:    "catalog",
			key:       "products/shoe.jpg",
			wantURL:   "https://cdn.example.com/catalog/products/shoe.jpg",
		},
		{
			name:      "spaces and special characters escaped",
			endpoint:  "minio:9000",
			publicURL: "http://localhost:9000",
			bucket:    "catalog",
			key:       "products/my new shoes & socks.jpg",
			wantURL:   "http://localhost:9000/catalog/products/my%20new%20shoes%20&%20socks.jpg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prov, err := storage.NewMinioStorageProvider(tc.endpoint, "admin", "supersecret", tc.publicURL)
			require.NoError(t, err)
			assert.NotNil(t, prov)

			url := prov.GetPublicURL(tc.bucket, tc.key)
			assert.Equal(t, tc.wantURL, url)
		})
	}
}
