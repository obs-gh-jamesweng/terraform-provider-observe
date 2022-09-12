---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "observe_dashboard Resource - terraform-provider-observe"
subcategory: ""
description: |-
  
---
# observe_dashboard



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Dashboard name. Must be unique within workspace.
- `stages` (String) Dashboard stages in JSON format.
- `workspace` (String) OID of workspace dashboard is contained in.

### Optional

- `icon_url` (String) Icon image.
- `layout` (String) Dashboard layout in JSON format.
- `parameter_values` (String) Dashboard parameter values in JSON format.
- `parameters` (String) Dashboard parameters in JSON format.

### Read-Only

- `id` (String) The ID of this resource.
- `oid` (String) The Observe ID for dashboard.
