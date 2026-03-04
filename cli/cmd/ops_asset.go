package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// --- get-asset ---

type getAssetOpts struct {
	f         *internal.Factory
	AssetCode string
}

func newCmdGetAsset(f *internal.Factory) *cobra.Command {
	o := &getAssetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get-asset",
		Short:   "获取 Asset 详情",
		Example: `  ahcli ops get-asset --asset-code asset-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.AssetCode == "" {
				return fmt.Errorf("--asset-code 为必填参数")
			}
			client := o.f.OpsClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/assets/%s", o.AssetCode), nil)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.AssetCode, "asset-code", "", "资产代码（必填）")
	return cmd
}

// --- get-access-config ---

type getAccessConfigOpts struct {
	f         *internal.Factory
	AssetCode string
}

func newCmdGetAccessConfig(f *internal.Factory) *cobra.Command {
	o := &getAccessConfigOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get-access-config",
		Short:   "获取 Asset 的 AccessToken",
		Example: `  ahcli ops get-access-config --asset-code asset-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.AssetCode == "" {
				return fmt.Errorf("--asset-code 为必填参数")
			}
			client := o.f.OpsClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/assets/%s/access-config", o.AssetCode), nil)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.AssetCode, "asset-code", "", "资产代码（必填）")
	return cmd
}

// --- get-credentials ---

type getCredentialsOpts struct {
	f        *internal.Factory
	UserID   string
	TenantID string
}

func newCmdGetCredentials(f *internal.Factory) *cobra.Command {
	o := &getCredentialsOpts{f: f}
	cmd := &cobra.Command{
		Use:   "get-credentials",
		Short: "一次性获取投放所需的全部凭证（组合命令）",
		Long: `将 ads-launch Step 0b 的三步串联为一个命令：
1. GET /api/v1/allocations/user/{userID} → 取 asset_code
2. GET /api/v1/assets/{assetCode} → 取 ad_account_id, pixel_id, page_id
3. GET /api/v1/assets/{assetCode}/access-config → 取 token`,
		Example: `  ahcli ops get-credentials --user-id usr-xxx --tenant-id tnt-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.UserID == "" {
				o.UserID = f.UserID()
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			if o.UserID == "" || o.TenantID == "" {
				return fmt.Errorf("--user-id 和 --tenant-id 为必填参数（或先执行 ahcli auth login）")
			}
			return getCredentialsRun(o)
		},
	}
	cmd.Flags().StringVar(&o.UserID, "user-id", "", "用户 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

func getCredentialsRun(o *getCredentialsOpts) error {
	client := o.f.OpsClient()

	// Step 1: 获取用户资产分配
	params := url.Values{"tenant_id": {o.TenantID}}
	allocResp, err := client.Get(fmt.Sprintf("/api/v1/allocations/user/%s", o.UserID), params)
	if err != nil {
		return fmt.Errorf("获取资产分配失败: %w", err)
	}

	// 解析 allocation 响应（单个对象或数组）
	assetCode, allocationCode, err := parseAllocation(allocResp.Data, o.UserID)
	if err != nil {
		return err
	}
	fmt.Fprintf(internal.Stderr, "[info] 找到资产分配: asset_code=%s\n", assetCode)

	// Step 2: 获取资产详情
	assetResp, err := client.Get(fmt.Sprintf("/api/v1/assets/%s", assetCode), nil)
	if err != nil {
		return fmt.Errorf("获取资产详情失败: %w", err)
	}

	// 解析资产详情，提取 ad_account_id 和属性
	var assetData struct {
		AdAccountID string `json:"ad_account_id"`
		Properties  []struct {
			PropertyType string `json:"property_type"`
			PropertyID   string `json:"property_id"`
		} `json:"properties"`
	}
	if unmarshalErr := json.Unmarshal(assetResp.Data, &assetData); unmarshalErr != nil {
		return fmt.Errorf("解析资产详情失败: %w", unmarshalErr)
	}

	// 从 properties 中提取 pixel_id 和 page_id
	var pixelID, pageID string
	for _, p := range assetData.Properties {
		switch p.PropertyType {
		case "pixel":
			pixelID = p.PropertyID
		case "facebook_page":
			pageID = p.PropertyID
		}
	}
	fmt.Fprintf(internal.Stderr, "[info] 资产详情: ad_account_id=%s, pixel_id=%s, page_id=%s\n",
		assetData.AdAccountID, pixelID, pageID)

	// Step 3: 获取 access token
	accessResp, err := client.Get(fmt.Sprintf("/api/v1/assets/%s/access-config", assetCode), nil)
	if err != nil {
		return fmt.Errorf("获取 access config 失败: %w", err)
	}

	var accessData struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(accessResp.Data, &accessData); err != nil {
		return fmt.Errorf("解析 access config 失败: %w", err)
	}

	// 组装输出结构
	result := map[string]string{
		"asset_code":      assetCode,
		"allocation_code": allocationCode,
		"ad_account_id":   assetData.AdAccountID,
		"page_id":         pageID,
		"pixel_id":        pixelID,
		"access_token":    accessData.Token,
	}

	resultBytes, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal result: %w", marshalErr)
	}
	return o.f.Print(resultBytes)
}

// parseAllocation 解析 allocation 响应，支持单个对象或数组格式。
func parseAllocation(data json.RawMessage, userID string) (assetCode, allocationCode string, err error) {
	var single struct {
		AssetCode      string `json:"asset_code"`
		AllocationCode string `json:"allocation_code"`
	}
	if json.Unmarshal(data, &single) == nil && single.AssetCode != "" {
		return single.AssetCode, single.AllocationCode, nil
	}

	var list []struct {
		AssetCode      string `json:"asset_code"`
		AllocationCode string `json:"allocation_code"`
	}
	if unmarshalErr := json.Unmarshal(data, &list); unmarshalErr != nil {
		return "", "", fmt.Errorf("解析资产分配响应失败: %w", unmarshalErr)
	}
	if len(list) == 0 {
		return "", "", fmt.Errorf("用户 %s 没有资产分配", userID)
	}
	return list[0].AssetCode, list[0].AllocationCode, nil
}
