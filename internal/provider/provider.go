package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	RetryWaitMinMs  types.Int64  `tfsdk:"retry_wait_min_ms"`
	RetryWaitMaxMs  types.Int64  `tfsdk:"retry_wait_max_ms"`
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
			"base_url":                 schema.StringAttribute{Optional: true, Description: "Base URL of the dotnet-ipam API. Can also be set via IPAM_BASE_URL."},
			"username":                 schema.StringAttribute{Optional: true, Description: "Username for Basic Auth. Can also be set via IPAM_USERNAME."},
			"password":                 schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password for Basic Auth. Can also be set via IPAM_PASSWORD."},
			"timeout_seconds":          schema.Int64Attribute{Optional: true, Description: "HTTP request timeout in seconds. Defaults to 30. Can also be set via IPAM_TIMEOUT_SECONDS."},
			"insecure_skip_tls_verify": schema.BoolAttribute{Optional: true, Description: "Skip TLS certificate verification. Defaults to false. Can also be set via IPAM_INSECURE_SKIP_TLS_VERIFY."},
			"max_retries":              schema.Int64Attribute{Optional: true, Description: "Maximum retries for transient 5xx/429 responses. Defaults to 2. Can also be set via IPAM_MAX_RETRIES."},
			"retry_wait_min_ms":        schema.Int64Attribute{Optional: true, Description: "Minimum wait between retries in milliseconds. Defaults to 200. Can also be set via IPAM_RETRY_WAIT_MIN_MS."},
			"retry_wait_max_ms":        schema.Int64Attribute{Optional: true, Description: "Maximum wait between retries in milliseconds. Defaults to 2000. Can also be set via IPAM_RETRY_WAIT_MAX_MS."},
		},
	}
}

func (p *ipamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ipamProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// During terraform validate (and some plan paths), provider configuration
	// values can remain unknown. In that mode we should not fail configuration;
	// install a placeholder client so schema/config validation can proceed.
	if config.BaseURL.IsUnknown() || config.Username.IsUnknown() || config.Password.IsUnknown() {
		placeholder, _ := newAPIClient(
			"http://127.0.0.1",
			"placeholder",
			"placeholder",
			30*time.Second,
			false,
			0,
			200*time.Millisecond,
			200*time.Millisecond,
		)
		data := &providerData{client: placeholder}
		resp.ResourceData = data
		resp.DataSourceData = data
		return
	}

	baseURL := config.BaseURL.ValueString()
	if baseURL == "" {
		baseURL = os.Getenv("IPAM_BASE_URL")
	}
	username := config.Username.ValueString()
	if username == "" {
		username = os.Getenv("IPAM_USERNAME")
	}
	password := config.Password.ValueString()
	if password == "" {
		password = os.Getenv("IPAM_PASSWORD")
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Missing base_url", "Set base_url in the provider block or via IPAM_BASE_URL.")
	}
	if username == "" {
		resp.Diagnostics.AddAttributeError(path.Root("username"), "Missing username", "Set username in the provider block or via IPAM_USERNAME.")
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Missing password", "Set password in the provider block or via IPAM_PASSWORD.")
	}
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
	} else if v := os.Getenv("IPAM_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			timeout = time.Duration(n) * time.Second
		}
	}

	maxRetries := int64(2)
	if !config.MaxRetries.IsNull() {
		maxRetries = config.MaxRetries.ValueInt64()
	} else if v := os.Getenv("IPAM_MAX_RETRIES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxRetries = n
		}
	}

	retryWaitMin := int64(200)
	if !config.RetryWaitMinMs.IsNull() {
		retryWaitMin = config.RetryWaitMinMs.ValueInt64()
	} else if v := os.Getenv("IPAM_RETRY_WAIT_MIN_MS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			retryWaitMin = n
		}
	}

	retryWaitMax := int64(2000)
	if !config.RetryWaitMaxMs.IsNull() {
		retryWaitMax = config.RetryWaitMaxMs.ValueInt64()
	} else if v := os.Getenv("IPAM_RETRY_WAIT_MAX_MS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			retryWaitMax = n
		}
	}

	if maxRetries < 0 {
		resp.Diagnostics.AddAttributeError(path.Root("max_retries"), "Invalid max_retries", "max_retries must be >= 0")
		return
	}
	if retryWaitMin <= 0 || retryWaitMax <= 0 || retryWaitMin > retryWaitMax {
		resp.Diagnostics.AddError("Invalid retry settings", "retry_wait_min_ms and retry_wait_max_ms must be > 0 and min <= max")
		return
	}

	insecure := false
	if !config.InsecureSkipTLS.IsNull() {
		insecure = config.InsecureSkipTLS.ValueBool()
	} else if v := os.Getenv("IPAM_INSECURE_SKIP_TLS_VERIFY"); v != "" {
		insecure, _ = strconv.ParseBool(v)
	}

	client, err := newAPIClient(
		baseURL,
		username,
		password,
		timeout,
		insecure,
		int(maxRetries),
		time.Duration(retryWaitMin)*time.Millisecond,
		time.Duration(retryWaitMax)*time.Millisecond,
	)
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
		NewBulkAllocationResource,
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
		NewBulkAllocationsDataSource,
		NewSubnetStatsDataSource,
	}
}

func providerDataFrom(v any) (*providerData, error) {
	if v == nil {
		placeholder, _ := newAPIClient(
			"http://127.0.0.1",
			"placeholder",
			"placeholder",
			30*time.Second,
			false,
			0,
			200*time.Millisecond,
			200*time.Millisecond,
		)
		return &providerData{client: placeholder}, nil
	}

	pd, ok := v.(*providerData)
	if !ok || pd == nil || pd.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}
	return pd, nil
}
