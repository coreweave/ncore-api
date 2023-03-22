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
)

type ipxeErrors struct {
  Errors []string
}

func (e *ipxeErrors) writeErrors(w http.ResponseWriter) {
  w.Header().Set("Content-Type", "application/json")
  enc := json.NewEncoder(w)
  enc.SetIndent("", "\t")
  if err := enc.Encode(&e); err != nil {
    http.Error(w, fmt.Sprintf(`%s - %s`, http.StatusText(http.StatusInternalServerError), &e), http.StatusInternalServerError)
  }
}

// NewHTTPServer creates an HTTPServer for the API.
func NewHTTPServer(i *ipxe.Service, p *payloads.Service) http.Handler {
	s := &HTTPServer{
		ipxe:     i,
		payloads: p,
		mux:      http.NewServeMux(),
	}
	// /payload/<macAddress> returns the PayloadId and PayloadDirectory as a json object
	s.mux.HandleFunc("/api/v2/payload/", s.handleGetNodePayload)
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
	return s.mux
}

// HTTPServer exposes payloads.Service via HTTP.
type HTTPServer struct {
	ipxe     *ipxe.Service
	payloads *payloads.Service
	mux      *http.ServeMux
}

func (s *HTTPServer) handleGetNodePayload(w http.ResponseWriter, r *http.Request) {
	macAddress := r.URL.Path[len("/api/v2/payload/"):]
	if macAddress == "" || strings.ContainsRune(macAddress, '/') {
		http.NotFound(w, r)
		return
	}

	payload, err := s.payloads.GetNodePayload(r.Context(), macAddress)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		// TODO: Add warning log
		return
	case err != nil:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
	case payload == nil:
		// TODO: Return a default payload object
		http.Error(w, fmt.Sprintf("payload not found for macAddress: %s", macAddress), http.StatusNotFound)
	default:
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		if err := enc.Encode(payload); err != nil {
			log.Printf("cannot json encode payload request: %v", err)
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
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	parameters, err := s.ipxe.GetIpxeConfig(r.Context(), macAddress)
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
      if ic.ImageDir == "" {
        errors = append(errors, "ImageDir is missing.")
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
    }
		if len(errors) > 0 {
      errorsJson := &ipxeErrors{
        Errors: errors,
      }
      errorsJson.writeErrors(w)
      return
		}
		config, err := s.ipxe.CreateIpxeImage(r.Context(), ic)
		if err != nil {
      errors = append(errors, err.Error())
      errorsJson := &ipxeErrors{
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
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)
	ipxeTemplate, err := s.ipxe.GetIpxeConfigTemplate(r.Context())
	if err != nil {
		log.Printf("handleGetNodeIpxeTemplate: error getting template: %v", err)
	}

	parameters, err := s.ipxe.GetIpxeConfig(r.Context(), macAddress)
	switch {
	case err != nil || parameters == nil:
		log.Printf("Getting API Ipxe default image for macAddress: %s", macAddress)
		parameters := s.ipxe.GetIpxeApiDefault()
		ipxeTemplate.Execute(w, parameters)
	case parameters == nil:
		http.Error(w, "parameters not found", http.StatusNotFound)
	default:
		ipxeTemplate.Execute(w, parameters)
	}
}

// Accepts an imageName
// Uses ipxeDefaultBucket to get presigned url for image
func (s *HTTPServer) handleGetIpxeImagePresignedUrls(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Path[len("/api/v2/ipxe/s3/"):]
	bucket := ""
  imageDir := ""
	lifetimeSecs := int64(900)
	if image == "" || strings.ContainsRune(image, '/') {
		http.NotFound(w, r)
		return
	}
	log.Printf("Request Host: %s", r.Host)
	log.Printf("Request RemoteAddr: %s", r.RemoteAddr)
	log.Printf("Request RequestURI: %s", r.RequestURI)

	imageUrlHttps, imageInitrdUrlHttps, imageKernelUrlHttps, imageRootFsUrlHttps, err := s.ipxe.GetIpxeImagePresignedUrls(bucket, imageDir, image, lifetimeSecs)
	switch {
	case err == context.Canceled, err == context.DeadlineExceeded:
		return
	case err != nil:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err)
	case imageUrlHttps == "":
		http.Error(w, "imageUrlHttps not found", http.StatusNotFound)
  case imageInitrdUrlHttps == "":
		http.Error(w, "imageInitrdUrlHttps not found", http.StatusNotFound)
  case imageKernelUrlHttps == "":
		http.Error(w, "imageKernelUrlHttps not found", http.StatusNotFound)
  case imageRootFsUrlHttps == "":
		http.Error(w, "imageRootFsUrlHttps not found", http.StatusNotFound)
	default:
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("imageUrlHttps: " + imageUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageInitrdUrlHttps: " + imageInitrdUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageKernelUrlHttps: " + imageKernelUrlHttps))
		w.Write([]byte("\n"))
		w.Write([]byte("imageRootFsUrlHttps: " + imageRootFsUrlHttps))
	}
}