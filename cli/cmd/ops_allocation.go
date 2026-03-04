package cmd

import (
	"fmt"
	"net/url"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// --- get-allocation ---

type getAllocationOpts struct {
	f        *internal.Factory
	UserID   string
	TenantID string
}

func newCmdGetAllocation(f *internal.Factory) *cobra.Command {
	o := &getAllocationOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get-allocation",
		Short:   "获取用户的资产分配",
		Example: `  ahcli ops get-allocation --user-id usr-xxx --tenant-id tnt-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.UserID == "" {
				o.UserID = f.UserID()
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			if o.UserID == "" || o.TenantID == "" {
				return fmt.Errorf("--user-id 和 --tenant-id 为必填参数（或先执行 ahcli auth login）")
			}
			return getAllocationRun(o)
		},
	}
	cmd.Flags().StringVar(&o.UserID, "user-id", "", "用户 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

func getAllocationRun(o *getAllocationOpts) error {
	client := o.f.OpsClient()
	params := url.Values{"tenant_id": {o.TenantID}}

	resp, err := client.Get(fmt.Sprintf("/api/v1/allocations/user/%s", o.UserID), params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}
