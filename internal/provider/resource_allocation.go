package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &allocationResource{}
var _ resource.ResourceWithImportState = &allocationResource{}

type allocationResource struct{ client *apiClient }

type allocationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SubnetID    types.String `tfsdk:"subnet_id"`
	Description types.String `tfsdk:"description"`
	IPAddress   types.String `tfsdk:"ip_address"`
	UserID      types.String `tfsdk:"user_id"`
	TenancyID   types.String `tfsdk:"tenancy_id"`
	AllocatedAt types.String `tfsdk:"allocated_at"`
	BulkID      types.String `tfsdk:"bulk_id"`
}

func NewAllocationResource() resource.Resource { return &allocationResource{} }
func (r *allocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocation"
}
func (r *allocationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *allocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"subnet_id":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"ip_address":   schema.StringAttribute{Computed: true},
		"user_id":      schema.StringAttribute{Computed: true},
		"tenancy_id":   schema.StringAttribute{Computed: true},
		"allocated_at": schema.StringAttribute{Computed: true},
		"bulk_id":      schema.StringAttribute{Computed: true},
	}}
}
func (r *allocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan allocationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var created allocationResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/allocations", map[string]any{
		"subnetId": plan.SubnetID.ValueString(), "description": plan.Description.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create allocation failed", err.Error())
		return
	}
	plan.ID = types.StringValue(created.ID)
	plan.IPAddress = types.StringValue(created.IpAddress)
	plan.UserID = types.StringValue(created.UserID)
	plan.TenancyID = types.StringValue(created.TenancyID)
	plan.AllocatedAt = types.StringValue(created.AllocatedAt)
	if created.BulkID == nil {
		plan.BulkID = types.StringNull()
	} else {
		plan.BulkID = types.StringValue(*created.BulkID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *allocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state allocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var allocations []allocationResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/allocations", nil, &allocations, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read allocations failed", err.Error())
		return
	}
	for _, a := range allocations {
		if a.ID == state.ID.ValueString() {
			state.SubnetID = types.StringValue(a.SubnetID)
			state.Description = types.StringValue(a.Description)
			state.IPAddress = types.StringValue(a.IpAddress)
			state.UserID = types.StringValue(a.UserID)
			state.TenancyID = types.StringValue(a.TenancyID)
			state.AllocatedAt = types.StringValue(a.AllocatedAt)
			if a.BulkID == nil {
				state.BulkID = types.StringNull()
			} else {
				state.BulkID = types.StringValue(*a.BulkID)
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *allocationResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}
func (r *allocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state allocationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/allocations/%s", state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete allocation failed", err.Error())
	}
}
func (r *allocationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
