package observe

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccObserveSourceDataset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "observe_workspace" "kubernetes" {
				  name = "Kubernetes"
				}

				data "observe_dataset" "observation" {
				  workspace = data.observe_workspace.kubernetes.id
				  name      = "Observation"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.observe_dataset.observation", "id"),
				),
			},
		},
	})
}
