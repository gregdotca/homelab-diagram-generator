package render

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/gregdotca/homelab-diagram-generator/internal/config"
	"github.com/gregdotca/homelab-diagram-generator/internal/icons"
	"github.com/gregdotca/homelab-diagram-generator/internal/layout"
)

const (
	iconSize    = 40
	badgeIconSz = 16
	fontSize    = 12
	titleFontSz = 20
	labelFontSz = 14
	smallFontSz = 10
	cardRadius  = 8
	groupRadius = 12
	vlanStripeW = 4
)

func SVG(scene *layout.Scene) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">`,
		scene.Width, scene.Height, scene.Width, scene.Height))
	b.WriteString("\n")

	writeStyles(&b, scene)
	writeDefs(&b)

	// background
	b.WriteString(fmt.Sprintf(`<rect width="%.0f" height="%.0f" fill="%s"/>`, scene.Width, scene.Height, scene.Background))
	b.WriteString("\n")

	// title
	b.WriteString(fmt.Sprintf(`<text x="%.0f" y="30" class="title">%s</text>`,
		scene.Width/2, esc(scene.Title)))
	b.WriteString("\n")

	// legend
	writeLegend(&b, scene)

	// sections (row containers)
	for _, sec := range scene.Sections {
		writeSection(&b, sec)
	}

	// groups and cards
	for _, g := range scene.Groups {
		writeGroup(&b, g, scene.Legend.Badges)
	}

	b.WriteString("</svg>\n")
	return b.String()
}

func writeStyles(b *strings.Builder, scene *layout.Scene) {
	b.WriteString("<style>\n")
	b.WriteString(fmt.Sprintf(`  text { font-family: %s; fill: #333; }`, scene.FontFamily))
	b.WriteString("\n")
	b.WriteString(`  .title { font-size: 20px; font-weight: 700; text-anchor: middle; fill: #1a1a2e; }`)
	b.WriteString("\n")
	b.WriteString(`  .group-label { font-size: 13px; font-weight: 600; fill: #495057; }`)
	b.WriteString("\n")
	b.WriteString(`  .card-label { font-size: 12px; font-weight: 600; fill: #212529; }`)
	b.WriteString("\n")
	b.WriteString(`  .card-detail { font-size: 10px; fill: #6c757d; }`)
	b.WriteString("\n")
	b.WriteString(`  .card-ip { font-size: 10px; fill: #495057; font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace; }`)
	b.WriteString("\n")
	b.WriteString(`  .legend-title { font-size: 12px; font-weight: 700; fill: #333; }`)
	b.WriteString("\n")
	b.WriteString(`  .legend-text { font-size: 10px; fill: #555; }`)
	b.WriteString("\n")
	b.WriteString(`  .service-text { font-size: 9px; fill: #6c757d; }`)
	b.WriteString("\n")
	b.WriteString(`  .section-label { font-size: 13px; font-weight: 700; fill: #868e96; letter-spacing: 0.5px; }`)
	b.WriteString("\n")
	b.WriteString(`  .badge-letter { font-size: 11px; font-weight: 700; fill: #ffffff; text-anchor: middle; }`)
	b.WriteString("\n")
	b.WriteString("</style>\n")
}

func writeDefs(b *strings.Builder) {
	b.WriteString("<defs>\n")
	b.WriteString(`  <filter id="card-shadow" x="-4%" y="-4%" width="108%" height="112%">`)
	b.WriteString("\n")
	b.WriteString(`    <feDropShadow dx="0" dy="1" stdDeviation="2" flood-opacity="0.1"/>`)
	b.WriteString("\n")
	b.WriteString("  </filter>\n")
	b.WriteString("</defs>\n")
}

func writeLegend(b *strings.Builder, scene *layout.Scene) {
	x := scene.Legend.X
	y := scene.Legend.Y

	// VLAN legend
	if len(scene.Legend.VLANs) > 0 {
		b.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="legend-title">VLANs</text>`, x, y))
		b.WriteString("\n")
		y += 4

		for _, vlan := range scene.Legend.VLANs {
			y += 16
			b.WriteString(fmt.Sprintf(`<circle cx="%.0f" cy="%.0f" r="5" fill="%s"/>`,
				x+6, y-4, vlan.Color))
			label := vlan.Label
			if vlan.CIDR != "" {
				label += " - " + vlan.CIDR
			}
			b.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="legend-text">%s</text>`,
				x+18, y, esc(label)))
			b.WriteString("\n")
		}
	}

	// Badge legend
	if len(scene.Legend.BadgeList) > 0 {
		y += 30
		b.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="legend-title">Badges</text>`, x, y))
		b.WriteString("\n")
		y += 4

		for _, badge := range scene.Legend.BadgeList {
			y += 16
			b.WriteString(fmt.Sprintf(`<circle cx="%.0f" cy="%.0f" r="7" fill="%s"/>`,
				x+7, y-4, badge.Color))
			b.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="badge-letter">%s</text>`,
				x+7, y, badge.Letter))
			b.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="legend-text">%s</text>`,
				x+20, y, esc(badge.Label)))
			b.WriteString("\n")
		}
	}
}

func writeSection(b *strings.Builder, sec layout.PositionedSection) {
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="16" fill="%s" stroke="#cdd3da" stroke-width="1"/>`,
		sec.X, sec.Y, sec.W, sec.H, sec.Background))
	b.WriteString("\n")

	if sec.Label != "" {
		labelX := sec.X + 24
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="section-label">%s</text>`,
			labelX, sec.Y+24, esc(sec.Label)))
		b.WriteString("\n")
	}
}

func writeGroup(b *strings.Builder, g layout.PositionedGroup, badges map[string]config.Badge) {
	singleDevice := len(g.Cards) == 1 && len(g.Group.Devices) == 1

	if !singleDevice && !g.HideLabel {
		// group background — use section color if available, else group style
		groupBg := g.Style.Background
		if g.SectionBackground != "" {
			groupBg = g.SectionBackground
		}
		b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%d" fill="%s" stroke="%s" stroke-width="1.5"/>`,
			g.X, g.Y, g.W, g.H, groupRadius, groupBg, g.Style.Border))
		b.WriteString("\n")

		// group label
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="group-label">%s</text>`,
			g.X+12, g.Y+23, esc(g.Label)))
		b.WriteString("\n")
	}

	// cards
	for _, card := range g.Cards {
		writeCard(b, card, badges)
	}
}

func writeCard(b *strings.Builder, card layout.PositionedCard, badges map[string]config.Badge) {
	x, y, w, h := card.X, card.Y, card.W, card.H

	// card background with shadow
	b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%d" fill="#ffffff" stroke="#e9ecef" stroke-width="1" filter="url(#card-shadow)"/>`,
		x, y, w, h, cardRadius))
	b.WriteString("\n")

	// VLAN color stripe on left edge
	if card.VLANColor != "" {
		b.WriteString(fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%d" height="%.1f" rx="2" fill="%s"/>`,
			x, y+4, vlanStripeW, h-8, card.VLANColor))
		b.WriteString("\n")
	}

	contentX := x + 10
	if card.VLANColor != "" {
		contentX = x + vlanStripeW + 8
	}

	// icon
	iconSz := float64(iconSize) * 0.75
	writeIconAt(b, card.Icon, contentX, y+6, iconSz)

	// label (to the right of icon)
	textX := contentX + iconSz + 6
	textMaxW := w - (textX - x) - 8

	labelY := writeWrappedText(b, textX, y+20, card.Label, "card-label", textMaxW, 7.0, 14)

	// details
	detailY := labelY + 2
	if card.Hardware != "" {
		detailY = writeWrappedText(b, textX, detailY, card.Hardware, "card-detail", textMaxW, 5.8, 12)
	}
	if card.OS != "" {
		detailY = writeWrappedText(b, textX, detailY, card.OS, "card-detail", textMaxW, 5.8, 12)
	}
	if card.Specs != "" {
		detailY = writeWrappedText(b, textX, detailY, card.Specs, "card-detail", textMaxW, 5.8, 12)
	}

	// services below details
	if len(card.Services) > 0 {
		detailY += 8
		for _, svc := range card.Services {
			if svc.Icon != "" {
				writeIconAt(b, svc.Icon, contentX, detailY-10, 12)
				b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="service-text">%s</text>`,
					contentX+16, detailY, esc(svc.Name)))
			} else {
				b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="service-text">• %s</text>`,
					contentX+4, detailY, esc(svc.Name)))
			}
			b.WriteString("\n")
			detailY += 16
		}
	}

	if card.IP != "" {
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="card-ip">%s</text>`,
			contentX+4, y+h-8, esc(card.IP)))
	}

	// badges in bottom-right corner as colored circles with letters
	if len(card.Badges) > 0 {
		sorted := make([]string, len(card.Badges))
		copy(sorted, card.Badges)
		sort.Strings(sorted)

		badgeR := 7.0
		bx := x + w - 6
		by := y + h - 6 - badgeR
		for i := len(sorted) - 1; i >= 0; i-- {
			bx -= badgeR*2 + 2
			badge := badges[sorted[i]]
			b.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="%s"/>`,
				bx+badgeR, by, badgeR, badge.Color))
			b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="badge-letter">%s</text>`,
				bx+badgeR, by+4, badge.Letter))
			b.WriteString("\n")
		}
	}
}

func writeIconAt(b *strings.Builder, name string, x, y, size float64) {
	svgContent := icons.Get(name)
	if svgContent == "" {
		svgContent = icons.Get("generic")
	}
	if svgContent == "" {
		return
	}

	inner := stripSVGWrapper(svgContent)
	inner = strings.ReplaceAll(inner, "currentColor", "#495057")

	scale := size / 24.0
	b.WriteString(fmt.Sprintf(`<g transform="translate(%.1f,%.1f) scale(%.3f)" fill="none" stroke="#495057" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">`, x, y, scale))
	b.WriteString(inner)
	b.WriteString("</g>\n")
}

func stripSVGWrapper(svg string) string {
	start := strings.Index(svg, ">")
	if start == -1 {
		return svg
	}
	end := strings.LastIndex(svg, "</svg>")
	if end == -1 {
		return svg[start+1:]
	}
	return svg[start+1 : end]
}

func esc(s string) string {
	return html.EscapeString(s)
}

func wrapText(text string, maxWidth float64, charWidth float64) []string {
	maxChars := int(maxWidth / charWidth)
	if maxChars < 1 {
		maxChars = 1
	}
	if len(text) <= maxChars {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	line := ""
	for _, word := range words {
		if line == "" {
			line = word
		} else if len(line)+1+len(word) <= maxChars {
			line += " " + word
		} else {
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func writeWrappedText(b *strings.Builder, x, y float64, text string, class string, maxWidth float64, charWidth float64, lineHeight float64) float64 {
	lines := wrapText(text, maxWidth, charWidth)
	for _, line := range lines {
		b.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="%s">%s</text>`,
			x, y, class, esc(line)))
		b.WriteString("\n")
		y += lineHeight
	}
	return y
}
