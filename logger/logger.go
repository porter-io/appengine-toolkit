package logger

import (
	gae "appengine"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/logging/v1beta3"
	"google.golang.org/appengine"
	"net/http"
)

type GCLLogger struct {
	Target   string
	Severity string
}

type GAELogger struct{}

func (l *GCLLogger) Log(res *responseLogger, r *http.Request) (err error) {
	if res.Status() == http.StatusOK {
		context := appengine.NewContext(r)
		client, err := google.DefaultClient(appengine.NewContext(r),
			logging.CloudPlatformScope)
		if err != nil {
			return err
		}
		service, err := logging.New(client)
		if err != nil {
			return err
		}

		projectId := appengine.AppID(context)

		labels := map[string]string{
			"compute.googleapis.com/resource_type": "instance",
			"compute.googleapis.com/resource_id":   appengine.InstanceID(),
		}

		meta := &logging.LogEntryMetadata{
			Severity:    l.Severity,
			ProjectId:   projectId,
			ServiceName: "compute.googleapis.com",
			Zone:        appengine.Datacenter(context),
		}
		entry := &logging.LogEntry{Metadata: meta, TextPayload: string(res.Body())}

		e := &logging.WriteLogEntriesRequest{
			CommonLabels: labels,
			Entries:      []*logging.LogEntry{entry},
		}
		_, err = service.Projects.Logs.Entries.Write(projectId, l.Target, e).Do()
	}
	return
}

func (l *GAELogger) Log(res *responseLogger, r *http.Request) (err error) {
	if res.Status() >= 500 {
		c := gae.NewContext(r)
		c.Errorf(string(res.Body()))
	}
	return nil
}

type Logger interface {
	Log(*responseLogger, *http.Request) error
}

func LoggingMiddleware(next http.Handler, l Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := responseLogger{w: w}
		next.ServeHTTP(&res, r)
		l.Log(&res, r)
	})
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
	body   []byte
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.body = append(l.body, b...)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Body() []byte {
	return l.body
}
