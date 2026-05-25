package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var subnetObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id": types.StringType, "cidr": types.StringType, "name": types.StringType,
	"description": types.StringType, "type": types.StringType,
	"tenancy_id": types.StringType, "created_at": types.StringType,
}}

var userObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id": types.StringType, "username": types.StringType, "role": types.StringType, "tenancy_id": types.StringType,
}}

var allocationObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id": types.StringType, "subnet_id": types.StringType, "ip_address": types.StringType,
	"description": types.StringType, "user_id": types.StringType,
	"allocated_at": types.StringType, "bulk_id": types.StringType,
}}

type tenancyDataSource struct{ client *apiClient }
type tenancyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewTenancyDataSource() datasource.DataSource { return &tenancyDataSource{} }
func (d *tenancyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenancy"
}
func (d *tenancyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *tenancyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Optional: true, Computed: true},
		"name":        schema.StringAttribute{Optional: true, Computed: true},
		"description": schema.StringAttribute{Computed: true},
		"created_at":  schema.StringAttribute{Computed: true},
	}}
}
func (d *tenancyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tenancyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError("Missing lookup key", "set either id or name")
		return
	}
	var tenancies []tenancyResponse
	if err := d.client.doJSON(ctx, http.MethodGet, "/api/tenancies", nil, &tenancies, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read tenancy failed", err.Error())
		return
	}
	for _, t := range tenancies {
		if (!config.ID.IsNull() && config.ID.ValueString() == t.ID) || (!config.Name.IsNull() && config.Name.ValueString() == t.Name) {
			config.ID = types.StringValue(t.ID)
			config.Name = types.StringValue(t.Name)
			config.Description = types.StringValue(t.Description)
			config.CreatedAt = types.StringValue(t.CreatedAt)
			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}
	resp.Diagnostics.AddError("Tenancy not found", "no matching tenancy found")
}

type usersDataSource struct{ client *apiClient }
type usersDataSourceModel struct {
	TenancyID types.String `tfsdk:"tenancy_id"`
	Users     types.List   `tfsdk:"users"`
}

func NewUsersDataSource() datasource.DataSource { return &usersDataSource{} }
func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}
func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"tenancy_id": schema.StringAttribute{Optional: true},
		"users": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"username":   schema.StringAttribute{Computed: true},
			"role":       schema.StringAttribute{Computed: true},
			"tenancy_id": schema.StringAttribute{Computed: true},
		}}},
	}}
}
func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var users []userResponse
	if err := d.client.doJSON(ctx, http.MethodGet, "/api/users", nil, &users, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read users failed", err.Error())
		return
	}
	items := make([]map[string]types.String, 0)
	for _, u := range users {
		if !config.TenancyID.IsNull() {
			if u.TenancyID == nil || *u.TenancyID != config.TenancyID.ValueString() {
				continue
			}
		}
		tid := types.StringNull()
		if u.TenancyID != nil {
			tid = types.StringValue(*u.TenancyID)
		}
		items = append(items, map[string]types.String{"id": types.StringValue(u.ID), "username": types.StringValue(u.Username), "role": types.StringValue(u.Role), "tenancy_id": tid})
	}
	list, diags := types.ListValueFrom(ctx, userObjectType, items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Users = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type sharedSubnetsDataSource struct{ client *apiClient }
type sharedSubnetsDataSourceModel struct {
	SharedSubnets types.List `tfsdk:"shared_subnets"`
}

func NewSharedSubnetsDataSource() datasource.DataSource { return &sharedSubnetsDataSource{} }
func (d *sharedSubnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_subnets"
}
func (d *sharedSubnetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *sharedSubnetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	subnetListSchema(resp, "shared_subnets")
}
func (d *sharedSubnetsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, diags, err := fetchSubnetList(ctx, d.client, "/api/subnets/shared")
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Read shared subnets failed", err.Error())
		return
	}
	state := sharedSubnetsDataSourceModel{SharedSubnets: list}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type privateSubnetsDataSource struct{ client *apiClient }
type privateSubnetsDataSourceModel struct {
	TenancyID      types.String `tfsdk:"tenancy_id"`
	PrivateSubnets types.List   `tfsdk:"private_subnets"`
}

func NewPrivateSubnetsDataSource() datasource.DataSource { return &privateSubnetsDataSource{} }
func (d *privateSubnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_subnets"
}
func (d *privateSubnetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *privateSubnetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"tenancy_id":      schema.StringAttribute{Required: true},
		"private_subnets": subnetListAttribute(),
	}}
}
func (d *privateSubnetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config privateSubnetsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	list, diags, err := fetchSubnetList(ctx, d.client, fmt.Sprintf("/api/tenancies/%s/subnets", config.TenancyID.ValueString()))
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Read private subnets failed", err.Error())
		return
	}
	config.PrivateSubnets = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type allocationDataSource struct{ client *apiClient }
type allocationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	SubnetID    types.String `tfsdk:"subnet_id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Description types.String `tfsdk:"description"`
	UserID      types.String `tfsdk:"user_id"`
	AllocatedAt types.String `tfsdk:"allocated_at"`
	BulkID      types.String `tfsdk:"bulk_id"`
}

func NewAllocationDataSource() datasource.DataSource { return &allocationDataSource{} }
func (d *allocationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocation"
}
func (d *allocationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *allocationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Required: true}, "subnet_id": schema.StringAttribute{Computed: true}, "ip_address": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true}, "user_id": schema.StringAttribute{Computed: true}, "allocated_at": schema.StringAttribute{Computed: true}, "bulk_id": schema.StringAttribute{Computed: true},
	}}
}
func (d *allocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config allocationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var allocations []allocationResponse
	if err := d.client.doJSON(ctx, http.MethodGet, "/api/allocations", nil, &allocations, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read allocation failed", err.Error())
		return
	}
	for _, a := range allocations {
		if a.ID == config.ID.ValueString() {
			config.SubnetID = types.StringValue(a.SubnetID)
			config.IPAddress = types.StringValue(a.IpAddress)
			config.Description = types.StringValue(a.Description)
			config.UserID = types.StringValue(a.UserID)
			config.AllocatedAt = types.StringValue(a.AllocatedAt)
			if a.BulkID == nil {
				config.BulkID = types.StringNull()
			} else {
				config.BulkID = types.StringValue(*a.BulkID)
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}
	resp.Diagnostics.AddError("Allocation not found", "no matching allocation found")
}

type allocationsDataSource struct{ client *apiClient }
type allocationsDataSourceModel struct {
	TagKey   types.String `tfsdk:"tag_key"`
	TagValue types.String `tfsdk:"tag_value"`
	Items    types.List   `tfsdk:"items"`
}

func NewAllocationsDataSource() datasource.DataSource { return &allocationsDataSource{} }
func (d *allocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocations"
}
func (d *allocationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *allocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"tag_key": schema.StringAttribute{Optional: true}, "tag_value": schema.StringAttribute{Optional: true},
		"items": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}, "subnet_id": schema.StringAttribute{Computed: true}, "ip_address": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true}, "user_id": schema.StringAttribute{Computed: true}, "allocated_at": schema.StringAttribute{Computed: true}, "bulk_id": schema.StringAttribute{Computed: true}}}},
	}}
}
func (d *allocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config allocationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	q := url.Values{}
	if !config.TagKey.IsNull() {
		q.Set("tagKey", config.TagKey.ValueString())
	}
	if !config.TagValue.IsNull() {
		q.Set("tagValue", config.TagValue.ValueString())
	}
	path := "/api/allocations"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var allocations []allocationResponse
	if err := d.client.doJSON(ctx, http.MethodGet, path, nil, &allocations, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read allocations failed", err.Error())
		return
	}
	items := make([]map[string]types.String, 0, len(allocations))
	for _, a := range allocations {
		bulk := types.StringNull()
		if a.BulkID != nil {
			bulk = types.StringValue(*a.BulkID)
		}
		items = append(items, map[string]types.String{"id": types.StringValue(a.ID), "subnet_id": types.StringValue(a.SubnetID), "ip_address": types.StringValue(a.IpAddress), "description": types.StringValue(a.Description), "user_id": types.StringValue(a.UserID), "allocated_at": types.StringValue(a.AllocatedAt), "bulk_id": bulk})
	}
	list, diags := types.ListValueFrom(ctx, allocationObjectType, items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Items = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type subnetStatsDataSource struct{ client *apiClient }
type subnetStatsDataSourceModel struct {
	SubnetID       types.String `tfsdk:"subnet_id"`
	TotalIPs       types.Int64  `tfsdk:"total_ips"`
	AllocatedCount types.Int64  `tfsdk:"allocated_count"`
	FreeCount      types.Int64  `tfsdk:"free_count"`
	ExcludedCount  types.Int64  `tfsdk:"excluded_count"`
}

func NewSubnetStatsDataSource() datasource.DataSource { return &subnetStatsDataSource{} }
func (d *subnetStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_stats"
}
func (d *subnetStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *subnetStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{"subnet_id": schema.StringAttribute{Required: true}, "total_ips": schema.Int64Attribute{Computed: true}, "allocated_count": schema.Int64Attribute{Computed: true}, "free_count": schema.Int64Attribute{Computed: true}, "excluded_count": schema.Int64Attribute{Computed: true}}}
}
func (d *subnetStatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subnetStatsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var stats subnetStatsResponse
	if err := d.client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/subnets/%s/stats", config.SubnetID.ValueString()), nil, &stats, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read subnet stats failed", err.Error())
		return
	}
	config.TotalIPs = types.Int64Value(stats.TotalIPs)
	config.AllocatedCount = types.Int64Value(stats.AllocatedCount)
	config.FreeCount = types.Int64Value(stats.FreeCount)
	config.ExcludedCount = types.Int64Value(stats.ExcludedCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func subnetListAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "cidr": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true}, "description": schema.StringAttribute{Computed: true}, "type": schema.StringAttribute{Computed: true}, "tenancy_id": schema.StringAttribute{Computed: true}, "created_at": schema.StringAttribute{Computed: true},
	}}}
}

func subnetListSchema(resp *datasource.SchemaResponse, field string) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{field: subnetListAttribute()}}
}

func fetchSubnetList(ctx context.Context, client *apiClient, endpoint string) (types.List, diag.Diagnostics, error) {
	var subnets []subnetResponse
	if err := client.doJSON(ctx, http.MethodGet, endpoint, nil, &subnets, http.StatusOK); err != nil {
		return types.ListNull(subnetObjectType), nil, err
	}
	items := make([]map[string]types.String, 0, len(subnets))
	for _, s := range subnets {
		tid := types.StringNull()
		if s.TenancyID != nil {
			tid = types.StringValue(*s.TenancyID)
		}
		items = append(items, map[string]types.String{"id": types.StringValue(s.ID), "cidr": types.StringValue(s.Cidr), "name": types.StringValue(s.Name), "description": types.StringValue(s.Description), "type": types.StringValue(s.Type), "tenancy_id": tid, "created_at": types.StringValue(s.CreatedAt)})
	}
	list, diags := types.ListValueFrom(ctx, subnetObjectType, items)
	return list, diags, nil
}
