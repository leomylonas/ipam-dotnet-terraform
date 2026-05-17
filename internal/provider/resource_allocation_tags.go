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

var _ resource.Resource = &allocationTagsResource{}
var _ resource.ResourceWithImportState = &allocationTagsResource{}

type allocationTagsResource struct{ client *apiClient }

func NewAllocationTagsResource() resource.Resource { return &allocationTagsResource{} }
func (r *allocationTagsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocation_tags"
}
func (r *allocationTagsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *allocationTagsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"allocation_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"tags":          schema.MapAttribute{Required: true, ElementType: types.StringType},
	}}
}
func (r *allocationTagsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan struct {
		ID           types.String `tfsdk:"id"`
		AllocationID types.String `tfsdk:"allocation_id"`
		Tags         types.Map    `tfsdk:"tags"`
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &payload, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/allocations/%s/tags", plan.AllocationID.ValueString()), payload, nil, http.StatusNoContent); err != nil {
		resp.Diagnostics.AddError("Create allocation tags failed", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.AllocationID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *allocationTagsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state struct {
		ID           types.String `tfsdk:"id"`
		AllocationID types.String `tfsdk:"allocation_id"`
		Tags         types.Map    `tfsdk:"tags"`
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var tags []tagResponse
	err := r.client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/allocations/%s/tags", state.AllocationID.ValueString()), nil, &tags, http.StatusOK, http.StatusNotFound)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read allocation tags failed", err.Error())
		return
	}
	vals := map[string]types.String{}
	for _, t := range tags {
		vals[t.Key] = types.StringValue(t.Value)
	}
	mapVal, diags := types.MapValueFrom(ctx, types.StringType, vals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = mapVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *allocationTagsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan struct {
		ID           types.String `tfsdk:"id"`
		AllocationID types.String `tfsdk:"allocation_id"`
		Tags         types.Map    `tfsdk:"tags"`
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := map[string]string{}
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &payload, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/allocations/%s/tags", plan.AllocationID.ValueString()), payload, nil, http.StatusNoContent); err != nil {
		resp.Diagnostics.AddError("Update allocation tags failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *allocationTagsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state struct {
		ID           types.String `tfsdk:"id"`
		AllocationID types.String `tfsdk:"allocation_id"`
		Tags         types.Map    `tfsdk:"tags"`
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/allocations/%s/tags", state.AllocationID.ValueString()), map[string]string{}, nil, http.StatusNoContent, http.StatusNotFound); err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete allocation tags failed", err.Error())
	}
}
func (r *allocationTagsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allocation_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
