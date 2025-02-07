/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

import (
	"encoding/json"
	"fmt"
	"strings"

	promv1a1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func convertToAPIExtV1JSON(val interface{}) apiextensionsv1.JSON {
	raw, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}

	out := apiextensionsv1.JSON{}
	out.Raw = raw

	return out
}

var _false = false

var pagerdutyAlerts = []string{
	"CacheBucketErrorState",
}

var smtpAlerts = []string{
	"DataSourceErrorState",
	"BucketPolicyErrorState",
}

// OSD Full alerts are silenced as there is no scenario in our static deployment configuration where an OSD is getting
// full without the cluster getting full.

var AlertmanagerConfigTemplate = promv1a1.AlertmanagerConfig{
	Spec: promv1a1.AlertmanagerConfigSpec{
		Route: &promv1a1.Route{
			Receiver: "null",
			Routes: []apiextensionsv1.JSON{
				convertToAPIExtV1JSON(promv1a1.Route{
					GroupBy:        []string{"alertname"},
					GroupWait:      "30s",
					GroupInterval:  "5m",
					RepeatInterval: "12h",
					Matchers: []promv1a1.Matcher{
						{
							Name:      "alertname",
							Value:     getRegexMatcher(smtpAlerts),
							MatchType: promv1a1.MatchRegexp,
						},
					},
					Receiver: "SendGrid",
				},
				),
				convertToAPIExtV1JSON(promv1a1.Route{
					GroupBy:        []string{"alertname"},
					GroupWait:      "30s",
					GroupInterval:  "5m",
					RepeatInterval: "12h",
					Matchers: []promv1a1.Matcher{
						{
							Name:      "alertname",
							Value:     getRegexMatcher(pagerdutyAlerts),
							MatchType: promv1a1.MatchRegexp,
						},
					},
					Receiver: "pagerduty",
				},
				),
				convertToAPIExtV1JSON(promv1a1.Route{
					GroupBy:        []string{"alertname"},
					GroupWait:      "30s",
					GroupInterval:  "5m",
					RepeatInterval: "5m",
					Matchers: []promv1a1.Matcher{
						{
							Name:      "alertname",
							Value:     "DeadMansSnitch",
							MatchType: promv1a1.MatchEqual,
						},
					},
					Receiver: "DeadMansSnitch",
				},
				),
			},
		},
		Receivers: []promv1a1.Receiver{
			{
				Name: "null",
			}, {
				Name: "pagerduty",
				PagerDutyConfigs: []promv1a1.PagerDutyConfig{{
					ServiceKey: &corev1.SecretKeySelector{Key: "", LocalObjectReference: corev1.LocalObjectReference{Name: ""}},
					Details:    []promv1a1.KeyValue{{Key: "", Value: ""}},
				}},
			}, {
				Name:           "DeadMansSnitch",
				WebhookConfigs: []promv1a1.WebhookConfig{{}},
			}, {
				Name: "SendGrid",
				EmailConfigs: []promv1a1.EmailConfig{
					{
						SendResolved: &_false,
						Smarthost:    "",
						From:         "",
						To:           "",
						AuthUsername: "",
						AuthPassword: &corev1.SecretKeySelector{Key: "", LocalObjectReference: corev1.LocalObjectReference{Name: ""}},
						Headers: []promv1a1.KeyValue{{
							Key: "subject",
							Value: "OpenShift Data Foundation Managed Service notification, " +
								"Action required on your managed OpenShift cluster!",
						}},
					},
				},
			},
		},
	},
}

func getRegexMatcher(alerts []string) string {
	return fmt.Sprintf("^%s$", strings.Join(alerts, "$|^"))
}
