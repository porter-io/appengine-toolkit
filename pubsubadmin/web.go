package pubsubadmin

import (
	"github.com/gorilla/mux"
	"google.golang.org/api/pubsub/v1beta2"
	"google.golang.org/appengine"
	"html/template"
	"net/http"
)

func RegisterHandler(r *mux.Router) {
	r.HandleFunc("/", HandleIndex)
	r.HandleFunc("/topic/create/", HandleCreateTopic)
	r.HandleFunc("/topic/delete/", HandleDeleteTopic)
	r.HandleFunc("/subscription/create/", HandleCreateSubscription)
	r.HandleFunc("/subscription/delete/", HandleDeleteSubscription)
}

const pageHTML = `
<html>
  <body>
    <h1>Topics</h1>
    {{ range .Topics }}
	  <p>
	  {{ .Name }}
	  <form action="topic/delete/" method="post">
	  	<input type="hidden" name="topic" value="{{ .Name }}">
	  	<input type="submit" value="Delete">
	  </form>
	  </p>
	{{ end }}
    <h1>Subscriptions</h1>
    {{ range .Subscriptions }}
	  <h2>{{ .Name }}</h2>
	  <ul>
		  <li>Topic: {{ .Topic }}</li>
		  <li>Endpoint: {{ .PushConfig.PushEndpoint }}</li>
	  </ul>
	  <form action="subscription/delete/" method="post">
	  	<input type="hidden" name="subscription" value="{{ .Name }}">
	  	<input type="submit" value="Delete">
	  </form>
	{{ end }}
	<h1>Add Topic</h1>
	<form action="topic/create/" method="post">
		<input type="text" name="topic">
		<input type="submit">
	</form>
	<h1>Add Subscription</h1>
	<form action="subscription/create/" method="post">
		<select name="topic">
		{{ range .Topics }}
		  <option value="{{ .Name }}">{{ .Name }}</option>
		{{ end}}
		</select>
		<input type="text" name="name">
		<input type="text" name="endpoint">
		<input type="submit">
	</form>
  </body>
</html>
`

var pageTemplate = template.Must(template.New("page").Parse(pageHTML))

type pageContext struct {
	Topics        []*pubsub.Topic
	Subscriptions []*pubsub.Subscription
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	topics, err := ListTopic(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	subscriptions, err := ListSubscription(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = pageTemplate.Execute(w, pageContext{
		Topics:        topics,
		Subscriptions: subscriptions,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleCreateTopic(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("topic")
	if topic != "" {
		if err := CreateTopic(appengine.NewContext(r), topic); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}

func HandleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("topic")
	if topic != "" {
		if err := DeleteTopic(appengine.NewContext(r), topic); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}

func HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("topic")
	name := r.FormValue("name")
	endpoint := r.FormValue("endpoint")
	if topic != "" && name != "" && endpoint != "" {
		if err := CreateSubscription(appengine.NewContext(r), topic, name, endpoint); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}

func HandleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	topic := r.FormValue("subscription")
	if topic != "" {
		if err := DeleteSubscription(appengine.NewContext(r), topic); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "../../", 301)
}
