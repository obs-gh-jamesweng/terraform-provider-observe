package observe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccObservePoller(t *testing.T) {
	randomPrefix := acctest.RandomWithPrefix("tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPreamble+`
				resource "observe_poller" "first" {
					workspace = data.observe_workspace.kubernetes.oid
					name      = "%s-%s"
					interval  = "1m"
					retries   = 5

					chunk {
					    enabled = true
						size = 1024
					}
					tags = {
						"k1"   = "v1"
						"k2"   = "v2"
					}
					http {
					    endpoint = "https://test.com"
						content_type = "application/json"
						headers = {
						    "token" = "test-token"
						}
					}
				}`, randomPrefix, "http"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("observe_poller.first", "name", randomPrefix+"-http"),
					resource.TestCheckResourceAttr("observe_poller.first", "interval", "1m0s"),
					resource.TestCheckResourceAttr("observe_poller.first", "retries", "5"),
					resource.TestCheckResourceAttr("observe_poller.first", "tags.k1", "v1"),
					resource.TestCheckResourceAttr("observe_poller.first", "tags.k2", "v2"),
					resource.TestCheckResourceAttr("observe_poller.first", "chunk.0.enabled", "true"),
					resource.TestCheckResourceAttr("observe_poller.first", "chunk.0.size", "1024"),
					resource.TestCheckResourceAttr("observe_poller.first", "http.0.endpoint", "https://test.com"),
					resource.TestCheckResourceAttr("observe_poller.first", "http.0.content_type", "application/json"),
					resource.TestCheckResourceAttr("observe_poller.first", "http.0.headers.token", "test-token"),
					resource.TestCheckResourceAttr("observe_poller.first", "pubsub.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(configPreamble+`
				resource "observe_poller" "second" {
					workspace = data.observe_workspace.kubernetes.oid
					name      = "%s-%s"
					interval  = "1m"
					retries   = 5

					chunk {
					    enabled = true
						size = 1024
					}
					tags = {
						"k1"   = "v1"
						"k2"   = "v2"
					}
					pubsub {
					    project_id = "gcp-test"
					    subscription_id = "sub-test"
						json_key = jsonencode({
							type: "service_account",
							project_id: "gcp-test"
						})
					}
				}`, randomPrefix, "pubsub"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("observe_poller.second", "name", randomPrefix+"-pubsub"),
					resource.TestCheckResourceAttr("observe_poller.second", "interval", "1m0s"),
					resource.TestCheckResourceAttr("observe_poller.second", "retries", "5"),
					resource.TestCheckResourceAttr("observe_poller.second", "tags.k1", "v1"),
					resource.TestCheckResourceAttr("observe_poller.second", "tags.k2", "v2"),
					resource.TestCheckResourceAttr("observe_poller.second", "chunk.0.enabled", "true"),
					resource.TestCheckResourceAttr("observe_poller.second", "chunk.0.size", "1024"),
					resource.TestCheckResourceAttr("observe_poller.second", "pubsub.0.project_id", "gcp-test"),
					resource.TestCheckResourceAttr("observe_poller.second", "pubsub.0.subscription_id", "sub-test"),
					resource.TestCheckResourceAttrSet("observe_poller.second", "pubsub.0.json_key"),
					resource.TestCheckResourceAttr("observe_poller.second", "http.#", "0"),
				),
			},
		},
	})
}
