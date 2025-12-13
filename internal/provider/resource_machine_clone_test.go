package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestMachineCloneResourceMetadata(t *testing.T) {
	r := NewMachineCloneResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "vboxweb",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "vboxweb_machine" {
		t.Errorf("expected TypeName 'vboxweb_machine', got %q", resp.TypeName)
	}
}

func TestMachineCloneResourceSchema(t *testing.T) {
	r := NewMachineCloneResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	schema := resp.Schema

	// Check required attributes
	requiredAttrs := []string{"name", "source"}
	for _, attrName := range requiredAttrs {
		attr, ok := schema.Attributes[attrName]
		if !ok {
			t.Errorf("expected %q attribute in schema", attrName)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected %q attribute to be required", attrName)
		}
	}

	// Check computed attributes
	computedAttrs := []string{"id", "current_state"}
	for _, attrName := range computedAttrs {
		attr, ok := schema.Attributes[attrName]
		if !ok {
			t.Errorf("expected %q attribute in schema", attrName)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected %q attribute to be computed", attrName)
		}
	}

	// Check optional/computed attributes
	optionalComputedAttrs := []string{"clone_mode", "state", "session_type", "wait_timeout"}
	for _, attrName := range optionalComputedAttrs {
		attr, ok := schema.Attributes[attrName]
		if !ok {
			t.Errorf("expected %q attribute in schema", attrName)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected %q attribute to be optional", attrName)
		}
		if !attr.IsComputed() {
			t.Errorf("expected %q attribute to be computed", attrName)
		}
	}

	// Check clone_options is optional list
	cloneOptionsAttr, ok := schema.Attributes["clone_options"]
	if !ok {
		t.Fatal("expected 'clone_options' attribute in schema")
	}
	if !cloneOptionsAttr.IsOptional() {
		t.Error("expected 'clone_options' attribute to be optional")
	}
}

func TestNormalizeDesiredState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"started", "started"},
		{"STARTED", "started"},
		{"Started", "started"},
		{"running", "started"},
		{"Running", "started"},
		{"on", "started"},
		{"ON", "started"},
		{"stopped", "stopped"},
		{"STOPPED", "stopped"},
		{"Stopped", "stopped"},
		{"poweredoff", "stopped"},
		{"PoweredOff", "stopped"},
		{"powered_off", "stopped"},
		{"off", "stopped"},
		{"OFF", "stopped"},
		{"  started  ", "started"},
		{"  stopped  ", "stopped"},
		{"unknown", "unknown"},
		{"", ""},
		{"paused", "paused"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeDesiredState(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeDesiredState(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseTimeout(t *testing.T) {
	defaultTimeout := 20 * time.Minute

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"empty string", "", defaultTimeout},
		{"whitespace only", "   ", defaultTimeout},
		{"valid 5m", "5m", 5 * time.Minute},
		{"valid 1h", "1h", 1 * time.Hour},
		{"valid 30s", "30s", 30 * time.Second},
		{"valid 1h30m", "1h30m", 1*time.Hour + 30*time.Minute},
		{"invalid format", "invalid", defaultTimeout},
		{"negative", "-5m", defaultTimeout},
		{"zero", "0s", defaultTimeout},
		{"valid 100ms", "100ms", 100 * time.Millisecond},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTimeout(tc.input)
			if result != tc.expected {
				t.Errorf("parseTimeout(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMachineCloneResourceConfigure_NilProviderData(t *testing.T) {
	r := &machineCloneResource{}

	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected errors: %v", resp.Diagnostics)
	}

	if r.client != nil {
		t.Error("expected client to be nil when ProviderData is nil")
	}
}

func TestNewMachineCloneResource(t *testing.T) {
	r := NewMachineCloneResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	_, ok := r.(*machineCloneResource)
	if !ok {
		t.Error("expected resource to be *machineCloneResource")
	}
}
