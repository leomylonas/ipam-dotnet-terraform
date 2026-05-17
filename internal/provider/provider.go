package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ipamProvider{}

type ipamProvider struct {
	version string
}

type ipamProviderModel struct {
	BaseURL         types.String `tfsdk:"base_url"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	TimeoutSeconds  types.Int64  `tfsdk:"timeout_seconds"`
	InsecureSkipTLS types.Bool   `tfsdk:"insecure_skip_tls_verify"`
}

type providerData struct {
	client *apiClient
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ipamProvider{version: version}
	}
}

func (p *ipamProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ipam"
	resp.Version = p.version
}

func (p *ipamProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url":                 schema.StringAttribute{Required: true},
			"username":                 schema.StringAttribute{Required: true},
			"password":                 schema.StringAttribute{Required: true, Sensitive: true},
			"timeout_seconds":          schema.Int64Attribute{Optional: true},
			"insecure_skip_tls_verify": schema.BoolAttribute{Optional: true},
		},
	}
}

func (p *ipamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ipamProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.TimeoutSeconds.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("timeout_seconds"), "Unknown timeout", "timeout_seconds must be known")
		return
	}

	timeout := 30 * time.Second
	if !config.TimeoutSeconds.IsNull() {
		timeout = time.Duration(config.TimeoutSeconds.ValueInt64()) * time.Second
	}
	insecure := false
	if !config.InsecureSkipTLS.IsNull() {
		insecure = config.InsecureSkipTLS.ValueBool()
	}

	client, err := newAPIClient(config.BaseURL.ValueString(), config.Username.ValueString(), config.Password.ValueString(), timeout, insecure)
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure IPAM client", err.Error())
		return
	}

	data := &providerData{client: client}
	resp.ResourceData = data
	resp.DataSourceData = data
}

func (p *ipamProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTenancyResource,
		NewUserResource,
		NewPrivateSubnetResource,
		NewSharedSubnetResource,
		NewSharedSubnetAccessResource,
		NewExclusionResource,
		NewAllocationResource,
		NewAllocationTagsResource,
	}
}

func (p *ipamProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTenancyDataSource,
		NewUsersDataSource,
		NewSharedSubnetsDataSource,
		NewPrivateSubnetsDataSource,
		NewAllocationDataSource,
		NewAllocationsDataSource,
		NewSubnetStatsDataSource,
	}
}

func providerDataFrom(v any) (*providerData, error) {
	pd, ok := v.(*providerData)
	if !ok || pd == nil || pd.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}
	return pd, nil
}
