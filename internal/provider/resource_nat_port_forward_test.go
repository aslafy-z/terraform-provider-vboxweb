package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNatPortForwardResourceMetadata(t *testing.T) {
	r := NewNatPortForwardResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "vboxweb",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "vboxweb_nat_port_forward" {
		t.Errorf("expected TypeName 'vboxweb_nat_port_forward', got %q", resp.TypeName)
	}
}

func TestNatPortForwardResourceSchema(t *testing.T) {
	r := NewNatPortForwardResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	schema := resp.Schema

	// Check required attributes
	requiredAttrs := []string{"machine_id", "adapter_slot", "name", "protocol", "guest_port"}
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
	computedOnlyAttrs := []string{"id", "effective_host_port"}
	for _, attrName := range computedOnlyAttrs {
		attr, ok := schema.Attributes[attrName]
		if !ok {
			t.Errorf("expected %q attribute in schema", attrName)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected %q attribute to be computed", attrName)
		}
	}

	// Check optional attributes with defaults
	optionalWithDefaults := []string{"host_ip", "guest_ip", "auto_host_port", "auto_host_port_min", "auto_host_port_max", "auto_host_ip_scope"}
	for _, attrName := range optionalWithDefaults {
		attr, ok := schema.Attributes[attrName]
		if !ok {
			t.Errorf("expected %q attribute in schema", attrName)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected %q attribute to be optional", attrName)
		}
	}

	// Check host_port is optional (can be auto-allocated)
	hostPortAttr, ok := schema.Attributes["host_port"]
	if !ok {
		t.Fatal("expected 'host_port' attribute in schema")
	}
	if !hostPortAttr.IsOptional() {
		t.Error("expected 'host_port' attribute to be optional")
	}
}

func TestNatPortForwardResourceConfigure_NilProviderData(t *testing.T) {
	r := &natPortForwardResource{}

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
