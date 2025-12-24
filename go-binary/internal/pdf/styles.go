package pdf

import (
	"fmt"

	"github.com/signintech/gopdf"
)

// PageSize represents standard page dimensions in points (1 inch = 72 points)
type PageSize struct {
	Width  float64
	Height float64
}

// Standard page sizes
var (
	PageA4      = PageSize{Width: 595.28, Height: 841.89}
	PageLetter  = PageSize{Width: 612, Height: 792}
	PageLegal   = PageSize{Width: 612, Height: 1008}
	PageA3      = PageSize{Width: 841.89, Height: 1190.55}
	PageTabloid = PageSize{Width: 792, Height: 1224} // 11 x 17 inches - best for wide tables
)

// Orientation constants
type Orientation string

const (
	Portrait  Orientation = "portrait"
	Landscape Orientation = "landscape"
)

// Alignment constants
const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

// Color represents RGB color values
type Color struct {
	R, G, B uint8
}

// Predefined colors for styling
var (
	ColorBlack      = Color{0, 0, 0}
	ColorWhite      = Color{255, 255, 255}
	ColorGray       = Color{128, 128, 128}
	ColorLightGray  = Color{240, 240, 240}
	ColorDarkGray   = Color{64, 64, 64}
	ColorBlue       = Color{0, 102, 204}
	ColorLightBlue  = Color{230, 242, 255}
	ColorGreen      = Color{0, 153, 76}
	ColorLightGreen = Color{230, 255, 238}
)

// Style represents text and cell styling options
type Style struct {
	FontFamily    string
	FontSize      float64
	FontStyle     string // "", "B", "I", "BI"
	TextColor     Color
	FillColor     Color
	BorderColor   Color
	BorderWidth   float64
	Alignment     int // 0=Left, 1=Center, 2=Right
	Padding       float64
	LineHeight    float64
	HasBackground bool
	HasBorder     bool
}

// DefaultStyle returns the default text style
func DefaultStyle() Style {
	return Style{
		FontFamily:    "Arial",
		FontSize:      10,
		FontStyle:     "",
		TextColor:     ColorBlack,
		FillColor:     ColorWhite,
		BorderColor:   ColorGray,
		BorderWidth:   0.5,
		Alignment:     0,
		Padding:       4,
		LineHeight:    1.2,
		HasBackground: false,
		HasBorder:     true,
	}
}

// TableStyle returns a standard style for table cells
func TableStyle() Style {
	s := DefaultStyle()
	s.HasBorder = true
	s.BorderColor = ColorDarkGray
	s.BorderWidth = 0.5
	return s
}

// HeaderStyle returns style for table headers
func HeaderStyle() Style {
	s := DefaultStyle()
	s.FontStyle = "B"
	s.FontSize = 11
	s.FillColor = ColorLightBlue
	s.TextColor = ColorDarkGray
	s.HasBackground = true
	return s
}

// ParseHexColor parses a hex color string (e.g. "FFFFFF" or "FF0000") to Color
func ParseHexColor(hex string) Color {
	if len(hex) == 6 {
		var r, g, b uint8
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		return Color{r, g, b}
	}
	return ColorBlack
}

// AlternatingRowStyle returns style for alternating rows
func AlternatingRowStyle(isEven bool) Style {
	s := DefaultStyle()
	if isEven {
		s.FillColor = ColorLightGray
		s.HasBackground = true
	}
	return s
}

// Options contains all conversion options
type Options struct {
	PageSize     PageSize
	Orientation  Orientation
	FontFamily   string
	FontSize     float64
	Margin       float64
	HeaderRow    bool
	AutoWidth    bool
	Title        string
	Author       string
	Subject      string
	Compression  bool
	Quality      string // "fast", "balanced", "best"
	HeaderText   string
	FooterText   string
}

// DefaultOptions returns sensible default options
func DefaultOptions() Options {
	return Options{
		PageSize:    PageA4,
		Orientation: Portrait,
		FontFamily:  "Arial",
		FontSize:    10,
		Margin:      20,
		HeaderRow:   true,
		AutoWidth:   true,
		Compression: true,
		Quality:     "balanced",
	}
}

// GetPageRect returns the gopdf.Rect for the configured page size and orientation
func (o Options) GetPageRect() *gopdf.Rect {
	w, h := o.PageSize.Width, o.PageSize.Height
	if o.Orientation == Landscape {
		w, h = h, w
	}
	return &gopdf.Rect{W: w, H: h}
}

// ContentWidth returns the usable content width after margins
func (o Options) ContentWidth() float64 {
	w := o.PageSize.Width
	if o.Orientation == Landscape {
		w = o.PageSize.Height
	}
	return w - (o.Margin * 2)
}

// ContentHeight returns the usable content height after margins
func (o Options) ContentHeight() float64 {
	h := o.PageSize.Height
	if o.Orientation == Landscape {
		h = o.PageSize.Width
	}
	return h - (o.Margin * 2)
}
