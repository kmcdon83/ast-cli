package wrappers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	scansApi "github.com/checkmarxDev/scans/pkg/api/scans"

	scansRestApi "github.com/checkmarxDev/scans/pkg/api/scans/rest/v1"
	"github.com/pkg/errors"
)

const (
	failedToParseGetAll = "Failed to parse list response"
	failedToParseTags   = "Failed to parse tags response"
)

type ScansHTTPWrapper struct {
	path        string
	contentType string
}

func NewHTTPScansWrapper(path string) ScansWrapper {
	return &ScansHTTPWrapper{
		path:        path,
		contentType: "application/json",
	}
}

func (s *ScansHTTPWrapper) Create(model *scansRestApi.Scan) (*scansRestApi.ScanResponseModel, *scansRestApi.ErrorModel, error) {
	jsonBytes, err := json.Marshal(model)
	if err != nil {
		return nil, nil, err
	}
	resp, err := SendHTTPRequest(http.MethodPost, s.path, bytes.NewBuffer(jsonBytes), true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	return handleScanResponseWithBody(resp, err, http.StatusCreated)
}

func (s *ScansHTTPWrapper) Get(params map[string]string) (*scansRestApi.ScansCollectionResponseModel, *scansRestApi.ErrorModel, error) {
	resp, err := SendHTTPRequestWithQueryParams(http.MethodGet, s.path, params, nil, DefaultTimeoutSeconds)
	if err != nil {
		return nil, nil, err
	}
	decoder := json.NewDecoder(resp.Body)

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError:
		errorModel := scansRestApi.ErrorModel{}
		err = decoder.Decode(&errorModel)
		if err != nil {
			return nil, nil, errors.Wrapf(err, failedToParseGetAll)
		}
		return nil, &errorModel, nil
	case http.StatusOK:
		model := scansRestApi.ScansCollectionResponseModel{}
		err = decoder.Decode(&model)
		if err != nil {
			return nil, nil, errors.Wrapf(err, failedToParseGetAll)
		}
		return &model, nil, nil
	case http.StatusNotFound:
		return nil, nil, errors.Errorf("scan not found")
	default:
		return nil, nil, errors.Errorf("response status code %d", resp.StatusCode)
	}
}

func (s *ScansHTTPWrapper) GetByID(scanID string) (*scansRestApi.ScanResponseModel, *scansRestApi.ErrorModel, error) {
	resp, err := SendHTTPRequest(http.MethodGet, s.path+"/"+scanID, nil, true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, nil, err
	}
	return handleScanResponseWithBody(resp, err, http.StatusOK)
}

func (s *ScansHTTPWrapper) GetWorkflowByID(scanID string) ([]*ScanTaskResponseModel, *scansRestApi.ErrorModel, error) {
	path := fmt.Sprintf("%s/%s/workflow", s.path, scanID)
	resp, err := SendHTTPRequest(http.MethodGet, path, nil, true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, nil, err
	}
	return handleWorkflowResponseWithBody(resp, err)
}

func handleWorkflowResponseWithBody(resp *http.Response, err error) ([]*ScanTaskResponseModel, *scansRestApi.ErrorModel, error) {
	if err != nil {
		return nil, nil, err
	}
	decoder := json.NewDecoder(resp.Body)

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError:
		errorModel := scansRestApi.ErrorModel{}
		err = decoder.Decode(&errorModel)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed to parse workflow response")
		}
		return nil, &errorModel, nil
	case http.StatusOK:
		model := []*ScanTaskResponseModel{}
		err = decoder.Decode(&model)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed to parse workflow response")
		}
		return model, nil, nil

	default:
		return nil, nil, errors.Errorf("response status code %d", resp.StatusCode)
	}
}

func (s *ScansHTTPWrapper) Delete(scanID string) (*scansRestApi.ErrorModel, error) {
	resp, err := SendHTTPRequest(http.MethodDelete, s.path+"/"+scanID, nil, true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, err
	}
	return handleScanResponseWithNoBody(resp, err, http.StatusNoContent)
}

func (s *ScansHTTPWrapper) Cancel(scanID string) (*scansRestApi.ErrorModel, error) {
	b, err := json.Marshal(&scansRestApi.CancelScanModel{
		Status: scansApi.ScanCanceled,
	})
	if err != nil {
		return nil, err
	}

	resp, err := SendHTTPRequest(http.MethodPatch, s.path+"/"+scanID, bytes.NewBuffer(b), true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, err
	}

	return handleScanResponseWithNoBody(resp, err, http.StatusNoContent)
}

func (s *ScansHTTPWrapper) Tags() (map[string][]string, *scansRestApi.ErrorModel, error) {
	resp, err := SendHTTPRequest(http.MethodGet, s.path+"/tags", nil, true, DefaultTimeoutSeconds)
	if err != nil {
		return nil, nil, err
	}
	decoder := json.NewDecoder(resp.Body)

	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusInternalServerError:
		errorModel := scansRestApi.ErrorModel{}
		err = decoder.Decode(&errorModel)
		if err != nil {
			return nil, nil, errors.Wrapf(err, failedToParseTags)
		}
		return nil, &errorModel, nil
	case http.StatusOK:
		tags := map[string][]string{}
		err = decoder.Decode(&tags)
		if err != nil {
			return nil, nil, errors.Wrapf(err, failedToParseTags)
		}
		return tags, nil, nil

	default:
		return nil, nil, errors.Errorf("response status code %d", resp.StatusCode)
	}
}
