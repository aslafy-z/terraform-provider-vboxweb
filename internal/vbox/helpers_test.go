package vbox

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestListToStrings_Nil(t *testing.T) {
	result := ListToStrings(types.ListNull(types.StringType))
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestListToStrings_Unknown(t *testing.T) {
	result := ListToStrings(types.ListUnknown(types.StringType))
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestListToStrings_Empty(t *testing.T) {
	list, diags := types.ListValue(types.StringType, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("failed to create list: %v", diags)
	}

	result := ListToStrings(list)
	if result == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestListToStrings_WithValues(t *testing.T) {
	list, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("one"),
		types.StringValue("two"),
		types.StringValue("three"),
	})
	if diags.HasError() {
		t.Fatalf("failed to create list: %v", diags)
	}

	result := ListToStrings(list)
	expected := []string{"one", "two", "three"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("expected element %d to be %q, got %q", i, v, result[i])
		}
	}
}

func TestListToStrings_WithNullValue(t *testing.T) {
	list, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("one"),
		types.StringNull(),
		types.StringValue("three"),
	})
	if diags.HasError() {
		t.Fatalf("failed to create list: %v", diags)
	}

	result := ListToStrings(list)
	expected := []string{"one", "three"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("expected element %d to be %q, got %q", i, v, result[i])
		}
	}
}

func TestListToStrings_WithUnknownValue(t *testing.T) {
	list, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("one"),
		types.StringUnknown(),
		types.StringValue("three"),
	})
	if diags.HasError() {
		t.Fatalf("failed to create list: %v", diags)
	}

	result := ListToStrings(list)
	expected := []string{"one", "three"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("expected element %d to be %q, got %q", i, v, result[i])
		}
	}
}

func TestMustString_Valid(t *testing.T) {
	result := mustString(types.StringValue("test"))
	if result != "test" {
		t.Errorf("expected 'test', got %q", result)
	}
}

func TestMustString_Null(t *testing.T) {
	result := mustString(types.StringNull())
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestMustString_Unknown(t *testing.T) {
	result := mustString(types.StringUnknown())
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestMustString_NonString(t *testing.T) {
	result := mustString(types.BoolValue(true))
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFmtAttr(t *testing.T) {
	result := fmtAttr(types.StringValue("test"))
	// The actual underlying type is basetypes.StringValue
	if result != "basetypes.StringValue" {
		t.Errorf("expected 'basetypes.StringValue', got %q", result)
	}
}
