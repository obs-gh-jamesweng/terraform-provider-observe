package observe

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	observe "github.com/observeinc/terraform-provider-observe/client"
)

// apply ValidateDiagFunc to every value in map
func validateMapValues(fn schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) (diags diag.Diagnostics) {
		for k, v := range i.(map[string]interface{}) {
			diags = append(diags, fn(v, path.IndexString(k))...)
		}
		return diags
	}
}

// Verify OID matches type
func validateOID(types ...observe.Type) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) (diags diag.Diagnostics) {
		oid, err := observe.NewOID(i.(string))
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				AttributePath: path,
			})
		}
		for _, t := range types {
			if oid.Type == t {
				return diags
			}
		}
		if len(types) > 0 {
			var s []string
			for _, t := range types {
				s = append(s, string(t))
			}
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "wrong type",
				Detail:        fmt.Sprintf("oid type must be %s", strings.Join(s, ", ")),
				AttributePath: path,
			})
		}
		return
	}
}

func validateTimeDuration(i interface{}, path cty.Path) diag.Diagnostics {
	s := i.(string)
	if _, err := time.ParseDuration(s); err != nil {
		return diag.Diagnostics{diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid field",
			Detail:        err.Error(),
			AttributePath: path,
		}}
	}
	return nil
}

// determine whether we expect a new dataset version
func datasetRecomputeOID(d *schema.ResourceDiff) bool {
	if len(d.GetChangedKeysPrefix("")) > 0 {
		return true
	}

	oid, err := observe.NewOID(d.Get("oid").(string))
	if err != nil || oid.Version == nil {
		return false
	}

	for _, v := range d.Get("inputs").(map[string]interface{}) {
		input, err := observe.NewOID(v.(string))
		if err == nil && input.Version != nil && *input.Version > *oid.Version {
			return true
		}
	}
	return false
}

func diffSuppressTimeDuration(k, old, new string, d *schema.ResourceData) bool {
	o, _ := time.ParseDuration(old)
	n, _ := time.ParseDuration(new)
	return o == n
}
