package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox"
)

type vboxwebProvider struct{}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func New() provider.Provider {
	return &vboxwebProvider{}
}

func (p *vboxwebProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vboxweb"
}

func (p *vboxwebProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "vboxwebsrv endpoint, for example http://host:18083/",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "VirtualBox webservice username.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "VirtualBox webservice password.",
			},
		},
	}
}

func (p *vboxwebProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := vbox.NewClient(cfg.Endpoint.ValueString(), cfg.Username.ValueString(), cfg.Password.ValueString())
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *vboxwebProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMachineCloneResource,
	}
}

func (p *vboxwebProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
