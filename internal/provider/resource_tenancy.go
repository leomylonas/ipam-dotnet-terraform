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

var _ resource.Resource = &tenancyResource{}
var _ resource.ResourceWithImportState = &tenancyResource{}

type tenancyResource struct{ client *apiClient }

type tenancyResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	AdminUsername types.String `tfsdk:"admin_username"`
	AdminPassword types.String `tfsdk:"admin_password"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func NewTenancyResource() resource.Resource { return &tenancyResource{} }
func (r *tenancyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenancy"
}
func (r *tenancyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *tenancyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":           schema.StringAttribute{Required: true},
		"description":    schema.StringAttribute{Required: true},
		"admin_username": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"admin_password": schema.StringAttribute{Required: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"created_at":     schema.StringAttribute{Computed: true},
	}}
}
func (r *tenancyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenancyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var created tenancyResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/tenancies", map[string]any{
		"name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
		"adminUsername": plan.AdminUsername.ValueString(), "adminPassword": plan.AdminPassword.ValueString(),
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create tenancy failed", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tenancyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenancyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tenancies []tenancyResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/tenancies", nil, &tenancies, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read tenancies failed", err.Error())
		return
	}
	for _, t := range tenancies {
		if t.ID == state.ID.ValueString() {
			state.Name = types.StringValue(t.Name)
			state.Description = types.StringValue(t.Description)
			state.CreatedAt = types.StringValue(t.CreatedAt)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *tenancyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenancyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var updated tenancyResponse
	err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/tenancies/%s", plan.ID.ValueString()), map[string]any{
		"name": plan.Name.ValueString(), "description": plan.Description.ValueString(),
	}, &updated, http.StatusOK)
	if err != nil {
		resp.Diagnostics.AddError("Update tenancy failed", err.Error())
		return
	}

	plan.CreatedAt = types.StringValue(updated.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *tenancyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenancyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/tenancies/%s", state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete tenancy failed", err.Error())
	}
}
func (r *tenancyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
