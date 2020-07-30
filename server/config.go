package server

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

// CertConfig 证书路径配置
type CertConfig struct {
	Public  string
	Private string
	Addr    string
}

// RadiusAuthConfig radius配置
type RadiusAuthConfig struct {
	NASID   string
	Server  string
	Secret  string
	Timeout time.Duration
	Retry   uint
}

// RadiusAcctConfig radius配置
type RadiusAcctConfig struct {
	NASID   string
	Server  string
	Secret  string
	Timeout time.Duration
	Retry   uint
}

// GroupConfig 用户组配置
type GroupConfig struct {
	Network        string
	Gateway        string
	DNS            []string
	TunnelAllDNS   bool
	Route          []string
	NoRoute        []string
	Keepalive      time.Duration
	IdelTimeout    time.Duration
	MaxSessionUser uint
}

// MainConfig 主配置项
type MainConfig struct {
	Listen       []string
	PublicDir    string
	LogFile      string
	PIDFile      string
	Auth         string
	Acct         string
	DefaultGroup string
	LZSCompress  bool
}

// Config 配置项
type Config struct {
	Main  *MainConfig
	Cert  []*CertConfig
	Auth  map[string]interface{}
	Acct  map[string]interface{}
	Group map[string]*GroupConfig
}

// NewConfig 读取载入配置文件
func NewConfig(path string) (*Config, error) {
	c := &Config{
		Main: &MainConfig{
			Listen: make([]string, 0),
		},
		Cert:  make([]*CertConfig, 0),
		Auth:  make(map[string]interface{}),
		Acct:  make(map[string]interface{}),
		Group: make(map[string]*GroupConfig),
	}
	loadOptions := ini.LoadOptions{
		AllowBooleanKeys: true,
		AllowShadows:     true,
		Insensitive:      true,
	}
	f, err := ini.LoadSources(loadOptions, path)
	if err != nil {
		return nil, err
	}
	sections := f.Sections()
	certSectionReg := regexp.MustCompile(`^cert[\t ]*:[\t ]*([^ ]+)$`)
	authSectionReg := regexp.MustCompile(`^auth[\t ]*:[\t ]*([^ ]+)$`)
	acctSectionReg := regexp.MustCompile(`^acct[\t ]*:[\t ]*([^ ]+)$`)
	groupSectionReg := regexp.MustCompile(`^group[\t ]*:[\t ]*([^ ]+)$`)

	for _, section := range sections {
		sectionName := strings.TrimSpace(section.Name())
		if sectionName == "default" {
			err = handleDefaultSection(section, c)
		} else if certSectionReg.MatchString(sectionName) {
			match := certSectionReg.FindStringSubmatch(sectionName)
			err = handleCertSection(section, c, match[1])
		} else if authSectionReg.MatchString(sectionName) {
			match := authSectionReg.FindStringSubmatch(sectionName)
			err = handleAuthSection(section, c, match[1])
		} else if acctSectionReg.MatchString(sectionName) {
			match := acctSectionReg.FindStringSubmatch(sectionName)
			err = handleAcctSection(section, c, match[1])
		} else if groupSectionReg.MatchString(sectionName) {
			match := groupSectionReg.FindStringSubmatch(sectionName)
			err = handleGroupSection(section, c, match[1])
		} else {
			err = errors.New("unkown section name: [" + sectionName + "]")
		}
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func handleDefaultSection(section *ini.Section, c *Config) error {
	c.Main.Listen = section.Key("Listen").ValueWithShadows()
	c.Main.PublicDir = section.Key("PublicDir").String()
	c.Main.LogFile = section.Key("LogFile").String()
	if c.Main.LogFile == "nil" {
		c.Main.LogFile = ""
	}
	c.Main.PIDFile = section.Key("PIDFile").String()
	if c.Main.PIDFile == "nil" {
		c.Main.PIDFile = ""
	}

	c.Main.Auth = section.Key("Auth").String()
	c.Main.DefaultGroup = section.Key("DefaultGroup").String()
	lzcompress, err := section.Key("LZSCompress").Bool()
	if err != nil {
		return err
	}
	c.Main.LZSCompress = lzcompress
	return nil
}

func handleCertSection(section *ini.Section, c *Config, name string) error {
	cert := &CertConfig{}
	cert.Public = section.Key("Public").String()
	cert.Private = section.Key("Private").String()
	cert.Addr = section.Key("Addr").String()
	c.Cert = append(c.Cert, cert)
	return nil
}

func handleAuthSection(section *ini.Section, c *Config, name string) error {
	authType := section.Key("Type").String()
	switch authType {
	case "radius":
		return handleRadiusAuthSection(section, c, name)
	}
	return errors.New("unkown auth type: " + authType)
}

func handleAcctSection(section *ini.Section, c *Config, name string) error {
	acctType := section.Key("Type").String()
	switch acctType {
	case "radius":
		return handleRadiusAcctSection(section, c, name)
	}
	return errors.New("unkown acct type: " + acctType)
}

func handleRadiusAuthSection(section *ini.Section, c *Config, name string) error {
	var err error

	_, exists := c.Auth[name]
	if exists {
		return errors.New("duplicate auth name: " + name)
	}

	radiusAuth := &RadiusAuthConfig{}

	radiusAuth.NASID = section.Key("NASID").String()
	radiusAuth.Server = section.Key("Server").String()
	radiusAuth.Secret = section.Key("Secret").String()
	radiusAuth.Timeout, err = section.Key("Timeout").Duration()
	if err != nil {
		return err
	}
	radiusAuth.Retry, err = section.Key("Retry").Uint()
	if err != nil {
		return err
	}

	c.Auth[name] = radiusAuth
	return nil
}

func handleRadiusAcctSection(section *ini.Section, c *Config, name string) error {
	var err error

	_, exists := c.Acct[name]
	if exists {
		return errors.New("duplicate acct name: " + name)
	}

	radiusAcct := &RadiusAcctConfig{}

	radiusAcct.NASID = section.Key("NASID").String()
	radiusAcct.Server = section.Key("Server").String()
	radiusAcct.Secret = section.Key("Secret").String()
	radiusAcct.Timeout, err = section.Key("Timeout").Duration()
	if err != nil {
		return err
	}
	radiusAcct.Retry, err = section.Key("Retry").Uint()
	if err != nil {
		return err
	}

	c.Acct[name] = radiusAcct
	return nil
}

func handleGroupSection(section *ini.Section, c *Config, name string) error {
	var err error

	_, exists := c.Group[name]
	if exists {
		return errors.New("duplicate group name: " + name)
	}

	groupConfig := &GroupConfig{}
	groupConfig.Network = section.Key("Network").String()
	groupConfig.Gateway = section.Key("Gateway").String()
	groupConfig.DNS = section.Key("DNS").ValueWithShadows()
	groupConfig.TunnelAllDNS, err = section.Key("TunnelAllDNS").Bool()
	if err != nil {
		return err
	}
	groupConfig.Route = section.Key("Route").ValueWithShadows()
	groupConfig.NoRoute = section.Key("NoRoute").ValueWithShadows()
	groupConfig.Keepalive, err = section.Key("KeepAlive").Duration()
	if err != nil {
		return err
	}
	groupConfig.IdelTimeout, err = section.Key("IdleTimeout").Duration()
	if err != nil {
		return err
	}
	groupConfig.MaxSessionUser, err = section.Key("MaxUserSession").Uint()
	if err != nil {
		return err
	}

	c.Group[name] = groupConfig
	return nil
}
