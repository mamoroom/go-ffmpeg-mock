package cloudstorage

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/mamoroom/go-ffmpeg-mock/domain/env"
)

type client struct {
	client *storage.Client
}

type Client interface {
	UploadByFilePath(ctx context.Context, srcFilePath string, dstFilePathWithoutExt string, aclRules []storage.ACLRule) (string, error)
	UploadByBinContent(ctx context.Context, bytes []byte, dstFilePath string, aclRules []storage.ACLRule) (string, error)
}

type ACL struct {
	// TODO: public-read以外の権限作る時
}

type ObjectBasePath string

func New(ctx context.Context) Client {
	c, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	return &client{client: c}
}

func (c *client) UploadByFilePath(ctx context.Context, srcFilePath string, dstFilePathWithoutExt string, aclRules []storage.ACLRule) (string, error) {
	f, err := os.Open(srcFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return c.UploadByBinContent(ctx, bytes, dstFilePathWithoutExt+filepath.Ext(srcFilePath), aclRules)
}

func (c *client) UploadByBinContent(ctx context.Context, bytes []byte, dstFilePath string, aclRules []storage.ACLRule) (string, error) {
	writer := c.client.Bucket(env.GCSAssetBucket()).Object(dstFilePath).NewWriter(ctx)
	writer.ContentType = http.DetectContentType(bytes)

	if aclRules == nil {
		// default Public公開
		aclRules = []storage.ACLRule{
			{
				Entity: storage.AllUsers,
				Role:   storage.RoleReader,
			},
		}
	}
	writer.ObjectAttrs.ACL = aclRules
	writer.ObjectAttrs.CacheControl = "no-store"

	_, err := writer.Write(bytes)
	if err != nil {
		writer.Close()
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return GetAssetURL(dstFilePath), nil
}

func GetAssetURL(path string) string {
	return fmt.Sprintf("%s/%s/%s", env.GetGCSPublicURL(), env.GCSAssetBucket(), path)
}

func GetObjectPathFromAssetURL(url string) string {
	return strings.Replace(url, fmt.Sprintf("%s/%s/", env.GetGCSPublicURL(), env.GCSAssetBucket()), "", -1)
}
