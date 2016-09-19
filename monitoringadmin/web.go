package monitoringadmin

import (
	"golang.org/x/net/context"
	"google.golang.org/api/cloudmonitoring/v2beta2"
	"html/template"
	"net/http"
	"strings"
)

const pageHTML = `
<html>
  <body>
    <h1>Metrics</h1>
    {{ range . }}
	  <h2>{{ .Name }}</h2>
	  <p>Description: {{ .Description }}</p>
	  <p>MetricType: {{ .TypeDescriptor.MetricType }}</p>
	  <p>ValueType: {{ .TypeDescriptor.ValueType }}</p>
	  <p>
	  <form action="metric/delete/" method="post">
	  	<input type="hidden" name="metric" value="{{ .Name }}">
	  	<input type="submit" value="Delete">
	  </form>
	  </p>
	{{ end }}
	<h1>Add Metric</h1>
	<form action="metric/create/" method="post">
	<p>Name:</p>
	<p><input type="text" name="metric"></p>
	<p>Description:</p>
	<p><input type="text" name="description"></p>
	<p>Type:</p>
	<p>
	<select name="type">
		<option value="delta">delta</option>
		<option value="gauge">gauge</option>
	</select>
	</p>
	<p>Value Type:</p>
	<p>
	<select name="valuetype">
		<option value="double">double</option>
		<option value="int64">int64</option>
	</select>
	</p>
	<p><input type="submit"></p>
	</form>
  </body>
</html>
`

var pageTemplate = template.Must(template.New("page").Parse(pageHTML))

func HandleIndex(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	allMetrics, err := ListMetric(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var metrics []*cloudmonitoring.MetricDescriptor

	for _, m := range allMetrics {
		if strings.HasPrefix(m.Name, "custom") {
			metrics = append(metrics, m)
		}
	}

	err = pageTemplate.Execute(w, metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleCreateTopic(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	metric := r.FormValue("metric")
	description := r.FormValue("description")
	metricType := r.FormValue("type")
	valueType := r.FormValue("valuetype")
	if metric != "" && metricType != "" && valueType != "" {
		if err := CreateMetric(ctx, metric,
			description, metricType, valueType); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}

func HandleDeleteTopic(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	metric := r.FormValue("metric")
	if metric != "" {
		if err := DeleteMetric(ctx, metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}
