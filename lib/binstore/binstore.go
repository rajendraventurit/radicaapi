package binstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	"github.com/BurntSushi/graphics-go/graphics"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rajendraventurit/radicaapi/lib/logger"
)

var localConf *config

const defConfPath = "/etc/radica/binstore.json"

type config struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
	Region string `json:"region"`
}

// Configure will configure the smtp server
func Configure() error {
	f, err := os.Open(defConfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	conf := config{}
	err = json.NewDecoder(f).Decode(&conf)
	if err != nil {
		return conErr("binstore.Configure", err)
	}
	// Keys are set to the env. IF an .aws file exists it may override
	os.Setenv("AWS_ACCESS_KEY_ID", conf.Key)
	os.Setenv("AWS_SECRET_ACCESS_KEY", conf.Secret)
	os.Setenv("AWS_REGION", conf.Region)
	localConf = &conf
	return nil
}

// DownloadS3 will attempt to download a file from S3
func DownloadS3(key string) ([]byte, error) {
	if localConf == nil {
		if err := Configure(); err != nil {
			return nil, err
		}
	}
	// Keys are set to the env. IF an .aws file exists it may override
	sess, err := session.NewSession()
	if err != nil {
		return nil, conErr("DownloadS3 NewSession", err)
	}

	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sess)

	wb := aws.WriteAtBuffer{}
	// Write the contents of S3 Object to the file
	_, err = downloader.Download(&wb, &s3.GetObjectInput{
		Bucket: aws.String(localConf.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, conErr("downloader.Download", err)
	}

	return wb.Bytes(), nil
}

// UploadS3 will upload an image to S3
func UploadS3(s3name string, r io.Reader) error {
	logger.Debugf("Starting UploadS3")
	if localConf == nil {
		if err := Configure(); err != nil {
			logger.Debugf("Failed conf %v", err)
			return err
		}
	}
	// Keys are set to the env. IF an .aws file exists it may override
	sess, err := session.NewSession()
	if err != nil {
		logger.Debugf("Failed NewSession %v", err)
		return conErr("uploadS3 NewSession", err)
	}

	// Create a uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload input parameters
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(localConf.Bucket),
		Key:    aws.String(s3name),
		Body:   r,
	}

	// Perform an upload.
	if _, err = uploader.Upload(upParams); err != nil {
		logger.Debugf("Failed upload %v", err)
		return conErr("uploader.Upload", err)
	}
	return nil
}

// Thumbnail will return a thumbnail version of a jpg image
func Thumbnail(i io.Reader, imgType string) ([]byte, error) {
	srcImage, _, err := image.Decode(i)
	if err != nil {
		return nil, err
	}
	// Dimension of new thumbnail 80 X 80
	dstImage := image.NewRGBA(image.Rect(0, 0, 80, 80))
	// Thumbnail function of Graphics
	graphics.Thumbnail(dstImage, srcImage)
	img := bytes.Buffer{}
	err = jpeg.Encode(bufio.NewWriter(&img), dstImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
	return img.Bytes(), err
}

func conErr(con string, err error) error {
	return fmt.Errorf("%s %v", con, err)
}
