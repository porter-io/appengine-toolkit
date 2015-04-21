package pubsubadmin

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/pubsub/v1beta2"
	"google.golang.org/appengine"
	"regexp"
)

func NewPubsubService(c context.Context) (s *pubsub.Service, err error) {
	client, err := google.DefaultClient(c, pubsub.CloudPlatformScope)
	if err != nil {
		return
	}
	s, err = pubsub.New(client)
	return
}

func CreateTopic(c context.Context, topic string) (err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullTopic(c, topic)
	_, err = service.Projects.Topics.Create(full, &pubsub.Topic{}).Do()
	return
}

func DeleteTopic(c context.Context, topic string) (err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullTopic(c, topic)
	_, err = service.Projects.Topics.Delete(full).Do()
	return
}

func ListTopic(c context.Context) (topics []*pubsub.Topic, err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullProject(c)
	resp, err := service.Projects.Topics.List(full).Do()
	if err != nil {
		return
	}
	topics = resp.Topics
	return
}

func ListTopicSubscription(c context.Context, topic string) (subscriptions []string, err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullTopic(c, topic)
	resp, err := service.Projects.Topics.Subscriptions.List(full).Do()
	if err != nil {
		return
	}
	subscriptions = resp.Subscriptions
	return
}

func CreateSubscription(c context.Context, topic, name, endpoint string) (err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	sTopic := fullTopic(c, topic)
	sSubscription := fullSubscription(c, name)
	s := pubsub.Subscription{
		PushConfig: &pubsub.PushConfig{
			PushEndpoint: endpoint,
		},
		Topic: sTopic,
	}
	_, err = service.Projects.Subscriptions.Create(sSubscription, &s).Do()
	return
}

func DeleteSubscription(c context.Context, subscription string) (err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullSubscription(c, subscription)
	_, err = service.Projects.Subscriptions.Delete(full).Do()
	return
}

func ListSubscription(c context.Context) (subscriptions []*pubsub.Subscription, err error) {
	service, err := NewPubsubService(c)
	if err != nil {
		return
	}
	full := fullProject(c)
	resp, err := service.Projects.Subscriptions.List(full).Do()
	if err != nil {
		return
	}
	subscriptions = resp.Subscriptions
	return
}

func fullProject(c context.Context) string {
	return fmt.Sprintf("projects/%s", appengine.AppID(c))
}

func fullTopic(c context.Context, s string) string {
	re, _ := regexp.Compile(`projects/([\w-]+)/topics/([\w-]+)`)
	if re.MatchString(s) {
		return s
	} else {
		return fmt.Sprintf("projects/%s/topics/%s", appengine.AppID(c), s)
	}
}

func fullSubscription(c context.Context, s string) string {
	re, _ := regexp.Compile(`projects/([\w-]+)/subscriptions/([\w-]+)`)
	if re.MatchString(s) {
		return s
	} else {
		return fmt.Sprintf("projects/%s/subscriptions/%s", appengine.AppID(c), s)
	}
}
