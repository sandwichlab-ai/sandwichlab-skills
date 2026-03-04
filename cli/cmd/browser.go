package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// screenshotRequest 截图请求参数，对应 POST /screenshot 的请求体
type screenshotRequest struct {
	URL      string `json:"url"`       // 目标网页 URL
	FullPage bool   `json:"full_page"` // 是否截取完整页面（包括滚动区域）
}

// screenshotData 截图响应数据
type screenshotData struct {
	Screenshot     string  `json:"screenshot"`       // Base64 编码的 PNG 图片
	URL            string  `json:"url"`              // 实际截图的 URL
	FullPage       bool    `json:"full_page"`        // 是否为全页截图
	Duration       float64 `json:"duration_seconds"` // 截图耗时（秒）
	CaptureMode    string  `json:"capture_mode"`     // 截取模式：full / clipped / viewport_fallback
	CapturedHeight int     `json:"captured_height"`  // 实际截取高度（像素）
	PageHeight     int     `json:"page_height"`      // 页面总高度（像素）
}

// --- screenshot ---

type screenshotOpts struct {
	f        *internal.Factory
	URL      string
	FullPage bool
	Save     string
}

func newCmdScreenshot(f *internal.Factory) *cobra.Command {
	o := &screenshotOpts{f: f}
	cmd := &cobra.Command{
		Use:   "screenshot",
		Short: "对网页进行截图",
		Example: `  ahcli browser screenshot --url https://example.com
  ahcli browser screenshot --url https://example.com --full-page --save page.png`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.URL == "" {
				return fmt.Errorf("--url 为必填参数")
			}
			return screenshotRun(o)
		},
	}
	cmd.Flags().StringVar(&o.URL, "url", "", "目标网页 URL（必填）")
	cmd.Flags().BoolVar(&o.FullPage, "full-page", false, "截取完整页面（包括滚动区域）")
	cmd.Flags().StringVar(&o.Save, "save", "", "将截图保存为 PNG 文件的路径")
	return cmd
}

func screenshotRun(o *screenshotOpts) error {
	// 构造请求体
	reqBody := screenshotRequest{
		URL:      o.URL,
		FullPage: o.FullPage,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	client := o.f.BrowserClient()
	resp, err := client.Post("/screenshot", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	// 指定了 --save 时，将 base64 解码后保存为 PNG 文件
	if o.Save != "" {
		var data screenshotData
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return fmt.Errorf("failed to parse screenshot response: %w", err)
		}

		pngBytes, err := base64.StdEncoding.DecodeString(data.Screenshot)
		if err != nil {
			return fmt.Errorf("failed to decode base64 screenshot: %w", err)
		}

		if err := os.WriteFile(o.Save, pngBytes, 0600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", o.Save, err)
		}

		fmt.Fprintf(os.Stderr, "截图已保存到 %s（%d 字节，模式: %s，耗时: %.1f 秒）\n",
			o.Save, len(pngBytes), data.CaptureMode, data.Duration)
		return nil
	}

	// 未指定 --save 时，输出完整的 JSON 响应（含 base64 数据）
	return o.f.Print(resp.Data)
}

// NewCmdBrowser 创建 BrowserService 子命令组。
func NewCmdBrowser(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browser",
		Short: "BrowserService 浏览器服务命令",
		Long: `调用 BrowserService 浏览器服务的 API，包括网页截图等功能。

可用命令:
  screenshot   对指定 URL 进行网页截图

截图支持两种输出模式:
  --save file.png   将 base64 解码后保存为 PNG 文件
  （不指定 --save） 输出完整 JSON 响应（含 base64 数据）

示例:
  ahcli browser screenshot --url https://example.com --save out.png
  ahcli browser screenshot --url https://example.com --full-page --save full.png`,
	}

	cmd.AddCommand(newCmdScreenshot(f))

	return cmd
}
