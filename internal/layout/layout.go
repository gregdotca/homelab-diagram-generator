package layout

import (
	"math"
	"sort"

	"github.com/gregdotca/hdg/internal/config"
)

type Scene struct {
	Width       float64
	Height      float64
	Title       string
	FontFamily  string
	Background  string
	Sections    []PositionedSection
	Groups      []PositionedGroup
	Legend      LegendBlock
}

type PositionedSection struct {
	Label          string
	X, Y, W, H    float64
	Background     string
}

type PositionedGroup struct {
	config.Group
	X, Y, W, H      float64
	Cards            []PositionedCard
	HideLabel        bool
	SectionBackground string
}

type PositionedCard struct {
	config.Device
	X, Y, W, H float64
	VLANColor   string
	GroupID     string
}

type LegendBlock struct {
	X, Y      float64
	VLANs     []config.VLAN
	BadgeList []config.Badge
	Badges    map[string]config.Badge
}

const (
	titleHeight     = 40
	legendWidth     = 280
	legendItemH     = 20
	cardHeight      = 85
	serviceRowH     = 18
	badgeSize       = 18
	minGroupWidth   = 260
)

func Compute(cfg *config.Config) *Scene {
	s := &Scene{
		Title:      cfg.Title,
		FontFamily: cfg.Style.FontFamily,
		Background: cfg.Style.Background,
	}

	cardW := cfg.Style.CardWidth
	cardPad := cfg.Style.CardPadding
	groupPad := cfg.Style.GroupPadding
	groupSpacing := cfg.Style.GroupSpacing
	groupVSpacing := groupSpacing + cardPad*2

	type groupMeasure struct {
		group      config.Group
		w, h       float64
		cols       int
		totalCardH float64
	}
	type rowMeasure struct {
		rowNum int
		groups []groupMeasure
		totalW float64
		maxH   float64
	}

	legendH := float64(len(cfg.VLANs))*legendItemH + float64(len(cfg.Badges))*legendItemH + 60

	// Assign row numbers: use explicit `row` field, or auto-assign
	rowBuckets := make(map[int][]config.Group)
	autoRow := 0
	for _, g := range cfg.Groups {
		r := g.Row
		if r == 0 {
			r = autoRow + 1
			autoRow = r
		}
		if r > autoRow {
			autoRow = r
		}
		rowBuckets[r] = append(rowBuckets[r], g)
	}

	// Sort row numbers
	var rowNums []int
	for r := range rowBuckets {
		rowNums = append(rowNums, r)
	}
	sort.Ints(rowNums)

	// Measure all groups, then split into visual sub-lines (max 5 groups per line)
	const maxPerLine = 5

	type logicalRow struct {
		rowNum   int
		subLines []rowMeasure
	}

	var logicalRows []logicalRow
	var overallMaxW float64

	for _, rn := range rowNums {
		groups := rowBuckets[rn]
		multiDeviceCount := 0
		for _, g := range groups {
			if len(g.Devices) > 1 {
				multiDeviceCount++
			}
		}
		multiGroupSizing := multiDeviceCount > 1
		groupLabelH := 30.0
		if !multiGroupSizing {
			groupLabelH = 0
		}

		var allMeasured []groupMeasure
		for _, g := range groups {
			cols := g.Columns
			if cols == 0 {
				cols = bestColumns(len(g.Devices), cardW, 1600)
			}
			rowCount := int(math.Ceil(float64(len(g.Devices)) / float64(cols)))

			singleDevice := len(g.Devices) == 1
			var gw float64
			if singleDevice {
				gw = cardW + cardPad*2
			} else {
				gw = float64(cols)*(cardW+groupSpacing) - groupSpacing + groupPad*2
				if gw < minGroupWidth {
					gw = minGroupWidth
				}
			}

			totalCardH := float64(cardHeight)
			for _, dev := range g.Devices {
				h := cardHeightFor(dev)
				if h > totalCardH {
					totalCardH = h
				}
			}

			var gh float64
			if singleDevice {
				gh = totalCardH + cardPad*2
			} else {
				gh = float64(rowCount)*(totalCardH+groupVSpacing) - groupVSpacing + groupPad*2 + groupLabelH
			}

			allMeasured = append(allMeasured, groupMeasure{
				group: g, w: gw, h: gh, cols: cols, totalCardH: totalCardH,
			})
		}

		lr := logicalRow{rowNum: rn}
		for i := 0; i < len(allMeasured); i += maxPerLine {
			end := i + maxPerLine
			if end > len(allMeasured) {
				end = len(allMeasured)
			}
			chunk := allMeasured[i:end]

			rm := rowMeasure{rowNum: rn}
			for _, gm := range chunk {
				rm.groups = append(rm.groups, gm)
				rm.totalW += gm.w
				if gm.h > rm.maxH {
					rm.maxH = gm.h
				}
			}
			rm.totalW += groupSpacing * float64(len(rm.groups)-1)
			if rm.totalW > overallMaxW {
				overallMaxW = rm.totalW
			}
			lr.subLines = append(lr.subLines, rm)
		}
		logicalRows = append(logicalRows, lr)
	}

	legendAreaW := legendWidth + groupPad

	s.Legend = LegendBlock{
		X: groupPad, Y: titleHeight + groupPad,
		VLANs: cfg.VLANs, BadgeList: cfg.BadgeList, Badges: cfg.Badges,
	}

	contentLeft := legendAreaW + groupPad
	currentY := titleHeight + groupPad

	sectionPad := 10.0
	sectionLabelH := 20.0

	sectionColors := []string{
		"#f2f7fd", // blue tint    (matches #2196F3)
		"#f3faf5", // green tint   (matches #4CAF50)
		"#fef9f1", // orange tint  (matches #FF9800)
		"#f8f4fb", // purple tint  (matches #9C27B0)
		"#fdf2f2", // red tint     (matches #F44336)
		"#fdf0f6", // pink tint    (matches #E91E63)
		"#f6f3f1", // brown tint   (matches #795548)
		"#f3f5f7", // grey tint    (matches #607D8B)
		"#fbf0f0", // dark red     (matches #D32F2F)
		"#f5f9f0", // light green  (matches #8BC34A)
	}

	for ri, lr := range logicalRows {
		sectionLabel := ""
		if sec, ok := cfg.SectionMap[lr.rowNum]; ok {
			sectionLabel = sec.Label
		}

		sectionBg := sectionColors[ri%len(sectionColors)]

		sectionY := currentY
		contentY := currentY + sectionPad

		multiDeviceGroups := 0
		for _, sl := range lr.subLines {
			for _, gm := range sl.groups {
				if len(gm.group.Devices) > 1 {
					multiDeviceGroups++
				}
			}
		}
		multiGroup := multiDeviceGroups > 1

		if sectionLabel != "" {
			if multiGroup {
				contentY += sectionLabelH + 13
			} else {
				contentY += sectionLabelH + 4
			}
		}

		extraPad := 0.0
		if multiGroup {
			extraPad = 8.0
		}

		var sectionMaxRight float64
		for si, rm := range lr.subLines {
			startX := contentLeft + extraPad
			currentX := startX

			for _, gm := range rm.groups {
				pg := PositionedGroup{
					Group:             gm.group,
					X: currentX, Y: contentY,
					W: gm.w, H: gm.h,
					HideLabel:         !multiGroup,
					SectionBackground: sectionBg,
				}

				singleDevice := len(gm.group.Devices) == 1
				for di, dev := range gm.group.Devices {
					col := di % gm.cols
					row := di / gm.cols
					var cx, cy float64
					if singleDevice {
						cx = currentX + cardPad
						cy = contentY + cardPad
					} else {
						labelOffset := 30.0
						if pg.HideLabel {
							labelOffset = 0
						}
						cx = currentX + groupPad + float64(col)*(cardW+groupSpacing)
						cy = contentY + labelOffset + groupPad + float64(row)*(gm.totalCardH+groupVSpacing)
					}

					vlanColor := ""
					if dev.VLAN != "" {
						vlanColor = cfg.VLANColor(dev.VLAN)
					}

					cardH := gm.totalCardH
					if singleDevice {
						cardH = cardHeightFor(dev)
					}

					pg.Cards = append(pg.Cards, PositionedCard{
						Device:    dev,
						X: cx, Y: cy,
						W: cardW, H: cardH,
						VLANColor: vlanColor,
						GroupID:   gm.group.ID,
					})
				}

				s.Groups = append(s.Groups, pg)
				groupRight := currentX + gm.w
				if groupRight > sectionMaxRight {
					sectionMaxRight = groupRight
				}
				currentX += gm.w + groupSpacing
			}

			contentY += rm.maxH + groupSpacing
			if si < len(lr.subLines)-1 {
				// no extra label spacing between sub-lines
			}
		}

		// Section height spans all sub-lines
		sectionH := contentY - sectionY - groupSpacing + sectionPad + extraPad

		sectionW := sectionMaxRight - contentLeft + sectionPad*2 + extraPad
		s.Sections = append(s.Sections, PositionedSection{
			Label:      sectionLabel,
			X:          contentLeft - sectionPad,
			Y:          sectionY,
			W:          sectionW,
			H:          sectionH,
			Background: sectionBg,
		})

		if ri == 0 {
			legendBottom := titleHeight + groupPad + legendH + groupPad
			minSectionBottom := legendBottom - groupSpacing
			if sectionY+sectionH < minSectionBottom {
				sectionH = minSectionBottom - sectionY
				s.Sections[len(s.Sections)-1] = PositionedSection{
					Label:      sectionLabel,
					X:          contentLeft - sectionPad,
					Y:          sectionY,
					W:          sectionW,
					H:          sectionH,
					Background: sectionBg,
				}
			}
		}
		currentY = sectionY + sectionH + groupSpacing
	}

	// Shrink canvas width to fit actual content
	var maxRight float64
	for _, sec := range s.Sections {
		right := sec.X + sec.W
		if right > maxRight {
			maxRight = right
		}
	}
	s.Width = maxRight + groupPad
	s.Height = currentY + groupPad

	return s
}


func bestColumns(deviceCount int, cardW, maxWidth float64) int {
	if deviceCount <= 2 {
		return deviceCount
	}
	maxCols := int(maxWidth / (cardW + 12))
	if maxCols < 1 {
		maxCols = 1
	}
	if deviceCount <= maxCols {
		return deviceCount
	}
	best := 2
	for c := 2; c <= maxCols; c++ {
		rows := int(math.Ceil(float64(deviceCount) / float64(c)))
		if rows <= 9 {
			best = c
			break
		}
	}
	return best
}

func wrappedLineCount(text string, maxChars int) int {
	if maxChars < 1 {
		maxChars = 1
	}
	if len(text) <= maxChars {
		return 1
	}
	lines := 1
	lineLen := 0
	for _, word := range splitWords(text) {
		wl := len(word)
		if lineLen == 0 {
			lineLen = wl
		} else if lineLen+1+wl <= maxChars {
			lineLen += 1 + wl
		} else {
			lines++
			lineLen = wl
		}
	}
	return lines
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func cardHeightFor(dev config.Device) float64 {
	// available text width: cardW(220) - vlanStripe(4) - pad(8) - icon(30) - gap(6) - rightPad(8) ≈ 164px
	// label chars: 164 / 7.0 ≈ 23, detail chars: 164 / 5.8 ≈ 28
	labelMaxChars := 23
	detailMaxChars := 28

	extraLines := 0
	extraLines += wrappedLineCount(dev.Label, labelMaxChars) - 1
	if dev.Hardware != "" {
		extraLines += wrappedLineCount(dev.Hardware, detailMaxChars) - 1
	}
	if dev.OS != "" {
		extraLines += wrappedLineCount(dev.OS, detailMaxChars) - 1
	}
	if dev.Specs != "" {
		extraLines += wrappedLineCount(dev.Specs, detailMaxChars) - 1
	}

	h := cardHeight + extraLines*12
	if len(dev.Services) > 0 {
		h += len(dev.Services) * int(serviceRowH)
	}
	return float64(h)
}


