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

var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

type userResource struct{ client *apiClient }

type userResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Role      types.String `tfsdk:"role"`
	TenancyID types.String `tfsdk:"tenancy_id"`
}

func NewUserResource() resource.Resource { return &userResource{} }
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	pd, err := providerDataFrom(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Provider not configured", err.Error())
		return
	}
	r.client = pd.client
}
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"username":   schema.StringAttribute{Required: true},
		"password":   schema.StringAttribute{Required: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"role":       schema.StringAttribute{Required: true},
		"tenancy_id": schema.StringAttribute{Optional: true},
	}}
}
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tenancyID any = nil
	if !plan.TenancyID.IsNull() {
		tenancyID = plan.TenancyID.ValueString()
	}
	var created userResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/users", map[string]any{
		"username": plan.Username.ValueString(), "password": plan.Password.ValueString(),
		"role": plan.Role.ValueString(), "tenancyId": tenancyID,
	}, &created, http.StatusCreated)
	if err != nil {
		resp.Diagnostics.AddError("Create user failed", err.Error())
		return
	}
	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var users []userResponse
	if err := r.client.doJSON(ctx, http.MethodGet, "/api/users", nil, &users, http.StatusOK); err != nil {
		resp.Diagnostics.AddError("Read users failed", err.Error())
		return
	}
	for _, u := range users {
		if u.ID == state.ID.ValueString() {
			state.Username = types.StringValue(u.Username)
			state.Role = types.StringValue(u.Role)
			if u.TenancyID == nil {
				state.TenancyID = types.StringNull()
			} else {
				state.TenancyID = types.StringValue(*u.TenancyID)
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var tenancyID any = nil
	if !plan.TenancyID.IsNull() {
		tenancyID = plan.TenancyID.ValueString()
	}
	var updated userResponse
	err := r.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/api/users/%s", plan.ID.ValueString()), map[string]any{
		"username": plan.Username.ValueString(), "role": plan.Role.ValueString(), "tenancyId": tenancyID,
	}, &updated, http.StatusOK)
	if err != nil {
		resp.Diagnostics.AddError("Update user failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/api/users/%s", state.ID.ValueString()), nil, nil, http.StatusNoContent, http.StatusNotFound)
	if err != nil && !isStatus(err, http.StatusNotFound) {
		resp.Diagnostics.AddError("Delete user failed", err.Error())
	}
}
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
