package observe

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccObserveApp(t *testing.T) {
	randomPrefix := acctest.RandomWithPrefix("tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configPreamble+`
				resource "observe_folder" "example" {
				  workspace = data.observe_workspace.default.oid
				  name      = "%[1]s"
				}

				resource "observe_app" "example" {
				  folder    = observe_folder.example.oid

				  module_id = "observeinc/example/observe"
				  version   = "0.1.0"

				  variables = {
					required_string = "ok"
				  }
				}`, randomPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("observe_app.example", "module_id", "observeinc/example/observe"),
					resource.TestCheckResourceAttr("observe_app.example", "version", "0.1.0"),
				),
			},
			/*
				{
					Config: fmt.Sprintf(configPreamble+`
						resource "observe_folder" "example" {
						  workspace = data.observe_workspace.default.oid
						  name      = "%[1]s"
						}

						resource "observe_datastream" "example" {
						  workspace = data.observe_workspace.default.oid
						  name      = "%[1]s"
						}

						resource "observe_app" "example" {
						  folder    = observe_folder.example.oid

						  module_id = "observeinc/example/observe"
						  version   = "0.2.0"

						  variables = {
							datastream = observe_datastream.example.id
						  }
						}`, randomPrefix),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("observe_app.example", "module_id", "observeinc/example/observe"),
						resource.TestCheckResourceAttr("observe_app.example", "version", "0.2.0"),
					),
				},
			*/
		},
	})
}