package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &bulkAllocationsDataSource{}

type bulkAllocationsDataSource struct{ client *apiClient }

type bulkAllocationsDataSourceModel struct {
	BulkID types.String `tfsdk:"bulk_id"`
	Items  types.List   `tfsdk:"items"`
}

func NewBulkAllocationsDataSource() datasource.DataSource { return &bulkAllocationsDataSource{} }
func (d *bulkAllocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_allocations"
}
func (d *bulkAllocationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	d.client = pd.client
}
func (d *bulkAllocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"bulk_id": schema.StringAttribute{Required: true},
		"items": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"ip_address":  schema.StringAttribute{Computed: true},
			"subnet_id":   schema.StringAttribute{Computed: true},
			"tenancy_id":  schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
		}}},
	}}
}
func (d *bulkAllocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bulkAllocationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var allocations []allocationResponse
	if err := d.client.doJSON(ctx, http.MethodGet, "/api/allocations", nil, &allocations, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read bulk allocations failed", err.Error())
		return
	}
	items := make([]map[string]types.String, 0)
	for _, a := range allocations {
		if a.BulkID == nil || *a.BulkID != config.BulkID.ValueString() {
			continue
		}
		items = append(items, map[string]types.String{
			"id":          types.StringValue(a.ID),
			"ip_address":  types.StringValue(a.IpAddress),
			"subnet_id":   types.StringValue(a.SubnetID),
			"tenancy_id":  types.StringValue(a.TenancyID),
			"description": types.StringValue(a.Description),
		})
	}
	obj := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"ip_address":  types.StringType,
		"subnet_id":   types.StringType,
		"tenancy_id":  types.StringType,
		"description": types.StringType,
	}}
	list, diags := types.ListValueFrom(ctx, obj, items)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Items = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
