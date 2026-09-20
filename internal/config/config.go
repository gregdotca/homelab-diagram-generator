package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title       string           `yaml:"title"`
	VLANs       []VLAN           `yaml:"vlans"`
	BadgeList   []Badge          `yaml:"badges"`
	Badges      map[string]Badge `yaml:"-"`
	Sections    []Section        `yaml:"sections"`
	SectionMap  map[int]Section  `yaml:"-"`
	Devices     []Device         `yaml:"devices"`
	Groups      []Group          `yaml:"groups"`
	Style       *DiagramStyle    `yaml:"style"`
}

type Section struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
}

type DiagramStyle struct {
	Background   string  `yaml:"background"`
	FontFamily   string  `yaml:"font_family"`
	CardWidth    float64 `yaml:"card_width"`
	CardPadding  float64 `yaml:"card_padding"`
	GroupPadding float64 `yaml:"group_padding"`
	GroupSpacing float64 `yaml:"group_spacing"`
}

type VLAN struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
	CIDR  string `yaml:"cidr"`
	Color string `yaml:"color"`
}

type Badge struct {
	ID     string `yaml:"id"`
	Label  string `yaml:"label"`
	Color  string `yaml:"color"`
	Letter string `yaml:"letter"`
}

type DeviceDefaults struct {
	Icon     string   `yaml:"icon"`
	OS       string   `yaml:"os"`
	Hardware string   `yaml:"hardware"`
	VLAN     string   `yaml:"vlan"`
	Badges   []string `yaml:"badges"`
}

type Group struct {
	ID       string         `yaml:"id"`
	Label    string         `yaml:"label"`
	Position string         `yaml:"position"`
	Section  string         `yaml:"section"`
	Row      int            `yaml:"-"`
	Style    GroupStyle     `yaml:"style"`
	Defaults DeviceDefaults `yaml:"defaults"`
	Devices  []Device       `yaml:"devices"`
	Columns  int            `yaml:"columns"`
}

type GroupStyle struct {
	Background string `yaml:"background"`
	Border     string `yaml:"border"`
}

type Device struct {
	ID         string      `yaml:"id"`
	Label      string      `yaml:"label"`
	Icon       string      `yaml:"icon"`
	OS         string      `yaml:"os"`
	Hardware   string      `yaml:"hardware"`
	Specs      string      `yaml:"specs"`
	IP         string      `yaml:"ip"`
	VLAN       string      `yaml:"vlan"`
	Badges     []string    `yaml:"badges"`
	Services   ServiceList `yaml:"services"`
	Connection string      `yaml:"connection"`
	Section    string      `yaml:"section"`
}

type ServiceList []Service

func (s *ServiceList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		for _, item := range value.Content {
			switch item.Kind {
			case yaml.ScalarNode:
				*s = append(*s, Service{Name: item.Value})
			case yaml.MappingNode:
				var svc Service
				if err := item.Decode(&svc); err != nil {
					return err
				}
				*s = append(*s, svc)
			}
		}
		return nil
	}
	return fmt.Errorf("services must be a list")
}

type Service struct {
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Title == "" {
		c.Title = "Homelab Diagram"
	}

	if c.Style == nil {
		c.Style = &DiagramStyle{}
	}
	if c.Style.Background == "" {
		c.Style.Background = "#ffffff"
	}
	if c.Style.FontFamily == "" {
		c.Style.FontFamily = "Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
	}
	if c.Style.CardWidth == 0 {
		c.Style.CardWidth = 220
	}
	if c.Style.CardPadding == 0 {
		c.Style.CardPadding = 8
	}
	if c.Style.GroupPadding == 0 {
		c.Style.GroupPadding = 10
	}
	if c.Style.GroupSpacing == 0 {
		c.Style.GroupSpacing = 15
	}

	defaultColors := []string{"#2196F3", "#4CAF50", "#FF9800", "#9C27B0", "#F44336", "#E91E63", "#795548", "#607D8B", "#D32F2F", "#8BC34A"}
	for i := range c.VLANs {
		if c.VLANs[i].Color == "" {
			c.VLANs[i].Color = defaultColors[i%len(defaultColors)]
		}
	}

	defaultBadgeColors := []string{"#2196F3", "#4CAF50", "#FF9800", "#9C27B0", "#F44336", "#E91E63", "#795548", "#607D8B", "#D32F2F", "#8BC34A"}
	c.Badges = make(map[string]Badge)
	for i, badge := range c.BadgeList {
		if badge.Letter == "" {
			badge.Letter = strings.ToUpper(badge.ID[:1])
		}
		if badge.Color == "" {
			badge.Color = defaultBadgeColors[i%len(defaultBadgeColors)]
		}
		c.BadgeList[i] = badge
		c.Badges[badge.ID] = badge
	}

	sectionIndex := make(map[string]int)
	c.SectionMap = make(map[int]Section)
	for i, sec := range c.Sections {
		row := i + 1
		sectionIndex[sec.ID] = row
		c.SectionMap[row] = sec
	}

	// Promote standalone devices to single-device groups
	for _, dev := range c.Devices {
		g := Group{
			ID:      dev.ID + "-group",
			Label:   dev.Label,
			Section: dev.Section,
			Devices: []Device{dev},
		}
		c.Groups = append(c.Groups, g)
	}

	// Apply group defaults to devices and resolve section → row
	for i := range c.Groups {
		if c.Groups[i].Section != "" {
			if row, ok := sectionIndex[c.Groups[i].Section]; ok {
				c.Groups[i].Row = row
			}
		}

		if c.Groups[i].Position == "" {
			c.Groups[i].Position = "auto"
		}
		if c.Groups[i].Style.Background == "" {
			c.Groups[i].Style.Background = "#ffffff"
		}
		if c.Groups[i].Style.Border == "" {
			c.Groups[i].Style.Border = "#d0d5db"
		}

		defaults := c.Groups[i].Defaults
		for j := range c.Groups[i].Devices {
			d := &c.Groups[i].Devices[j]
			if d.Icon == "" && defaults.Icon != "" {
				d.Icon = defaults.Icon
			}
			if d.Icon == "" {
				d.Icon = "generic"
			}
			if d.OS == "" && defaults.OS != "" {
				d.OS = defaults.OS
			}
			if d.Hardware == "" && defaults.Hardware != "" {
				d.Hardware = defaults.Hardware
			}
			if d.VLAN == "" && defaults.VLAN != "" {
				d.VLAN = defaults.VLAN
			}
			if len(d.Badges) == 0 && len(defaults.Badges) > 0 {
				d.Badges = make([]string, len(defaults.Badges))
				copy(d.Badges, defaults.Badges)
			}
		}
	}
}

func (c *Config) validate() error {
	deviceIDs := make(map[string]bool)
	for _, g := range c.Groups {
		for _, d := range g.Devices {
			if d.ID == "" {
				return fmt.Errorf("device in group %q has no id", g.ID)
			}
			if deviceIDs[d.ID] {
				return fmt.Errorf("duplicate device id: %q", d.ID)
			}
			deviceIDs[d.ID] = true
		}
	}

	return nil
}

func (c *Config) FindDevice(id string) (*Device, *Group) {
	for gi := range c.Groups {
		for di := range c.Groups[gi].Devices {
			if c.Groups[gi].Devices[di].ID == id {
				return &c.Groups[gi].Devices[di], &c.Groups[gi]
			}
		}
	}
	return nil, nil
}

func (c *Config) VLANColor(vlanID string) string {
	for _, v := range c.VLANs {
		if v.ID == vlanID {
			return v.Color
		}
	}
	return "#999999"
}
