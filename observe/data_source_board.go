package observe

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	observe "github.com/observeinc/terraform-provider-observe/client"
)

func dataSourceBoard() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBoardRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// computed
			"oid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dataset": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBoardRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		client = meta.(*observe.Client)
		id     = data.Get("id").(string)
	)

	board, err := client.GetBoard(ctx, id)
	if err != nil {
		err = fmt.Errorf("failed to retrieve board %q: %w", id, err)
		return diag.FromErr(err)
	}

	data.SetId(board.ID)
	return boardToResourceData(board, data)
}