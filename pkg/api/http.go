package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/coreweave/ncore-api/pkg/ipxe"
	"github.com/coreweave/ncore-api/pkg/payloads"
	"github.com/go-chi/chi/v5"
)

type httpErrors struct {
	StatusCode int      `json:"-"`
	Errors     []string `json:"errors"`
}

func (e *httpErrors) writeErrors(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(&e); err != nil {
		http.Error(w, fmt.Sprintf(`%s - %s`, http.StatusText(http.StatusInternalServerError), e.Errors), http.StatusInternalServerError)
	}
}

func formatHttpErrors(statusCode int, messages []string) *httpErrors {
	return &httpErrors{
		StatusCode: statusCode,
		Errors:     messages,
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func containsImageTagType(images []ipxe.IpxeImageTagType, iitt ipxe.IpxeImageTagType) bool {
	for _, a := range images {
		if a == iitt {
			return true
		}
	}
	return false
}

// NewHTTPServer creates an HTTPServer for the API.
func NewHTTPServer(i *ipxe.Service, p *payloads.Service) http.Handler {
	s := &HTTPServer{
		ipxe:     i,
		payloads: p,
		router:   chi.NewRouter(),
	}
	s.router.Get("/", s.handleGetRoot)
	s.router.Route("/api/v2/payload", func(r chi.Router) {
		r.Get("/{macAddress}", s.handleGetNodePayload)
		r.Put("/", s.handlePutNodePayload)
		r.Get("/config/{payloadId}", s.handleGetPayloadParameters)
	})
	s.router.Route("/api/v2/ipxe", func(r chi.Router) {
		r.Get("/config/{macAddress}", s.handleGetNodeIpxe)
		r.Put("/", s.handlePutNodeIpxe)
		r.Get("/template/{macAddress}", s.handleGetNodeIpxeTemplate)
		r.Get("/images/", s.handleGetIpxeImages)
		r.Put("/images/", s.handlePutIpxeImages)
		r.Put("/images/{imageName}", s.handlePutIpxeImages)
		r.Delete("/images/", s.handleDeleteIpxeImages)
		r.Get("/s3/{imageName}", s.handleGetIpxeImagePresignedUrls)
	})
	return s.router
}

// HTTPServer exposes payloads.Service via HTTP.
type HTTPServer struct {
	ipxe     *ipxe.Service
	payloads *payloads.Service
	router   *chi.Mux
}

func (s *HTTPServer) handleGetRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ncore-api"))
}

func (s *HTTPServer) handleGetNodePayload(w http.ResponseWriter, r *http.Request) {
	var errors []string
	macAddress := chi.URLParam(r, "macAddress")
	macAddress = strings.Replace(strings.ToLower(macAddress), ":", "", -1)

	// if query includes a hostname instead of full macAddress
	if len(macAddress) == 7 {
		macAddress = "%" + macAddress[1:]
	}

	if len(macAddress) != 12 && len(macAddress) != 7 {
		errors = append(errors, "Invalid mac_address")
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	log.Printf("Checking node_payloads for macAddress: %s", macAddress)
	assignedNodePayload, err := s.payloads.GetNodePayload(r.Context(), macAddress)

	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		// TODO: Add warning log
		return
	case err != nil:
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	case assignedNodePayload == nil && len(macAddress) == 7:
		errors = append(errors, fmt.Sprintf("payload not found for hostname: %s", macAddress))
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	case assignedNodePayload == nil:
		// PayloadId, MacAddress missing from payloads.node_payloads
		// or PayloadId, PayloadDirectory, PayloadSchema missing from payloads.payloads
		var defaultNodePayload *payloads.NodePayload

		requestIp := strings.Split(r.RemoteAddr, ":")[0]
		log.Printf("Checking subnet_default_payloads for requestIp: %s", requestIp)
		subnetPayload, err := s.payloads.GetSubnetDefaultPayload(r.Context(), requestIp)

		switch {
		case err != nil:
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		case subnetPayload != nil:
			log.Printf("Using subnetPayload for macAddress: %s", macAddress)
			defaultNodePayload = &payloads.NodePayload{
				PayloadId:        subnetPayload.PayloadId,
				PayloadDirectory: subnetPayload.PayloadDirectory,
				MacAddress:       macAddress,
			}
		case subnetPayload == nil:
			log.Printf("Using defaultPayload for macAddress: %s", macAddress)
			defaultPayload := s.payloads.GetDefaultPayload(r.Context())
			defaultNodePayload = &payloads.NodePayload{
				PayloadId:        defaultPayload.PayloadId,
				PayloadDirectory: defaultPayload.PayloadDirectory,
				MacAddress:       macAddress,
			}
		}
		defaultNodePayloadDb := &payloads.NodePayloadDb{
			PayloadId:  defaultNodePayload.PayloadId,
			MacAddress: macAddress,
		}
		log.Printf("Adding defaulted node_payloads entry: %v", defaultNodePayloadDb)

		if err := s.payloads.AddNodePayload(r.Context(), defaultNodePayloadDb); err != nil {
			log.Printf("AddNodePayload: %s", err.Error())
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(defaultNodePayload); err != nil {
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(assignedNodePayload); err != nil {
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		}
	}
}

func (s *HTTPServer) handlePutNodePayload(w http.ResponseWriter, r *http.Request) {
	var errors []string
	if r.Header.Get("Content-type") != "application/json" {
		var e = formatHttpErrors(http.StatusUnsupportedMediaType, errors)
		e.writeErrors(w)
		return
	}

	defer r.Body.Close()
	var npd *payloads.NodePayloadDb
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&npd); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	if npd.PayloadId == "" {
		errors = append(errors, "PayloadId is missing.")
	}
	if npd.MacAddress == "" {
		errors = append(errors, "MacAddress is missing.")
	}
	if len(errors) > 0 {
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	npd.MacAddress = strings.Replace(strings.ToLower(npd.MacAddress), ":", "", -1)

	// if query includes a hostname instead of full macAddress
	if len(npd.MacAddress) == 7 {
		npd.MacAddress = "%" + npd.MacAddress[1:]
	}

	if len(npd.MacAddress) != 12 && len(npd.MacAddress) != 7 {
		errors = append(errors, "Invalid mac_address")
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	payloads := s.payloads.GetAvailablePayloads(r.Context())

	if !contains(payloads, npd.PayloadId) {
		errors = append(errors, "PayloadId doesn't exist")
		errors = append(errors, fmt.Sprintf(`Available Payloads: %v`, payloads))
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	config, err := s.payloads.UpdateNodePayload(r.Context(), npd)
	if err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(config); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
}

func (s *HTTPServer) handleGetPayloadParameters(w http.ResponseWriter, r *http.Request) {
	var errors []string
	payloadId := chi.URLParam(r, "payloadId")
	if payloadId == "" || strings.ContainsRune(payloadId, '/') {
		http.NotFound(w, r)
		return
	}
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	parameters, err := s.payloads.GetPayloadParameters(r.Context(), payloadId)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		// TODO: Add warning log
		return
	case err != nil:
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	case parameters == nil:
		var e = formatHttpErrors(http.StatusNotFound, errors)
		e.writeErrors(w)
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		}
	}
}

func (s *HTTPServer) handleGetNodeIpxe(w http.ResponseWriter, r *http.Request) {
	var errors []string
	macAddress := chi.URLParam(r, "macAddress")
	macAddress = strings.Replace(strings.ToLower(macAddress), ":", "", -1)

	// if query includes a hostname instead of full macAddress
	if len(macAddress) == 7 {
		macAddress = "%" + macAddress[1:]
	}

	if len(macAddress) != 12 && len(macAddress) != 7 {
		errors = append(errors, "Invalid mac_address")
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	parameters, err := s.ipxe.GetNodeIpxeConfig(r.Context(), macAddress)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		// TODO: Add warning log
		return
	case err != nil || parameters == nil:
		log.Printf("Getting API Ipxe default image for macAddress: %s", macAddress)
		parameters := s.ipxe.GetIpxeApiDefault()
		s.ipxe.SetHostname(r.Context(), parameters, macAddress)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			errors = append(errors, err.Error())
			var e = formatHttpErrors(http.StatusInternalServerError, errors)
			e.writeErrors(w)
			return
		}
	}
}

func (s *HTTPServer) handlePutNodeIpxe(w http.ResponseWriter, r *http.Request) {
	var errors []string
	if r.Header.Get("Content-type") != "application/json" {
		var e = formatHttpErrors(http.StatusUnsupportedMediaType, errors)
		e.writeErrors(w)
		return
	}

	defer r.Body.Close()
	var indc *ipxe.IpxeNodeDbConfig
	var iitt *ipxe.IpxeImageTagType
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&indc); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	if indc.ImageTag == "" {
		errors = append(errors, "ImageTag is missing.")
	}
	if indc.ImageType == "" {
		errors = append(errors, "ImageType is missing.")
	}
	if indc.MacAddress == "" {
		errors = append(errors, "MacAddress is missing.")
	}
	if len(errors) > 0 {
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	indc.MacAddress = strings.Replace(strings.ToLower(indc.MacAddress), ":", "", -1)

	// if query includes a hostname instead of full macAddress
	if len(indc.MacAddress) == 7 {
		indc.MacAddress = "%" + indc.MacAddress[1:]
	}

	if len(indc.MacAddress) != 12 && len(indc.MacAddress) != 7 {
		errors = append(errors, "Invalid mac_address")
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	images := s.ipxe.GetAvailableImages(r.Context())

	iitt = &ipxe.IpxeImageTagType{
		ImageTag:  indc.ImageTag,
		ImageType: indc.ImageType,
	}

	if !containsImageTagType(images, *iitt) {
		errors = append(errors, "Image doesn't exist")
		errors = append(errors, fmt.Sprintf(`Available Images: %v`, images))
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	config, err := s.ipxe.UpdateNodeImage(r.Context(), indc)
	if err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(config); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
}

func (s *HTTPServer) handleGetIpxeImages(w http.ResponseWriter, r *http.Request) {
	// TODO: List all images
	http.NotFound(w, r)
}

func (s *HTTPServer) handlePutIpxeImages(w http.ResponseWriter, r *http.Request) {
	var errors []string
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	if r.Header.Get("Content-type") != "application/json" {
		var e = formatHttpErrors(http.StatusUnsupportedMediaType, errors)
		e.writeErrors(w)
		return
	}

	defer r.Body.Close()

	var ic *ipxe.IpxeDbConfig
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ic); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	if ic.ImageName == "" {
		errors = append(errors, "ImageName is missing.")
	}
	if ic.ImageCmdline == "" {
		errors = append(errors, "ImageCmdline is missing.")
	}
	if ic.ImageBucket == "" {
		errors = append(errors, "ImageBucket is missing.")
	}
	if ic.ImageTag == "" {
		errors = append(errors, "ImageTag is missing.")
	}
	if ic.ImageType == "" {
		errors = append(errors, "ImageType is missing.")
	}
	if len(errors) > 0 {
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	config, err := s.ipxe.CreateIpxeImage(r.Context(), ic)
	if err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(config); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
}

func (s *HTTPServer) handleDeleteIpxeImages(w http.ResponseWriter, r *http.Request) {
	var errors []string
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	if r.Header.Get("Content-type") != "application/json" {
		var e = formatHttpErrors(http.StatusUnsupportedMediaType, errors)
		e.writeErrors(w)
		return
	}
	defer r.Body.Close()
	var iddc *ipxe.IpxeImageTagType
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&iddc); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	if iddc.ImageTag == "" {
		errors = append(errors, "ImageTag is missing.")
	}
	if iddc.ImageType == "" {
		errors = append(errors, "ImageType is missing.")
	}
	if len(errors) > 0 {
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	config, err := s.ipxe.DeleteIpxeImage(r.Context(), iddc)
	if err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(config); err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}
}

func (s *HTTPServer) handleGetNodeIpxeTemplate(w http.ResponseWriter, r *http.Request) {
	var errors []string

	macAddress := chi.URLParam(r, "macAddress")
	macAddress = strings.ToLower(macAddress)
	macAddress = strings.Replace(macAddress, ":", "", -1)
	macAddress = strings.Replace(macAddress, "%3a", "", -1)

	if len(macAddress) != 12 {
		errors = append(errors, "Invalid mac_address")
		var e = formatHttpErrors(http.StatusBadRequest, errors)
		e.writeErrors(w)
		return
	}

	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)
	ipxeTemplate, err := s.ipxe.GetIpxeConfigTemplate(r.Context())
	if err != nil {
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	}

	assignedIpxeConfig, err := s.ipxe.GetNodeIpxeConfig(r.Context(), macAddress)
	switch {
	case assignedIpxeConfig == nil || err != nil:
		// image_tag, image_type, mac_address missing from ipxe.node_images
		var defaultNodeIpxeConfig *ipxe.IpxeConfig

		requestIp := strings.Split(r.RemoteAddr, ":")[0]
		log.Printf("Checking subnet_default_images for requestIp: %s", requestIp)
		subnetIpxeConfig := s.ipxe.GetSubnetDefaultIpxeConfig(r.Context(), requestIp)

		switch {
		case subnetIpxeConfig != nil:
			log.Printf("Using subnetIpxeConfig for macAddress: %s", macAddress)
			defaultNodeIpxeConfig = subnetIpxeConfig
		case subnetIpxeConfig == nil:
			log.Printf("Using API default IpxeConfig for macAddress: %s", macAddress)
			defaultNodeIpxeConfig = s.ipxe.GetIpxeApiDefault()
		}
		defaultNodeIpxeDbConfig := &ipxe.IpxeNodeDbConfig{
			ImageTag:   defaultNodeIpxeConfig.ImageTag,
			ImageType:  defaultNodeIpxeConfig.ImageType,
			MacAddress: macAddress,
		}
		log.Printf("Adding defaulted node_images entry: %v", defaultNodeIpxeDbConfig)
		if err := s.ipxe.CreateNodeIpxeConfig(r.Context(), defaultNodeIpxeDbConfig); err != nil {
			log.Printf("CreateNodeIpxeConfig: %s", err.Error())
		}
		s.ipxe.SetHostname(r.Context(), defaultNodeIpxeConfig, macAddress)
		ipxeTemplate.Execute(w, defaultNodeIpxeConfig)
	default:
		ipxeTemplate.Execute(w, assignedIpxeConfig)

	}
}

// Accepts an imageName
// Uses ipxeDefaultBucket to get presigned url for image
func (s *HTTPServer) handleGetIpxeImagePresignedUrls(w http.ResponseWriter, r *http.Request) {
	var errors []string
	imageName := chi.URLParam(r, "imageName")
	bucket := ""
	lifetimeSecs := int64(900)
	if imageName == "" || strings.ContainsRune(imageName, '/') {
		http.NotFound(w, r)
		return
	}
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.ipxe.GetIpxeImagePresignedUrls(bucket, imageName, lifetimeSecs)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		return
	case err != nil:
		errors = append(errors, err.Error())
		var e = formatHttpErrors(http.StatusInternalServerError, errors)
		e.writeErrors(w)
		return
	default:
		if imageInitrdUrlHttps == "" {
			errors = append(errors, "imageInitrdUrlHttps not found")
		}
		if imageKernelUrlHttps == "" {
			errors = append(errors, "imageInitrdUrlHttps not found")
		}
		if imageRootFsUrlHttps == "" {
			errors = append(errors, "imageRootFsUrlHttps not found")
		}
		if len(errors) > 0 {
			var e = formatHttpErrors(http.StatusBadRequest, errors)
			e.writeErrors(w)
			return
		}
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("imageInitrdUrlHttps: " + imageInitrdUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageKernelUrlHttps: " + imageKernelUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageRootFsUrlHttps: " + imageRootFsUrlHttps))
	}
}
