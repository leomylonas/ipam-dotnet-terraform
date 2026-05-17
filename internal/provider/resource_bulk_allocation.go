package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &bulkAllocationResource{}
var _ resource.ResourceWithImportState = &bulkAllocationResource{}

type bulkAllocationResource struct{ client *apiClient }

type bulkAllocationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	SubnetID        types.String `tfsdk:"subnet_id"`
	AllocationCount types.Int64  `tfsdk:"allocation_count"`
	Description     types.String `tfsdk:"description"`
	BulkID          types.String `tfsdk:"bulk_id"`
	AllocationIDs   types.List   `tfsdk:"allocation_ids"`
	IPAddresses     types.List   `tfsdk:"ip_addresses"`
}

func NewBulkAllocationResource() resource.Resource { return &bulkAllocationResource{} }
func (r *bulkAllocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_allocation"
}
func (r *bulkAllocationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *bulkAllocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"subnet_id":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"allocation_count": schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
		"description":      schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"bulk_id":          schema.StringAttribute{Computed: true},
		"allocation_ids":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"ip_addresses":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}}
}
func (r *bulkAllocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bulkAllocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var created []allocationResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/allocations/bulk", map[string]any{
		"subnetId":    plan.SubnetID.ValueString(),
		"count":       plan.AllocationCount.ValueInt64(),
		"description": plan.Description.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create bulk allocation failed", err.Error())
		return
	}
	if len(created) == 0 || created[0].BulkID == nil {
		resp.Diagnostics.AddError("Create bulk allocation failed", "API returned empty or missing bulk ID")
		return
	}
	plan.ID = types.StringValue(*created[0].BulkID)
	plan.BulkID = types.StringValue(*created[0].BulkID)
	allocIDs := make([]string, 0, len(created))
	ips := make([]string, 0, len(created))
	for _, a := range created {
		allocIDs = append(allocIDs, a.ID)
		ips = append(ips, a.IpAddress)
	}
	idList, diags := types.ListValueFrom(ctx, types.StringType, allocIDs)
	resp.Diagnostics.Append(diags...)
	ipList, diags := types.ListValueFrom(ctx, types.StringType, ips)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AllocationIDs = idList
	plan.IPAddresses = ipList
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *bulkAllocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bulkAllocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var allocations []allocationResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/allocations", nil, &allocations, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read bulk allocation failed", err.Error())
		return
	}
	bulk := state.BulkID.ValueString()
	allocIDs := make([]string, 0)
	ips := make([]string, 0)
	for _, a := range allocations {
		if a.BulkID != nil && *a.BulkID == bulk {
			allocIDs = append(allocIDs, a.ID)
			ips = append(ips, a.IpAddress)
		}
	}
	if len(allocIDs) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}
	idList, diags := types.ListValueFrom(ctx, types.StringType, allocIDs)
	resp.Diagnostics.Append(diags...)
	ipList, diags := types.ListValueFrom(ctx, types.StringType, ips)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AllocationIDs = idList
	state.IPAddresses = ipList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *bulkAllocationResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}
func (r *bulkAllocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bulkAllocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var ids []string
	resp.Diagnostics.Append(state.AllocationIDs.ElementsAs(ctx, &ids, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, id := range ids {
		err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/allocations/%s", id), nil, nil, http.StatusNoContent, http.StatusNotFound)
		if err != nil && !isStatus(err, http.StatusNotFound) {
			resp.Diagnostics.AddError("Delete bulk allocation failed", err.Error())
			return
		}
	}
}
func (r *bulkAllocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bulk_id"), req.ID)...)
}
