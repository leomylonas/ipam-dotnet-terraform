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

var _ resource.Resource = &sharedSubnetAccessResource{}
var _ resource.ResourceWithImportState = &sharedSubnetAccessResource{}

type sharedSubnetAccessResource struct{ client *apiClient }

type sharedSubnetAccessResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SubnetID  types.String `tfsdk:"subnet_id"`
	TenancyID types.String `tfsdk:"tenancy_id"`
}

func NewSharedSubnetAccessResource() resource.Resource { return &sharedSubnetAccessResource{} }
func (r *sharedSubnetAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_subnet_access"
}
func (r *sharedSubnetAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *sharedSubnetAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"subnet_id":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"tenancy_id": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
	}}
}
func (r *sharedSubnetAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharedSubnetAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/subnets/shared/%s/access", plan.SubnetID.ValueString()), map[string]any{
		"tenancyId": plan.TenancyID.ValueString(),
	}, nil, http.StatusNoContent)
	if err != nil {
		resp.Diagnostics.AddError("Grant shared subnet access failed", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.SubnetID.ValueString() + "/" + plan.TenancyID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sharedSubnetAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharedSubnetAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var subnets []subnetResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/subnets/shared", nil, &subnets, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read shared subnet access failed", err.Error())
		return
	}
	for _, s := range subnets {
		if s.ID == state.SubnetID.ValueString() {
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *sharedSubnetAccessResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}
func (r *sharedSubnetAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharedSubnetAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/subnets/shared/%s/access/%s", state.SubnetID.ValueString(), state.TenancyID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Revoke shared subnet access failed", err.Error())
	}
}
func (r *sharedSubnetAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "expected {subnet_id}/{tenancy_id}")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subnet_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenancy_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
