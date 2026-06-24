// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	l4g "github.com/alecthomas/log4go"
	"github.com/disintegration/imaging"
	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
	s3 "github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/credentials"
	"github.com/rwcarlsen/goexif/exif"
	_ "golang.org/x/image/bmp"
)

const (
	/*
	  EXIF Image Orientations
	  1        2       3      4         5            6           7          8

	  888888  888888      88  88      8888888888  88                  88  8888888888
	  88          88      88  88      88  88      88  88          88  88      88  88
	  8888      8888    8888  8888    88          8888888888  8888888888          88
	  88          88      88  88
	  88          88  888888  888888
	*/
	Upright            = 1
	UprightMirrored    = 2
	UpsideDown         = 3
	UpsideDownMirrored = 4
	RotatedCWMirrored  = 5
	RotatedCCW         = 6
	RotatedCCWMirrored = 7
	RotatedCW          = 8

	MaxImageSize                 = 6048 * 4032 // 24 megapixels, roughly 36MB as a raw image
	IMAGE_THUMBNAIL_PIXEL_WIDTH  = 120
	IMAGE_THUMBNAIL_PIXEL_HEIGHT = 100
	IMAGE_PREVIEW_PIXEL_WIDTH    = 1024
)

// Similar to s3.New() but allows initialization of signature v2 or signature v4 client.
// If signV2 input is false, function always returns signature v4.
//
// Additionally this function also takes a user defined region, if set
// disables automatic region lookup.
func s3New(endpoint, accessKey, secretKey string, secure bool, signV2 bool, region string) (*s3.Client, error) {
	var creds *credentials.Credentials
	if signV2 {
		creds = credentials.NewStatic(accessKey, secretKey, "", credentials.SignatureV2)
	} else {
		creds = credentials.NewStatic(accessKey, secretKey, "", credentials.SignatureV4)
	}
	return s3.NewWithCredentials(endpoint, creds, secure, region)
}

// -- 8< -- SNIP -- 8< -- 


func DoUploadFile(teamId string, channelId string, userId string, rawFilename string, data []byte) (*model.FileInfo, *model.AppError) {
	filename := filepath.Base(rawFilename)

	info, err := model.GetInfoForBytes(filename, data)
	if err != nil {
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}

	info.Id = model.NewId()
	info.CreatorId = userId

	pathPrefix := "teams/" + teamId + "/channels/" + channelId + "/users/" + userId + "/" + info.Id + "/"
	info.Path = pathPrefix + filename

	if info.IsImage() {
		// Check dimensions before loading the whole thing into memory later on
		if info.Width*info.Height > MaxImageSize {
			err := model.NewLocAppError("uploadFile", "api.file.upload_file.large_image.app_error", map[string]interface{}{"Filename": filename}, "")
			err.StatusCode = http.StatusBadRequest
			return nil, err
		}

		nameWithoutExtension := filename[:strings.LastIndex(filename, ".")]
		info.PreviewPath = pathPrefix + nameWithoutExtension + "_preview.jpg"
		info.ThumbnailPath = pathPrefix + nameWithoutExtension + "_thumb.jpg"
	}

	if err := WriteFile(data, info.Path); err != nil {
		return nil, err
	}

	if result := <-Srv.Store.FileInfo().Save(info); result.Err != nil {
		return nil, result.Err
	}

	return info, nil
}
// -- 8< -- SNIP -- 8< -- 
 
