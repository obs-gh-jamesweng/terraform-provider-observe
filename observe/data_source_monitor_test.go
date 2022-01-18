package observe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccObserveSourceMonitor(t *testing.T) {
	randomPrefix := acctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPreamble+`
					resource "observe_monitor" "first" {
						workspace = data.observe_workspace.default.oid
						name      = "%s"
						disabled  = true

						inputs = {
							"observation" = data.observe_dataset.observation.oid
						}

						stage {}

						rule {
							group_by      = "none"

							count {
								compare_function   = "less_or_equal"
								compare_values     = [1]
								lookback_time      = "1m"
							}
						}

						notification_spec {
							selection       = "count"
							selection_value = 1
						}
					}

					data "observe_monitor" "lookup" {
						workspace  = data.observe_workspace.default.oid
						id         = observe_monitor.first.id
					}
				`, randomPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "name", randomPrefix),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "disabled", "true"),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "stage.0.pipeline", ""),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection", "count"),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection_value", "1"),
				),
			},
			{
				Config: fmt.Sprintf(configPreamble+`
					resource "observe_monitor" "first" {
						workspace = data.observe_workspace.default.oid
						name      = "%s"

						inputs = {
							"observation" = data.observe_dataset.observation.oid
						}

						stage {
							pipeline = <<-EOF
								filter false
							EOF
						}

						rule {
							group_by      = "none"

							count {
								compare_function   = "less_or_equal"
								compare_values     = [1]
								lookback_time      = "1m"
							}
						}

						notification_spec {
							selection       = "count"
							selection_value = 1
						}
					}

					data "observe_monitor" "lookup" {
						workspace  = data.observe_workspace.default.oid
						id         = observe_monitor.first.id
					}
					`, randomPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "name", randomPrefix),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "stage.0.pipeline", "filter false\n"),
				),
			},
		},
	})
}

func TestAccObserveSourceMonitorLookup(t *testing.T) {
	randomPrefix := acctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPreamble+`
					resource "observe_monitor" "first" {
						workspace = data.observe_workspace.default.oid
						name      = "%[1]s"

						inputs = {
							"observation" = data.observe_dataset.observation.oid
						}

						stage {}

						rule {
							group_by      = "none"

							count {
								compare_function   = "less_or_equal"
								compare_values     = [1]
								lookback_time      = "1m"
							}
						}

						notification_spec {
							selection       = "count"
							selection_value = 1
						}
					}

					data "observe_monitor" "lookup" {
					    depends_on = [observe_monitor.first]
						workspace  = data.observe_workspace.default.oid
						name       = "%[1]s"
					}
				`, randomPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "name", randomPrefix),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "stage.0.pipeline", ""),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection", "count"),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection_value", "1"),
				),
			},
			{
				Config: fmt.Sprintf(configPreamble+`
					resource "observe_monitor" "first" {
						workspace = data.observe_workspace.default.oid
						name      = "%[1]s"

						inputs = {
							"observation" = data.observe_dataset.observation.oid
						}

						stage {}

						rule {
							group_by      = "none"
							source_column = "OBSERVATION_INDEX"

							threshold {
								compare_function = "greater"
								compare_values = [ 75, ]
								lookback_time = "5m0s"
							}
						}

						notification_spec {
							importance      = "informational"
							selection       = "any"
							selection_value = 50
						}
					}

					data "observe_monitor" "lookup" {
					    depends_on = [observe_monitor.first]
						workspace  = data.observe_workspace.default.oid
						name       = "%[1]s"
					}
				`, randomPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "name", randomPrefix),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "stage.0.pipeline", ""),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection", "any"),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "notification_spec.0.selection_value", "50"),
					resource.TestCheckResourceAttr("data.observe_monitor.lookup", "rule.0.threshold.0.compare_function", "greater"),
				),
			},
		},
	})
}
