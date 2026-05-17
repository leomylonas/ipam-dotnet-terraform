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

var _ resource.Resource = &sharedSubnetResource{}
var _ resource.ResourceWithImportState = &sharedSubnetResource{}

type sharedSubnetResource struct{ client *apiClient }

type sharedSubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Cidr        types.String `tfsdk:"cidr"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewSharedSubnetResource() resource.Resource { return &sharedSubnetResource{} }
func (r *sharedSubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_subnet"
}
func (r *sharedSubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *sharedSubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"cidr":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"name":        schema.StringAttribute{Required: true},
		"description": schema.StringAttribute{Required: true},
		"created_at":  schema.StringAttribute{Computed: true},
	}}
}
func (r *sharedSubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharedSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var created subnetResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/subnets/shared", map[string]any{
		"cidr": plan.Cidr.ValueString(), "name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create shared subnet failed", err.Error())
		return
	}
	plan.ID = types.StringValue(created.ID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sharedSubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharedSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var subnets []subnetResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/subnets/shared", nil, &subnets, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read shared subnets failed", err.Error())
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
func (r *sharedSubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sharedSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var updated subnetResponse
	err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/subnets/shared/%s", plan.ID.ValueString()), map[string]any{
		"name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
	}, &updated, http.StatusOK)
	if err != nil {
		resp.Diagnostics.AddError("Update shared subnet failed", err.Error())
		return
	}
	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *sharedSubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharedSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/subnets/shared/%s", state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete shared subnet failed", err.Error())
	}
}
func (r *sharedSubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
