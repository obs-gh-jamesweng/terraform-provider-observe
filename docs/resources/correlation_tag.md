---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "observe_correlation_tag Resource - terraform-provider-observe"
subcategory: ""
description: |-
  A correlation tag can be attached to columns of a dataset. These tags are later used to correlate multiple datasets.
---
# observe_correlation_tag

A correlation tag can be attached to columns of a dataset. These tags are later used to correlate multiple datasets.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `column` (String) The column to which the correlation tag should be attached.
- `dataset` (String) OID of the dataset to which the correlation tag should be attached.
- `name` (String) The name to attach.

### Optional

- `path` (String) If the column is of type "object", a correlation tag can be attached to a
key nested within the object. Standard Javascript notation can be used to specify the path to the key.
For example, say the object has the following structure -
{
  "a": {
    "b": {
      "c": "value"
    }
  }
}
Then the path to the key "c" would be "a.b.c" or "a['b']['c']"

### Read-Only

- `id` (String) The ID of this resource.
