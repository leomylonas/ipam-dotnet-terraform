package provider

type tenancyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

type subnetResponse struct {
	ID          string  `json:"id"`
	Cidr        string  `json:"cidr"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	TenancyID   *string `json:"tenancyId"`
	CreatedAt   string  `json:"createdAt"`
}

type exclusionResponse struct {
	ID          string `json:"id"`
	SubnetID    string `json:"subnetId"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
}

type userResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	TenancyID *string `json:"tenancyId"`
}

type allocationResponse struct {
	ID          string  `json:"id"`
	IpAddress   string  `json:"ipAddress"`
	UserID      string  `json:"userId"`
	SubnetID    string  `json:"subnetId"`
	Description string  `json:"description"`
	AllocatedAt string  `json:"allocatedAt"`
	BulkID      *string `json:"bulkId"`
}

type tagResponse struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type subnetStatsResponse struct {
	SubnetID       string `json:"subnetId"`
	TotalIPs       int64  `json:"totalIps"`
	AllocatedCount int64  `json:"allocatedCount"`
	FreeCount      int64  `json:"freeCount"`
	ExcludedCount  int64  `json:"excludedCount"`
}
