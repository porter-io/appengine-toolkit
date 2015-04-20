package fetcher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/pubsub/v1beta2"
	"google.golang.org/appengine"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type FetchRequest struct {
	URLs    []string        `json:"urls"`
	Topic   string          `json:"topic"`
	Context context.Context `json:"-"`
	Request *http.Request   `json:"-"`
}

type FetchResponse struct {
	URL     string
	Content []byte
}

type FetchError struct {
	URL   string
	Error error
}

type Fetcher struct {
	Raw   bool
	Topic string
}

type FetchStat struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
}

func NewFetcher(raw bool, topic string) *Fetcher {
	return &Fetcher{Raw: raw, Topic: topic}
}

func (f *FetchRequest) AppID() string {
	return appengine.AppID(f.Context)
}

func (f *Fetcher) fetchURL(r *FetchRequest, url string) (result *FetchResponse, err error) {
	resp, err := urlfetch.Client(r.Context).Get(url)
	if err != nil {
		return
	}
	var content []byte
	if f.Raw {
		content, err = httputil.DumpResponse(resp, true)
	} else {
		defer resp.Body.Close()
		content, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil {
		return
	}
	result = &FetchResponse{
		URL:     url,
		Content: content,
	}
	return
}

func (f *Fetcher) Fetch(request *FetchRequest) (result []*FetchResponse, errors []*FetchError) {
	resc, errc := make(chan *FetchResponse), make(chan *FetchError)

	for _, url := range request.URLs {
		go func(url string) {
			resp, err := f.fetchURL(request, url)
			if err != nil {
				errc <- &FetchError{URL: url, Error: err}
				return
			}
			resc <- resp
		}(url)
	}

	result = make([]*FetchResponse, 0)
	errors = make([]*FetchError, 0)

	for i := 0; i < len(request.URLs); i++ {
		select {
		case res := <-resc:
			result = append(result, res)
		case err := <-errc:
			errors = append(errors, err)
		}
	}
	return
}

func (f *Fetcher) Retry(request *FetchRequest, errors []*FetchError) error {
	retryRequest := FetchRequest{Topic: request.Topic}
	retryRequest.URLs = make([]string, len(errors))
	for i := range errors {
		retryRequest.URLs[i] = errors[i].URL
	}
	content, err := json.Marshal(&retryRequest)
	if err != nil {
		return err
	}

	t := &taskqueue.Task{
		Path:    request.Request.URL.Path,
		Payload: content,
		Method:  "POST",
	}
	if _, err := taskqueue.Add(request.Context,
		t, request.Request.Header.Get("X-AppEngine-QueueName")); err != nil {
		return err
	}
	return nil
}

func (f *Fetcher) Publish(request *FetchRequest, entries []*FetchResponse) (err error) {
	client, err := google.DefaultClient(request.Context, pubsub.CloudPlatformScope)
	if err != nil {
		return
	}
	service, err := pubsub.New(client)
	if err != nil {
		return
	}
	messages := make([]*pubsub.PubsubMessage, len(entries))
	for i := range entries {
		messages[i] = &pubsub.PubsubMessage{
			Data: base64.StdEncoding.EncodeToString(entries[i].Content),
		}
	}
	pr := pubsub.PublishRequest{
		Messages: messages,
	}
	var topic string
	if request.Topic != "" {
		topic = request.Topic
	} else {
		topic = f.Topic
	}
	if topic == "" {
		return fmt.Errorf("fetcher: topic is empty")
	}

	full := fmt.Sprintf("projects/%s/topics/%s", request.AppID(), topic)
	_, err = service.Projects.Topics.Publish(full, &pr).Do()
	if err != nil {
		return
	}
	return
}

func (f *Fetcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	request := FetchRequest{Context: c, Request: r}

	// Decode request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Do fetch
	result, errors := f.Fetch(&request)

	// Handle errors
	if len(errors) > 0 {
		if err := f.Retry(&request, errors); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Do publish
	if len(result) > 0 {
		err = f.Publish(&request, result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Write stat to response
	s := FetchStat{Total: len(request.URLs), Success: len(result), Fail: len(errors)}
	body, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, body)
}
