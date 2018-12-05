package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// NOTE: The `_links.download` key is not documented in the Confluence REST API Reference. However,
//       it is referenced here: https://community.atlassian.com/t5/Answers-Developer-Questions/confluence-pages-get-attached-files-via-REST-API/qaq-p/529873#M67529
type Attachment struct {
	ID string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Links struct {
		Download string `json:"download,omitempty"`
	} `json:"_links,omitempty"`
}

type AttachmentResults struct {
	ResultPagination
	Results []Attachment `json:"results"`
}

func (w *Wiki) getAttachmentByFilenameEndpoint(contentID string, filename string) (*url.URL, error) {
	return url.ParseRequestURI(fmt.Sprintf("%v/content/%v/child/attachment?filename=%v", w.endPoint.String(), contentID, filename))
}

func (w *Wiki) createAttachmentEndpoint(contentID string) (*url.URL, error) {
	return url.ParseRequestURI(fmt.Sprintf("%v/content/%v/child/attachment", w.endPoint.String(), contentID))
}

func (w *Wiki) updateAttachmentEndpoint(contentID string, attachmentID string) (*url.URL, error) {
	return url.ParseRequestURI(fmt.Sprintf("%v/content/%v/child/attachment/%v/data", w.endPoint.String(), contentID, attachmentID))
}

func (w *Wiki) GetAttachment(contentID string, filename string) (*AttachmentResults, error) {
	endpoint, err := w.getAttachmentByFilenameEndpoint(contentID, filename)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var results AttachmentResults
	err = json.Unmarshal(res, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}

func (w *Wiki) GetAttachmentData(contentID string, filename string) ([]byte, error) {
	result, err := w.GetAttachment(contentID, filename)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", w.root.String() + result.Results[0].Links.Download, nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (w *Wiki) CreateAttachment(contentID string, path string) (*AttachmentResults, error) {
	endpoint, err := w.createAttachmentEndpoint(contentID)
	if err != nil {
		return nil, err
	}
	return w.createOrUpdateAttachment(endpoint, path)
}

func (w *Wiki) UpdateAttachment(contentID string, path string, attachmentID string) (*AttachmentResults, error) {
	endpoint, err := w.updateAttachmentEndpoint(contentID, attachmentID)
	if err != nil {
		return nil, err
	}
	return w.createOrUpdateAttachment(endpoint, path)
}

func (w *Wiki) createOrUpdateAttachment(endpoint *url.URL, path string) (*AttachmentResults, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint.String(), body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// From https://docs.atlassian.com/ConfluenceServer/rest/latest/#api/content/{id}/child/attachment-createAttachments
	//   In order to protect against XSRF attacks, because this method accepts multipart/form-data,
	//   it has XSRF protection on it. This means you must submit a header of
	//   X-Atlassian-Token: nocheck with the request, otherwise it will be blocked.
	req.Header.Add("X-Atlassian-Token", "nocheck")

	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var results AttachmentResults
	err = json.Unmarshal(res, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
