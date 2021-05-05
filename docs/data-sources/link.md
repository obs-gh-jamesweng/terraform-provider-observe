---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "observe_link Data Source - terraform-provider-observe"
subcategory: ""
description: |-
  
---

# observe_link (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **fields** (List of String) Array of field mappings that provides a link between source and target datasets. A mapping between a `source_field` and a `target_field` is represented using a colon separated "<source_field>:<target_field>" format. If the source and target field share the same name, only "<source_field>".
- **source** (String) OID of source dataset.
- **target** (String) OID of target dataset.

### Optional

- **id** (String) The ID of this resource.

