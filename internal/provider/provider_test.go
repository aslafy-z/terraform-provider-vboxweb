package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProviderMetadata(t *testing.T) {
	p := New()

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "vboxweb" {
		t.Errorf("expected TypeName 'vboxweb', got %q", resp.TypeName)
	}
}

func TestProviderSchema(t *testing.T) {
	p := New()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}

	schema := resp.Schema

	// Check endpoint attribute
	endpointAttr, ok := schema.Attributes["endpoint"]
	if !ok {
		t.Fatal("expected 'endpoint' attribute in schema")
	}
	if !endpointAttr.IsRequired() {
		t.Error("expected 'endpoint' attribute to be required")
	}

	// Check username attribute
	usernameAttr, ok := schema.Attributes["username"]
	if !ok {
		t.Fatal("expected 'username' attribute in schema")
	}
	if !usernameAttr.IsRequired() {
		t.Error("expected 'username' attribute to be required")
	}

	// Check password attribute
	passwordAttr, ok := schema.Attributes["password"]
	if !ok {
		t.Fatal("expected 'password' attribute in schema")
	}
	if !passwordAttr.IsRequired() {
		t.Error("expected 'password' attribute to be required")
	}
	if !passwordAttr.IsSensitive() {
		t.Error("expected 'password' attribute to be sensitive")
	}
}

func TestProviderResources(t *testing.T) {
	p := New().(*vboxwebProvider)

	resources := p.Resources(context.Background())

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	// Verify the resource factory works
	resource := resources[0]()
	if resource == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestProviderDataSources(t *testing.T) {
	p := New().(*vboxwebProvider)

	dataSources := p.DataSources(context.Background())

	if dataSources != nil && len(dataSources) != 0 {
		t.Errorf("expected no data sources, got %d", len(dataSources))
	}
}

// testAccProtoV6ProviderFactories creates provider factories for acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"vboxweb": providerserver.NewProtocol6WithError(New()),
}

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}

	_, ok := p.(*vboxwebProvider)
	if !ok {
		t.Error("expected provider to be *vboxwebProvider")
	}
}
