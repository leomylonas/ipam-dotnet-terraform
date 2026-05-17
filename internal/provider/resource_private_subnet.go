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

var _ resource.Resource = &privateSubnetResource{}
var _ resource.ResourceWithImportState = &privateSubnetResource{}

type privateSubnetResource struct{ client *apiClient }

type privateSubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	TenancyID   types.String `tfsdk:"tenancy_id"`
	Cidr        types.String `tfsdk:"cidr"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewPrivateSubnetResource() resource.Resource { return &privateSubnetResource{} }
func (r *privateSubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_subnet"
}
func (r *privateSubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *privateSubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"tenancy_id":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"cidr":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"name":        schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Required: true},
		"created_at":  schema.StringAttribute{Computed: true},
	}}
}
func (r *privateSubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan privateSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var created subnetResponse
	err := r.client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/tenancies/%s/subnets", plan.TenancyID.ValueString()), map[string]any{
		"cidr": plan.Cidr.ValueString(), "name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create private subnet failed", err.Error())
		return
	}
	plan.ID = types.StringValue(created.ID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *privateSubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privateSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var subnets []subnetResponse
	err := r.client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/tenancies/%s/subnets", state.TenancyID.ValueString()), nil, &subnets, http.StatusOK, http.StatusNotFound)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read private subnet failed", err.Error())
		return
	}
	for _, s := range subnets {
		if s.ID == state.ID.ValueString() {
			state.Name = types.StringValue(s.Name)
			state.Description = types.StringValue(s.Description)
			state.Cidr = types.StringValue(s.Cidr)
			state.CreatedAt = types.StringValue(s.CreatedAt)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *privateSubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan privateSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var updated subnetResponse
	err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/tenancies/%s/subnets/%s", plan.TenancyID.ValueString(), plan.ID.ValueString()), map[string]any{
		"name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
	}, &updated, http.StatusOK)
	if err != nil {
		resp.Diagnostics.AddError("Update private subnet failed", err.Error())
		return
	}
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *privateSubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privateSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/tenancies/%s/subnets/%s", state.TenancyID.ValueString(), state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete private subnet failed", err.Error())
	}
}
func (r *privateSubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "expected {tenancy_id}/{subnet_id}")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenancy_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
