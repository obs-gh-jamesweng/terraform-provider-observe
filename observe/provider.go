package observe

import (
	"context"
	"time"

	"github.com/observeinc/terraform-provider-observe/version"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns observe terraform provider
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"customer": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OBSERVE_CUSTOMER", nil),
				Description: "Observe API URL",
			},
			"token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("OBSERVE_TOKEN", nil),
				Description:   "Observe Token",
				ConflictsWith: []string{"user_email", "user_password"},
			},
			"user_email": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("OBSERVE_USER_EMAIL", nil),
				Description:  "Observe User Email",
				RequiredWith: []string{"user_password"},
			},
			"user_password": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("OBSERVE_USER_PASSWORD", nil),
				Description:  "Observe User Password",
				RequiredWith: []string{"user_email"},
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OBSERVE_DOMAIN", "observeinc.com"),
				Description: "Observe root domain",
			},
			"insecure": {
				Type:        schema.TypeBool,
				DefaultFunc: schema.EnvDefaultFunc("OBSERVE_INSECURE", false),
				Optional:    true,
				Description: "Skip TLS verification",
			},
			"retry_count": {
				Type:        schema.TypeInt,
				Default:     3,
				Optional:    true,
				Description: "Maximum number of retries on temporary network failures",
			},
			"retry_wait": {
				Type:             schema.TypeString,
				Default:          "3s",
				Optional:         true,
				ValidateDiagFunc: validateTimeDuration,
				DiffSuppressFunc: diffSuppressTimeDuration,
				Description:      "Time between retries",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"observe_dataset":   dataSourceDataset(),
			"observe_fk":        dataSourceForeignKey(),
			"observe_workspace": dataSourceWorkspace(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"observe_dataset":   resourceDataset(),
			"observe_fk":        resourceForeignKey(),
			"observe_workspace": resourceWorkspace(),
		},
	}

	// this is a bit circular: we need a client for the provider, but we need
	// userAgent to create the client, and we need the provider to get
	// userAgent. So we create provider, grab userAgent, and finally attach the
	// ConfigureContextFunc.
	userAgent := provider.UserAgent("terraform-provider-observe", version.ProviderVersion)
	provider.ConfigureContextFunc = getConfigureContextFunc(userAgent)
	return provider
}

func getConfigureContextFunc(userAgent string) schema.ConfigureContextFunc {
	return func(ctx context.Context, data *schema.ResourceData) (client interface{}, diags diag.Diagnostics) {
		c := Config{
			CustomerID:   data.Get("customer").(string),
			Token:        data.Get("token").(string),
			UserEmail:    data.Get("user_email").(string),
			UserPassword: data.Get("user_password").(string),
			Domain:       data.Get("domain").(string),
			Insecure:     data.Get("insecure").(bool),
			RetryCount:   data.Get("retry_count").(int),
			UserAgent:    userAgent,
		}

		if retryWait := data.Get("retry_wait").(string); retryWait != "" {
			// already validated format
			c.RetryWait, _ = time.ParseDuration(retryWait)
		}

		if c.Insecure {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Insecure API session",
			})
		}

		client, err := c.Client()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create client",
				Detail:   err.Error(),
			})
			return nil, diags
		}
		return client, diags
	}
}
