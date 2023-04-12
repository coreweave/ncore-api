package ipxe

import (
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"
)

type IpxeConfig struct {
	ImageName           string
	ImageBucket         string
	ImageTag            string
	ImageType           string
	ImageInitrdUrlHttp  string
	ImageInitrdUrlHttps string
	ImageKernelUrlHttp  string
	ImageKernelUrlHttps string
	ImageRootFsUrlHttp  string
	ImageRootFsUrlHttps string
	ImageCmdline        string
	Hostname            string
}

type IpxeNodeDbConfig struct {
	ImageTag   string
	ImageType  string
	MacAddress string
}

type IpxeDbConfig struct {
	ImageName    string
	ImageBucket  string
	ImageTag     string
	ImageType    string
	ImageCmdline string
}

func (ic *IpxeConfig) dto() *IpxeConfig {
	return &IpxeConfig{
		ImageName:           ic.ImageName,
		ImageBucket:         ic.ImageBucket,
		ImageTag:            ic.ImageTag,
		ImageType:           ic.ImageType,
		ImageInitrdUrlHttp:  strings.Replace(ic.ImageInitrdUrlHttps, "https", "http", 1),
		ImageInitrdUrlHttps: ic.ImageInitrdUrlHttps,
		ImageKernelUrlHttp:  strings.Replace(ic.ImageKernelUrlHttps, "https", "http", 1),
		ImageKernelUrlHttps: ic.ImageKernelUrlHttps,
		ImageRootFsUrlHttp:  strings.Replace(ic.ImageRootFsUrlHttps, "https", "http", 1),
		ImageRootFsUrlHttps: ic.ImageRootFsUrlHttps,
		ImageCmdline:        ic.ImageCmdline,
		Hostname:            ic.Hostname,
	}
}

func (idc *IpxeDbConfig) dto() *IpxeDbConfig {
	return &IpxeDbConfig{
		ImageName:    idc.ImageName,
		ImageBucket:  idc.ImageBucket,
		ImageTag:     idc.ImageTag,
		ImageType:    idc.ImageType,
		ImageCmdline: idc.ImageCmdline,
	}
}

type IpxeImageTagType struct {
	ImageTag  string
	ImageType string
}

func (s *Service) SetHostname(ctx context.Context, ic *IpxeConfig, macAddress string) *IpxeConfig {
	ic.Hostname = string('g') + macAddress[6:12]
	return ic
}

// GetAvailableImages returns a list of available {image_tag image_type}
func (s *Service) GetAvailableImages(ctx context.Context) []IpxeImageTagType {
	return s.db.GetAvailableImages(ctx)
}

func (s *Service) UpdateNodeImage(ctx context.Context, config *IpxeNodeDbConfig) (*IpxeNodeDbConfig, error) {
	return s.db.UpdateNodeImage(ctx, config)
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) GetNodeIpxeConfig(ctx context.Context, macAddress string) (*IpxeConfig, error) {
	var ic *IpxeConfig
	var idc *IpxeDbConfig
	if macAddress == "" {
		return nil, ValidationError{"missing macAddress"}
	}
	idc, err := s.db.GetIpxeDbConfig(ctx, macAddress)
	if err != nil {
		log.Printf("GetNodeIpxeConfig: failed to get IpxeConfig from database. %v", err)
		return nil, nil
	}
	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
		idc.ImageBucket,
		idc.ImageName,
		900,
	)
	if err != nil {
		log.Printf("GetIpxe error: %v", err)
		return nil, err
	}
	ic = &IpxeConfig{
		ImageName:           idc.ImageName,
		ImageBucket:         idc.ImageBucket,
		ImageTag:            idc.ImageTag,
		ImageType:           idc.ImageType,
		ImageInitrdUrlHttps: imageInitrdUrlHttps,
		ImageKernelUrlHttps: imageKernelUrlHttps,
		ImageRootFsUrlHttps: imageRootFsUrlHttps,
		ImageCmdline:        idc.ImageCmdline,
	}
	s.SetHostname(ctx, ic, macAddress)
	return ic.dto(), nil
}

// GetSubnetDefaultIpxeConfig checks ipxe.subnet_default_images for cidr container ipAddress
// returns an IpxeConfig for the corresponding image_tag and image_type.
func (s *Service) GetSubnetDefaultIpxeConfig(ctx context.Context, ipAddress string) *IpxeConfig {
	var ic *IpxeConfig
	var idc *IpxeDbConfig
	if ipAddress == "" {
		return nil
	}
	idc, err := s.db.GetSubnetDefaultIpxeDbConfig(ctx, ipAddress)
	if err != nil {
		log.Printf("GetSubnetDefaultIpxeDbConfig: failed to get IpxeConfig from database. %v", err)
		return nil
	}
	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
		idc.ImageBucket,
		idc.ImageName,
		900,
	)
	if err != nil {
		log.Printf("GetSubnetDefaultIpxeConfig error for ipAddress: %s - %v", ipAddress, err)
		return nil
	}
	ic = &IpxeConfig{
		ImageName:           idc.ImageName,
		ImageBucket:         idc.ImageBucket,
		ImageTag:            idc.ImageTag,
		ImageType:           idc.ImageType,
		ImageInitrdUrlHttps: imageInitrdUrlHttps,
		ImageKernelUrlHttps: imageKernelUrlHttps,
		ImageRootFsUrlHttps: imageRootFsUrlHttps,
		ImageCmdline:        idc.ImageCmdline,
	}
	return ic.dto()
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) GetIpxeConfigTemplate(ctx context.Context) (template.Template, error) {

	ipxeTemplate := getIpxeConfigTemplate(s.ipxeTemplateFile)

	return ipxeTemplate, nil
}

// CreateNodeIpxeConfig inserts an IpxeNodeDbConfig into ipxe.node_images.
func (s *Service) CreateNodeIpxeConfig(ctx context.Context, ipxeNodeDbConfig *IpxeNodeDbConfig) error {
	err := s.db.CreateNodeIpxeConfig(ctx, ipxeNodeDbConfig)
	return err
}

// GetIpxe returns an IpxeConfig for macAddress.
func (s *Service) CreateIpxeImage(ctx context.Context, config *IpxeDbConfig) (*IpxeConfig, error) {

	ic, err := s.db.CreateIpxeImage(ctx, config)
	if err != nil {
		log.Printf("CreateIpxeImage: failed to insert IpxeDbConfig: %v", err)
		return nil, err
	}
	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
		ic.ImageBucket,
		ic.ImageName,
		900,
	)
	ic.ImageInitrdUrlHttps = imageInitrdUrlHttps
	ic.ImageKernelUrlHttps = imageKernelUrlHttps
	ic.ImageRootFsUrlHttps = imageRootFsUrlHttps
	return ic.dto(), err
}

// DeleteIpxeImage deletes an entry in ipxe.images matching image_tag and image_type.
func (s *Service) DeleteIpxeImage(ctx context.Context, config *IpxeImageTagType) (*IpxeDbConfig, error) {
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
	imageName string,
	lifetimeSecs int64,
) (
	string,
	string,
	string,
	error,
) {
	if bucket == "" {
		bucket = s.ipxeDefaultBucket
	}
	if imageName == "" {
		imageName = s.ipxeDefaultImage
	}
	imageInitrd := fmt.Sprintf(`%s/initrd.img`, imageName)
	imageKernel := fmt.Sprintf(`%s/vmlinuz`, imageName)
	imageRootFs := fmt.Sprintf(`%s/rootfs.cpio.gz`, imageName)

	imageInitrdReq, err := s.s3Presigner.GetObject(bucket, imageInitrd, lifetimeSecs)
	if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", err
	}
	imageKernelReq, err := s.s3Presigner.GetObject(bucket, imageKernel, lifetimeSecs)
	if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", err
	}
	imageRootFsReq, err := s.s3Presigner.GetObject(bucket, imageRootFs, lifetimeSecs)
	if err != nil {
		log.Printf("GetIpxePresignedUrls error: %v", err)
		return "", "", "", err
	}
	imageInitrdUrl := imageInitrdReq.URL
	imageKernelUrl := imageKernelReq.URL
	imageRootFsUrl := imageRootFsReq.URL
	return imageInitrdUrl, imageKernelUrl, imageRootFsUrl, err
}

func (s *Service) GetIpxeApiDefault() *IpxeConfig {
	var ic IpxeConfig
	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.GetIpxeImagePresignedUrls(
		s.ipxeDefaultBucket,
		s.ipxeDefaultImage,
		900,
	)
	if err != nil {
		log.Printf("GetIpxeApiDefault error: %v", err)
		imageInitrdUrlHttps = err.Error()
		imageKernelUrlHttps = err.Error()
		imageRootFsUrlHttps = err.Error()
	}
	defaultCmdline := fmt.Sprintf(`%s/cmdline`, s.ipxeDefaultImage)
	bytes, err := s.s3Svc.GetObject(s.ipxeDefaultBucket, defaultCmdline)
	if err != nil {
		log.Printf("GetIpxeApiDefault: failed to GetObject for defaultCmdline")
	}
	ic.ImageCmdline = string(bytes)
	ic.ImageTag = s.ipxeDefaultImageTag
	ic.ImageType = s.ipxeDefaultImageType
	ic.ImageName = s.ipxeDefaultImage
	ic.ImageBucket = s.ipxeDefaultBucket
	ic.ImageInitrdUrlHttps = imageInitrdUrlHttps
	ic.ImageKernelUrlHttps = imageKernelUrlHttps
	ic.ImageRootFsUrlHttps = imageRootFsUrlHttps
	return ic.dto()
}
