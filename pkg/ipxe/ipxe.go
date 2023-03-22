package ipxe

import (
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"
)

type IpxeConfig struct {
  ImageDir      string
	ImageName     string
	ImageBucket   string
	ImageTag      string
	ImageType     string
	ImageUrlHttp  string
	ImageUrlHttps string
  ImageInitrdUrlHttp  string
	ImageInitrdUrlHttps string
  ImageKernelUrlHttp  string
	ImageKernelUrlHttps string
  ImageRootFsUrlHttp  string
	ImageRootFsUrlHttps string
  ImageCmdline string
}

type IpxeDbConfig struct {
  ImageDir    string
	ImageName   string
	ImageBucket string
	ImageTag    string
	ImageType   string
  ImageCmdline string
}

func (ic *IpxeConfig) dto() *IpxeConfig {
	return &IpxeConfig{
    ImageDir:      ic.ImageDir,
		ImageName:     ic.ImageName,
		ImageBucket:   ic.ImageBucket,
		ImageTag:      ic.ImageTag,
		ImageType:     ic.ImageType,
		ImageUrlHttp:  strings.Replace(ic.ImageUrlHttps, "https", "http", 1),
		ImageUrlHttps: ic.ImageUrlHttps,
    ImageInitrdUrlHttp:  strings.Replace(ic.ImageInitrdUrlHttps, "https", "http", 1),
    ImageInitrdUrlHttps: ic.ImageInitrdUrlHttps,
    ImageKernelUrlHttp:  strings.Replace(ic.ImageKernelUrlHttps, "https", "http", 1),
    ImageKernelUrlHttps: ic.ImageKernelUrlHttps,
    ImageRootFsUrlHttp:  strings.Replace(ic.ImageRootFsUrlHttps, "https", "http", 1),
    ImageRootFsUrlHttps: ic.ImageRootFsUrlHttps,
    ImageCmdline: ic.ImageCmdline,
	}
}

func (idc *IpxeDbConfig) dto() *IpxeDbConfig {
	return &IpxeDbConfig{
    ImageDir:      idc.ImageDir,
		ImageName:     idc.ImageName,
		ImageBucket:   idc.ImageBucket,
		ImageTag:      idc.ImageTag,
		ImageType:     idc.ImageType,
    ImageCmdline:  idc.ImageCmdline,
	}
}

type IpxeDbDeleteConfig struct {
	ImageTag    string
	ImageType   string
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) GetIpxeConfig(ctx context.Context, macAddress string) (*IpxeConfig, error) {
	var ic *IpxeConfig
	var idc *IpxeDbConfig
	if macAddress == "" {
		return nil, ValidationError{"missing macAddress"}
	}
	idc, err := s.db.GetIpxeDbConfig(ctx, macAddress)
	if err != nil {
		log.Printf("GetIpxeConfig: failed to get IpxeConfig from database. %v", err)
		return nil, err
	}
  imageUrlHttps, imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
    idc.ImageBucket,
    idc.ImageDir,
    idc.ImageName,
    900,
  )
	if err != nil {
		log.Printf("GetIpxe error: %v", err)
		return nil, err
	}
	ic = &IpxeConfig{
    ImageDir:      idc.ImageDir,
		ImageName:     idc.ImageName,
		ImageBucket:   idc.ImageBucket,
		ImageTag:      idc.ImageTag,
		ImageType:     idc.ImageType,
		ImageUrlHttps: imageUrlHttps,
    ImageInitrdUrlHttps: imageInitrdUrlHttps,
    ImageKernelUrlHttps: imageKernelUrlHttps,
    ImageRootFsUrlHttps: imageRootFsUrlHttps,
		ImageCmdline:     idc.ImageCmdline,
	}
	return ic.dto(), nil
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) GetIpxeConfigTemplate(ctx context.Context) (template.Template, error) {

	ipxeTemplate := GetIpxeConfigTemplate(s.ipxeTemplateFile)

	return ipxeTemplate, nil
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) CreateIpxeImage(ctx context.Context, config *IpxeDbConfig) (*IpxeConfig, error) {

	ic, err := s.db.CreateIpxeImage(ctx, config)
	if err != nil {
		log.Printf("CreateIpxeImage: failed to insert IpxeDbConfig: %v", err)
		return nil, err
	}
	imageUrlHttps, imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
      ic.ImageBucket,
      ic.ImageDir,
      ic.ImageName,
      900,
  )
	ic.ImageUrlHttps = imageUrlHttps
	ic.ImageInitrdUrlHttps = imageInitrdUrlHttps
	ic.ImageKernelUrlHttps = imageKernelUrlHttps
	ic.ImageRootFsUrlHttps = imageRootFsUrlHttps
	return ic.dto(), err
}

// DeleteIpxeImage deletes an entry in ipxe.images matching image_tag and image_type.
func (s *Service) DeleteIpxeImage(ctx context.Context, config *IpxeDbDeleteConfig) (*IpxeDbConfig, error) {
	idc, err := s.db.DeleteIpxeImage(ctx, config)
	if err != nil {
		log.Printf("DeleteIpxeImage: failed to delete IpxeDbDeleteConfig: %v", err)
		return nil, err
	}
	return idc, err
}

// GetIpxePresignedUrl returns a url string for the given bucket.
func (s *Service) GetIpxeImagePresignedUrls(
	bucket string,
  imageDir string,
	imageName string,
	lifetimeSecs int64,
) (
    string,
    string,
    string,
    string,
    error,
) {
	if bucket == "" {
		bucket = s.ipxeDefaultBucket
	}
  if imageDir == "" {
		imageDir = s.ipxeDefaultImageDir
	}
  image := fmt.Sprintf(`%s/%s`, imageDir, imageName)
  imageInitrd := fmt.Sprintf(`%s/initrd.img`, imageDir)
  imageKernel := fmt.Sprintf(`%s/vmlinuz`, imageDir)
  imageRootFs := fmt.Sprintf(`%s/rootfs.cpio.gz`, imageDir)
	imageReq, err := s.s3Presigner.GetObject(bucket, image, lifetimeSecs)
  if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", "", err
	}
	imageInitrdReq, err := s.s3Presigner.GetObject(bucket, imageInitrd, lifetimeSecs)
  if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", "", err
	}
  imageKernelReq, err := s.s3Presigner.GetObject(bucket, imageKernel, lifetimeSecs)
  if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", "", err
	}
	imageRootFsReq, err := s.s3Presigner.GetObject(bucket, imageRootFs, lifetimeSecs)
	if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", "", err
	}
	imageUrl := imageReq.URL
	imageInitrdUrl := imageInitrdReq.URL
	imageKernelUrl := imageKernelReq.URL
	imageRootFsUrl := imageRootFsReq.URL
	return imageUrl, imageInitrdUrl, imageKernelUrl, imageRootFsUrl, err
}

func (s *Service) GetIpxeApiDefault() *IpxeConfig {
	var ic IpxeConfig
	imageUrlHttps, imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
      s.ipxeDefaultBucket,
      s.ipxeDefaultImageDir,
      s.ipxeDefaultImage,
      900,
  )
	if err != nil {
		log.Printf("GetIpxeApiDefault error: %v", err)
		imageUrlHttps = err.Error()
    imageInitrdUrlHttps = err.Error()
    imageKernelUrlHttps = err.Error()
    imageRootFsUrlHttps = err.Error()
	}
  defaultCmdline := fmt.Sprintf(`%s/cmdline`, s.ipxeDefaultImageDir)
  bytes, err := s.s3Svc.GetObject( s.ipxeDefaultBucket, defaultCmdline)
  if err != nil {
    log.Printf("GetIpxeApiDefault: failed to GetObject for defaultCmdline")
  }
  ic.ImageCmdline = string(bytes)
  ic.ImageDir = s.ipxeDefaultImageDir
	ic.ImageTag = "default"
	ic.ImageType = "default"
	ic.ImageUrlHttps = imageUrlHttps
	ic.ImageName = s.ipxeDefaultImage
	ic.ImageBucket = s.ipxeDefaultBucket
  ic.ImageInitrdUrlHttps = imageInitrdUrlHttps
	ic.ImageKernelUrlHttps = imageKernelUrlHttps
	ic.ImageRootFsUrlHttps = imageRootFsUrlHttps
	return ic.dto()
}