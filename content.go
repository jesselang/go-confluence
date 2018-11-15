package confluence

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type ContentAncestor struct {
	Id string `json:"id"`
}

type Content struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Title  string `json:"title"`
	Body   struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Ancestors []ContentAncestor `json:"ancestors"`
}

type ChildResults struct {
	ResultPagination
	Results []Content `json:"results"`
}

func (w *Wiki) existingContentEndpoint(contentID string) (*url.URL, error) {
	return url.ParseRequestURI(w.endPoint.String() + "/content/" + contentID)
}

func (w *Wiki) contentChildPagesEndpoint(contentID string) (*url.URL, error) {
	return url.ParseRequestURI(w.endPoint.String() + "/content/" + contentID + "/child/page")
}

func (w *Wiki) newContentEndpoint() (*url.URL, error) {
	return url.ParseRequestURI(w.endPoint.String() + "/content")
}

func (w *Wiki) DeleteContent(contentID string) error {
	contentEndPoint, err := w.existingContentEndpoint(contentID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", contentEndPoint.String(), nil)
	if err != nil {
		return err
	}

	_, err = w.sendRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wiki) GetContent(contentID string, expand []string) (*Content, error) {
	contentEndPoint, err := w.existingContentEndpoint(contentID)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Set("expand", strings.Join(expand, ","))
	contentEndPoint.RawQuery = data.Encode()

	req, err := http.NewRequest("GET", contentEndPoint.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var content Content
	err = json.Unmarshal(res, &content)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (w *Wiki) UpdateContent(content *Content) (*Content, error) {
	contentEndPoint, err := w.existingContentEndpoint(content.Id)
	if err != nil {
		return nil, err
	}
	return w.internalCreateOrUpdateContent(content, contentEndPoint, "PUT")
}

func (w *Wiki) CreateContent(content *Content) (*Content, error) {
	contentEndPoint, err := w.newContentEndpoint()
	if err != nil {
		return nil, err
	}
	return w.internalCreateOrUpdateContent(content, contentEndPoint, "POST")
}

func (w *Wiki) internalCreateOrUpdateContent(content *Content, contentEndPoint *url.URL, method string) (*Content, error) {
	jsonBody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, contentEndPoint.String(), bytes.NewReader(jsonBody))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

func (w *Wiki) GetContentChildPages(contentID string, expand []string) (*ChildResults, error) {
	contentEndPoint, err := w.contentChildPagesEndpoint(contentID)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Set("expand", strings.Join(expand, ","))
	contentEndPoint.RawQuery = data.Encode()

	req, err := http.NewRequest("GET", contentEndPoint.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var content ChildResults
	err = json.Unmarshal(res, &content)
	if err != nil {
		return nil, err
	}

	return &content, nil
}
