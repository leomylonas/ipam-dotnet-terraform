package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &exclusionResource{}
var _ resource.ResourceWithImportState = &exclusionResource{}

type exclusionResource struct{ client *apiClient }

type exclusionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SubnetID    types.String `tfsdk:"subnet_id"`
	Start       types.String `tfsdk:"start"`
	End         types.String `tfsdk:"end"`
	Description types.String `tfsdk:"description"`
}

func NewExclusionResource() resource.Resource { return &exclusionResource{} }
func (r *exclusionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exclusion"
}
func (r *exclusionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *exclusionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"subnet_id":   schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"start":       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"end":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description": schema.StringAttribute{Required: true},
	}}
}
func (r *exclusionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan exclusionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var created exclusionResponse
	err := r.client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/subnets/%s/exclusions", plan.SubnetID.ValueString()), map[string]any{
		"start": plan.Start.ValueString(), "end": plan.End.ValueString(), "description": plan.Description.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create exclusion failed", err.Error())
		return
	}
	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *exclusionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state exclusionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var exclusions []exclusionResponse
	err := r.client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/subnets/%s/exclusions", state.SubnetID.ValueString()), nil, &exclusions, http.StatusOK, http.StatusNotFound)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read exclusion failed", err.Error())
		return
	}
	for _, e := range exclusions {
		if e.ID == state.ID.ValueString() {
			state.Description = types.StringValue(e.Description)
			state.Start = types.StringValue(e.Start)
			state.End = types.StringValue(e.End)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *exclusionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan exclusionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var updated exclusionResponse
	err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/subnets/%s/exclusions/%s", plan.SubnetID.ValueString(), plan.ID.ValueString()), map[string]any{
		"description": plan.Description.ValueString(),
	}, &updated, http.StatusOK)
	if err != nil {
		resp.Diagnostics.AddError("Update exclusion failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *exclusionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state exclusionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/subnets/%s/exclusions/%s", state.SubnetID.ValueString(), state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete exclusion failed", err.Error())
	}
}
func (r *exclusionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "expected {subnet_id}/{exclusion_id}")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subnet_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
