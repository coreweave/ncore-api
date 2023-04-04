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
	"github.com/coreweave/ncore-api/pkg/systems"
)

type jsonErrors struct {
	Errors []string
}

func (e *jsonErrors) writeErrors(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(&e); err != nil {
		http.Error(w, fmt.Sprintf(`%s - %s`, http.StatusText(http.StatusInternalServerError), &e), http.StatusInternalServerError)
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

// NewHTTPServer creates an HTTPServer for the API.
func NewHTTPServer(i *ipxe.Service, p *payloads.Service, sys *systems.Service) http.Handler {
	s := &HTTPServer{
		ipxe:     i,
		payloads: p,
		systems:  sys,
		mux:      http.NewServeMux(),
	}
	// /payload/<macAddress> returns the PayloadId and PayloadDirectory as a json object
	s.mux.HandleFunc("/api/v2/payload/", s.handleNodePayload)
	// /payload/config/<payloadId> returns the payload parameters as a json object
	s.mux.HandleFunc("/api/v2/payload/config/", s.handleGetPayloadParameters)
	// /ipxe/config/<macAddress> returns the IpxeConfig as a json object
	s.mux.HandleFunc("/api/v2/ipxe/config/", s.handleGetNodeIpxe)
	// /ipxe/images/ accepts a json object containing ImageName, ImageBucket, ImageTag, and ImageType
	s.mux.HandleFunc("/api/v2/ipxe/images/", s.handleIpxeImages)
	// /ipxe/template/<macAddress> returns the IpxeConfig as a templated ipxe menu
	s.mux.HandleFunc("/api/v2/ipxe/template/", s.handleGetNodeIpxeTemplate)
	// /ipxe/s3/<imageName> returns the presigned url to download the image as text
	s.mux.HandleFunc("/api/v2/ipxe/s3/", s.handleGetIpxeImagePresignedUrls)
    // /systems/<macAddress> returns the systemId as a json object
	s.mux.HandleFunc("/api/v2/systems/", s.handleSystems)
	return s.mux
}

// HTTPServer exposes payloads.Service via HTTP.
type HTTPServer struct {
	ipxe     *ipxe.Service
	payloads *payloads.Service
	systems  *systems.Service
	mux      *http.ServeMux
}

func (s *HTTPServer) handleNodePayload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "PUT" && r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method not allowed. Only GET, PUT, and DELETE allowed"))
		return
	}
	if (r.Method == "PUT" || r.Method == "DELETE") && r.Header.Get("Content-type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("415 - Unsupported Media Type. Only application/json content-type allowed"))
		return
	}

	switch {
	case r.Method == "GET":
		var errors []string
		macAddress := r.URL.Path[len("/api/v2/payload/"):]
		if macAddress == "" || strings.ContainsRune(macAddress, '/') {
			http.NotFound(w, r)
			return
		}

		macAddress = strings.Replace(strings.ToLower(macAddress), ":", "", -1)

		// if query includes a hostname instead of full macAddress
		if len(macAddress) == 7 {
			macAddress = "%" + macAddress[1:]
		}

		if len(macAddress) != 12 && len(macAddress) != 7 {
			errors = append(errors, "Invalid mac_address")
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}

		log.Printf("Checking node_payloads for macAddress: %s", macAddress)
		assignedNodePayload, err := s.payloads.GetNodePayload(r.Context(), macAddress)

		switch {
		case err == context.Canceled, err == context.DeadlineExceeded:
			// TODO: Add warning log
			return
		case err != nil:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err)
			return
		case assignedNodePayload == nil && len(macAddress) == 7:
			http.Error(w, fmt.Sprintf("payload not found for hostname: %s", macAddress), http.StatusNotFound)
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
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Println(err)
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
				log.Print(err.Error())
			}
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetIndent("", "\t")
			if err := enc.Encode(defaultNodePayload); err != nil {
				log.Printf("cannot json encode payload request: %v", err)
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetIndent("", "\t")
			if err := enc.Encode(assignedNodePayload); err != nil {
				log.Printf("cannot json encode payload request: %v", err)
			}
		}
	case r.Method == "PUT":
		defer r.Body.Close()
		var npd *payloads.NodePayloadDb
		var errors []string
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&npd); err != nil {
			errors = append(errors, fmt.Sprintf(`cannot json decode NodePayload request: %v`, err))
		} else {
			if npd.PayloadId == "" {
				errors = append(errors, "PayloadId is missing.")
			}
			if npd.MacAddress == "" {
				errors = append(errors, "MacAddress is missing.")
			}
		}
		if len(errors) > 0 {
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}

		npd.MacAddress = strings.Replace(strings.ToLower(npd.MacAddress), ":", "", -1)

		// if query includes a hostname instead of full macAddress
		if len(npd.MacAddress) == 7 {
			npd.MacAddress = "%" + npd.MacAddress[1:]
		}

		if len(npd.MacAddress) != 12 && len(npd.MacAddress) != 7 {
			var errors []string
			errors = append(errors, "Invalid mac_address")
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}

		payloads := s.payloads.GetAvailablePayloads(r.Context())

		if !contains(payloads, npd.PayloadId) {
			errors = append(errors, "PayloadId doesn't exist")
			errors = append(errors, fmt.Sprintf(`Available Payloads: %v`, payloads))
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}

		log.Printf("payloads - %v", payloads)
		config, err := s.payloads.UpdateNodePayload(r.Context(), npd)
		if err != nil {
			errors = append(errors, err.Error())
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(config); err != nil {
			log.Printf("cannot json encode payload response: %v", err)
		}
	}
}

func (s *HTTPServer) handleGetPayloadParameters(w http.ResponseWriter, r *http.Request) {
	payloadId := r.URL.Path[len("/api/v2/payload/config/"):]
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
	case parameters == nil:
		http.Error(w, "parameters not found", http.StatusNotFound)
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			log.Printf("cannot json encode payload request: %v", err)
		}
	}
}

func (s *HTTPServer) handleGetNodeIpxe(w http.ResponseWriter, r *http.Request) {
	macAddress := r.URL.Path[len("/api/v2/ipxe/config/"):]
	if macAddress == "" || strings.ContainsRune(macAddress, '/') {
		http.NotFound(w, r)
		return
	}

	macAddress = strings.Replace(strings.ToLower(macAddress), ":", "", -1)

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
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			log.Printf("cannot json encode payload request: %v", err)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(parameters); err != nil {
			log.Printf("cannot json encode payload request: %v", err)
		}
	}
}

func (s *HTTPServer) handleIpxeImages(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	if r.Method != "GET" && r.Method != "PUT" && r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method not allowed. Only GET, PUT, and DELETE allowed"))
		return
	}
	if (r.Method == "PUT" || r.Method == "DELETE") && r.Header.Get("Content-type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("415 - Unsupported Media Type. Only application/json content-type allowed"))
		return
	}

	switch {
	case r.Method == "GET":
		imageName := r.URL.Path[len("/api/v2/ipxe/images/"):]
		if imageName == "" || strings.ContainsRune(imageName, '/') {
			// TODO: List all images
			http.NotFound(w, r)
			return
		}
		// TODO: Get IpxeDbConfig for imageName
		return
	case r.Method == "PUT":
		defer r.Body.Close()
		var ic *ipxe.IpxeDbConfig
		var errors []string
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&ic); err != nil {
			errors = append(errors, fmt.Sprintf(`cannot json decode IpxeDbConfig request: %v`, err))
		} else {
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
		}
		if len(errors) > 0 {
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}
		config, err := s.ipxe.CreateIpxeImage(r.Context(), ic)
		if err != nil {
			errors = append(errors, err.Error())
			errorsJson := &jsonErrors{
				Errors: errors,
			}
			errorsJson.writeErrors(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(config); err != nil {
			log.Printf("cannot json encode payload response: %v", err)
		}
	case r.Method == "DELETE":
		defer r.Body.Close()
		var iddc *ipxe.IpxeDbDeleteConfig
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&iddc); err != nil {
			log.Printf("cannot json decode IpxeDbConfig request: %v", err)
			http.Error(w, fmt.Sprintf(`%s - failed to decode %s`, http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
			return
		}
		var errors []string
		if iddc.ImageTag == "" {
			errors = append(errors, "ImageTag is missing.")
		}
		if iddc.ImageType == "" {
			errors = append(errors, "ImageType is missing.")
		}
		if len(errors) > 0 {
			http.Error(w, fmt.Sprintf(`%s - %s`, http.StatusText(http.StatusUnprocessableEntity), errors), http.StatusUnprocessableEntity)
			return
		}
		config, err := s.ipxe.DeleteIpxeImage(r.Context(), iddc)
		if err != nil {
			http.Error(w, fmt.Sprintf(`%s - %s`, http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
			return
		}
		// TODO: Add to database
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(config); err != nil {
			log.Printf("cannot json encode payload request: %v", err)
		}
	}
}

func (s *HTTPServer) handleGetNodeIpxeTemplate(w http.ResponseWriter, r *http.Request) {
	macAddress := r.URL.Path[len("/api/v2/ipxe/template/"):]
	if macAddress == "" || strings.ContainsRune(macAddress, '/') {
		http.NotFound(w, r)
		return
	}

	macAddress = strings.Replace(strings.ToLower(macAddress), ":", "", -1)

	if len(macAddress) != 12 {
		var errors []string
		errors = append(errors, "Invalid mac_address")
		errorsJson := &jsonErrors{
			Errors: errors,
		}
		errorsJson.writeErrors(w)
		return
	}

	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)
	ipxeTemplate, err := s.ipxe.GetIpxeConfigTemplate(r.Context())
	if err != nil {
		log.Printf("handleGetNodeIpxeTemplate: error getting template: %v", err)
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
			log.Print(err.Error())
		}
		ipxeTemplate.Execute(w, defaultNodeIpxeConfig)
	default:
		ipxeTemplate.Execute(w, assignedIpxeConfig)

	}
}

// Accepts an imageName
// Uses ipxeDefaultBucket to get presigned url for image
func (s *HTTPServer) handleGetIpxeImagePresignedUrls(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Path[len("/api/v2/ipxe/s3/"):]
	bucket := ""
	lifetimeSecs := int64(900)
	if image == "" || strings.ContainsRune(image, '/') {
		http.NotFound(w, r)
		return
	}
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.ipxe.GetIpxeImagePresignedUrls(bucket, image, lifetimeSecs)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		return
	case err != nil:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
	case imageInitrdUrlHttps == "":
		http.Error(w, "imageInitrdUrlHttps not found", http.StatusNotFound)
	case imageKernelUrlHttps == "":
		http.Error(w, "imageKernelUrlHttps not found", http.StatusNotFound)
	case imageRootFsUrlHttps == "":
		http.Error(w, "imageRootFsUrlHttps not found", http.StatusNotFound)
	default:
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("imageInitrdUrlHttps: " + imageInitrdUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageKernelUrlHttps: " + imageKernelUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageRootFsUrlHttps: " + imageRootFsUrlHttps))
	}
}

func (s *HTTPServer) handleSystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method not allowed. Only GET, PUT, and DELETE allowed"))
		return
	}
}
