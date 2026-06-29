package urlvisitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// SafetyCheck 用于减少误访问非预期目标的风险。它不是完整的策略引擎，
// 但会为当前工具提供保守的默认安全检查。
func SafetyCheck(ctx context.Context, cfg Config) error {
	targets := cfg.TargetURLs()
	parsedTargets := make([]*url.URL, 0, len(targets))
	for _, target := range targets {
		parsed, err := checkTarget(ctx, cfg, target)
		if err != nil {
			return err
		}
		parsedTargets = append(parsedTargets, parsed)
	}

	if cfg.Safety.RequireAuthorizationConfirm {
		if err := confirmAuthorization(parsedTargets); err != nil {
			return err
		}
	}
	return nil
}

func checkTarget(ctx context.Context, cfg Config, target string) (*url.URL, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败：%w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("不支持的 URL 协议 %q；只允许 http 和 https", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return nil, errors.New("url 必须包含主机名")
	}
	if len(cfg.Safety.AllowedHosts) > 0 && !hostAllowed(parsed.Hostname(), cfg.Safety.AllowedHosts) {
		return nil, fmt.Errorf("主机 %q 不在 safety.allowed_hosts 白名单中", parsed.Hostname())
	}
	if !cfg.Safety.AllowPrivateNetworks {
		if err := rejectPrivateNetwork(ctx, parsed.Hostname()); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}

func hostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, candidate := range allowed {
		candidate = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(candidate), "."))
		if host == candidate {
			return true
		}
	}
	return false
}

func rejectPrivateNetwork(ctx context.Context, host string) error {
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("安全检查解析主机失败：%w", err)
	}
	for _, ipAddr := range ips {
		ip := ipAddr.IP
		if isPrivateOrLocal(ip) {
			return fmt.Errorf("解析到的 IP %s 属于内网或本机地址；只有测试已授权内网目标时才应设置 safety.allow_private_networks=true", ip)
		}
	}
	return nil
}

func isPrivateOrLocal(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func confirmAuthorization(targets []*url.URL) error {
	fmt.Println("目标列表：")
	for _, target := range targets {
		fmt.Printf("  - %s\n", target.String())
	}
	fmt.Print("请输入 YES 确认你已获得授权，可以向该目标发送访问流量：")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return errors.New("未提供授权确认")
	}
	if strings.TrimSpace(scanner.Text()) != "YES" {
		return errors.New("授权确认未通过")
	}
	return nil
}
