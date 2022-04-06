package env

import (
	"os"
)

func GetGCSPublicURL() string {
	return "https://storage.googleapis.com"
}

func GCPProjectID() string {
	return os.Getenv("GCP_PROJECT_ID")
}

func GCSAssetBucket() string {
	return os.Getenv("GCS_ASSET_BUCKET")
}
