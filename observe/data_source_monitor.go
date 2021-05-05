package observe

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	observe "github.com/observeinc/terraform-provider-observe/client"
)

func dataSourceMonitor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMonitorRead,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateOID(observe.TypeWorkspace),
				Description:      schemaDatasetWorkspaceDescription,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// computed values
			"oid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: schemaMonitorOIDDescription,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: schemaMonitorNameDescription,
			},
			"icon_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: schemaMonitorIconDescription,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: schemaMonitorDescriptionDescription,
			},
			"inputs": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: schemaDatasetInputsDescription,
			},
			"stage": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				// we need to declare optional, otherwise we won't get block
				// formatting in state
				Optional:    true,
				Description: schemaDatasetStageDescription,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: schemaDatasetStageAliasDescription,
						},
						"input": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: schemaDatasetStageInputDescription,
						},
						"pipeline": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: schemaDatasetStagePipelineDescription,
						},
					},
				},
			},
			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_column": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_by_columns": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"count": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compare_function": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"compare_values": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeFloat},
									},
									"lookback_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"change": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"change_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"compare_function": {
										Type:     schema.TypeString,
										Required: true,
									},
									"aggregate_function": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"compare_values": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeFloat},
									},
									"lookback_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"baseline_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"facet": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"facet_function": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"facet_values": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"time_function": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"time_value": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"lookback_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"promote": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary_key": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"kind_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"notification_spec": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"importance": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"merge": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"selection": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"selection_value": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceMonitorRead(ctx context.Context, data *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	data.SetId(data.Get("id").(string))
	return resourceMonitorRead(ctx, data, meta)
}