package vbox

import (
  "fmt"

  "github.com/hashicorp/terraform-plugin-framework/attr"
  "github.com/hashicorp/terraform-plugin-framework/types"
)

// ListToStrings converts a terraform types.List of strings into a []string.
// The caller must ensure the list's element type is string.
func ListToStrings(v types.List) []string {
  if v.IsNull() || v.IsUnknown() {
    return nil
  }
  elems := v.Elements()
  out := make([]string, 0, len(elems))
  for _, e := range elems {
    av, ok := e.(attr.Value)
    if !ok {
      continue
    }
    sv, ok := av.(types.String)
    if !ok {
      // Defensive: ignore non-string
      continue
    }
    if sv.IsNull() || sv.IsUnknown() {
      continue
    }
    out = append(out, sv.ValueString())
  }
  return out
}

func mustString(v attr.Value) string {
  sv, ok := v.(types.String)
  if !ok {
    return ""
  }
  if sv.IsNull() || sv.IsUnknown() {
    return ""
  }
  return sv.ValueString()
}

func fmtAttr(v attr.Value) string {
  return fmt.Sprintf("%T", v)
}
