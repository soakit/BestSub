package mihomo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/component/proxydialer"
	"github.com/metacubex/mihomo/component/resolver"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/dns"
	"gopkg.in/yaml.v3"

	"github.com/metacubex/mihomo/common/structure"
)

var decoder = structure.NewDecoder(structure.Option{
	TagName:          "proxy",
	WeaklyTypedInput: true,
	KeyReplacer:      structure.DefaultKeyReplacer,
})

func init() {
	rs := dns.NewResolver(dns.Config{
		Main: []dns.NameServer{
			{Net: "https", Addr: "https://doh.pub/dns-query"},
			{Net: "https", Addr: "https://dns.alidns.com/dns-query"},
		},
		// Default 用于引导解析 DoH 服务器自身的域名
		Default: []dns.NameServer{
			{Net: "udp", Addr: "119.29.29.29:53"},
			{Net: "udp", Addr: "223.5.5.5:53"},
		},
	})
	resolver.SystemResolver = rs.Resolver
}

// createAdapter 从单条 YAML 代理配置创建 outbound.ProxyAdapter
// chainDialer 可选，用于链式代理：inner 通过该 dialer 连接自己的服务器
func createAdapter(raw []byte, chainDialer ...constant.Dialer) (outbound.ProxyAdapter, error) {
	var mapping map[string]any
	if err := yaml.Unmarshal(raw, &mapping); err != nil {
		return nil, fmt.Errorf("parse raw: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("nil mapping")
	}
	proxyType, _ := mapping["type"].(string)

	// 链式代理 dialer，设到 option.DialerForAPI 后构造函数会自动使用
	var dialer constant.Dialer
	if len(chainDialer) > 0 {
		dialer = chainDialer[0]
	}

	switch proxyType {
	case "ss":
		o := &outbound.ShadowSocksOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewShadowSocks(*o)
	case "ssr":
		o := &outbound.ShadowSocksROption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewShadowSocksR(*o)
	case "socks5":
		o := &outbound.Socks5Option{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewSocks5(*o)
	case "http":
		o := &outbound.HttpOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewHttp(*o)
	case "vmess":
		o := &outbound.VmessOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewVmess(*o)
	case "vless":
		o := &outbound.VlessOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewVless(*o)
	case "snell":
		o := &outbound.SnellOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewSnell(*o)
	case "trojan":
		o := &outbound.TrojanOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewTrojan(*o)
	case "hysteria":
		o := &outbound.HysteriaOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewHysteria(*o)
	case "hysteria2":
		o := &outbound.Hysteria2Option{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewHysteria2(*o)
	case "wireguard":
		o := &outbound.WireGuardOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewWireGuard(*o)
	case "tuic":
		o := &outbound.TuicOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewTuic(*o)
	case "ssh":
		o := &outbound.SshOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewSsh(*o)
	case "mieru":
		o := &outbound.MieruOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewMieru(*o)
	case "anytls":
		o := &outbound.AnyTLSOption{}
		if err := decoder.Decode(mapping, o); err != nil {
			return nil, err
		}
		o.DialerForAPI = dialer
		return outbound.NewAnyTLS(*o)
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxyType)
	}
}

// NewTransport 创建单代理的 http.Transport
func NewTransport(raw []byte, iface ...string) (*http.Transport, error) {
	var mapping map[string]any
	if err := yaml.Unmarshal(raw, &mapping); err != nil {
		return nil, fmt.Errorf("parse raw: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("nil mapping")
	}
	if len(iface) > 0 && iface[0] != "" {
		mapping["interface-name"] = iface[0]
	}

	processed, _ := yaml.Marshal(mapping)
	proxy, err := createAdapter(processed)
	if err != nil {
		return nil, err
	}
	return buildTransport(proxy), nil
}

// NewChainedTransport 创建链式代理的 http.Transport
// 链路: 本地 -> outer -> inner -> 目标网站
func NewChainedTransport(outer, inner []byte, iface ...string) (*http.Transport, error) {
	// outer 直连互联网，网卡绑定在此
	var outerMapping map[string]any
	if err := yaml.Unmarshal(outer, &outerMapping); err != nil {
		return nil, fmt.Errorf("parse outer: %w", err)
	}
	if outerMapping == nil {
		return nil, fmt.Errorf("nil outer mapping")
	}
	if len(iface) > 0 && iface[0] != "" {
		outerMapping["interface-name"] = iface[0]
	}
	outerProcessed, _ := yaml.Marshal(outerMapping)
	outerAdapter, err := createAdapter(outerProcessed)
	if err != nil {
		return nil, fmt.Errorf("create outer adapter: %w", err)
	}

	// inner 通过 outer 的通道连接自己的服务器
	outerDialer := proxydialer.New(outerAdapter, false)
	innerAdapter, err := createAdapter(inner, outerDialer)
	if err != nil {
		return nil, fmt.Errorf("create inner adapter: %w", err)
	}

	return buildTransport(innerAdapter), nil
}

// buildTransport 从 ProxyAdapter 构建 http.Transport
func buildTransport(proxy outbound.ProxyAdapter) *http.Transport {
	return &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, portStr, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			port, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				port = 0
			}
			return proxy.DialContext(ctx, &constant.Metadata{
				Host:    host,
				DstPort: uint16(port),
			})
		},
	}
}
