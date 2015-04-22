package monitoringadmin

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudmonitoring/v2beta2"
	"google.golang.org/appengine"
	"strings"
)

func NewMonitoringService(c context.Context) (s *cloudmonitoring.Service, err error) {
	client, err := google.DefaultClient(c, cloudmonitoring.MonitoringScope)
	if err != nil {
		return
	}
	s, err = cloudmonitoring.New(client)
	return
}

func CreateMetric(c context.Context, metric, description, metricType, valueType string) (err error) {
	service, err := NewMonitoringService(c)
	if err != nil {
		return
	}
	var full string
	if strings.HasPrefix(metric, "custom.cloudmonitoring.googleapis.com/") {
		full = metric
	} else {
		full = fmt.Sprintf("custom.cloudmonitoring.googleapis.com/%s", metric)
	}
	m := cloudmonitoring.MetricDescriptor{
		Name:        full,
		Description: description,
		TypeDescriptor: &cloudmonitoring.MetricDescriptorTypeDescriptor{
			MetricType: metricType,
			ValueType:  valueType,
		},
	}
	_, err = service.MetricDescriptors.Create(appengine.AppID(c), &m).Do()
	return
}

func DeleteMetric(c context.Context, metric string) (err error) {
	service, err := NewMonitoringService(c)
	if err != nil {
		return
	}
	_, err = service.MetricDescriptors.Delete(appengine.AppID(c), metric).Do()
	return
}

func ListMetric(c context.Context) (metrics []*cloudmonitoring.MetricDescriptor, err error) {
	service, err := NewMonitoringService(c)
	if err != nil {
		return
	}
	r := cloudmonitoring.ListMetricDescriptorsRequest{Kind: "cloudmonitoring#listMetricDescriptorsRequest"}
	resp, err := service.MetricDescriptors.List(appengine.AppID(c), &r).Do()
	if err != nil {
		return
	}
	metrics = resp.Metrics
	return
}
