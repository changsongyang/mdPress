// Package colorutil provides the shared CSS-color helpers mdpress uses to pick
// legible ink over a configured background. It is the single home for the
// luminance heuristic and the safe-color validation that internal/cover and
// internal/output/epub previously duplicated (and had let drift apart).
package colorutil

import (
	"regexp"
	"strconv"
	"strings"
)

// luminanceThreshold is the perceived luminance cutoff (0-255) separating a
// "light" background — which wants dark ink — from a dark one. 186 is the
// common midpoint used by WCAG-style contrast pickers.
const luminanceThreshold = 186

// cssColorPattern matches safe CSS color values (hex, rgb, rgba, hsl, hsla,
// named colors). It is intentionally strict so a configured background can be
// interpolated into generated CSS without escaping concerns.
var cssColorPattern = regexp.MustCompile(`^(?i)(?:#[0-9a-f]{3,8}|(?:rgb|rgba|hsl|hsla)\([\d\s,%.]+\)|[a-z]{1,30})$`)

// IsSafeColor reports whether s is a CSS color value safe to interpolate into a
// stylesheet (hex, rgb[a], hsl[a], or a bare named color). Leading and trailing
// whitespace is ignored.
func IsSafeColor(s string) bool {
	return cssColorPattern.MatchString(strings.TrimSpace(s))
}

// namedColorLight classifies common CSS named colors as perceptually light
// (true) or dark (false). Names absent from the map are assumed dark, which
// keeps light text as the safe default for unknown backgrounds.
var namedColorLight = map[string]bool{
	// Light backgrounds -> dark ink.
	"white": true, "ivory": true, "snow": true, "beige": true,
	"linen": true, "seashell": true, "floralwhite": true, "ghostwhite": true,
	"whitesmoke": true, "lightyellow": true, "lightgray": true,
	"lightgrey": true, "gainsboro": true, "aliceblue": true,
	"antiquewhite": true, "azure": true, "cornsilk": true, "honeydew": true,
	"lavenderblush": true, "lemonchiffon": true, "mintcream": true,
	"oldlace": true, "papayawhip": true, "wheat": true,
	// Dark anchors, documented for clarity (any unlisted name is also
	// treated as dark).
	"black": false, "navy": false, "maroon": false, "midnightblue": false,
	"darkblue": false, "darkslategray": false, "darkslategrey": false,
}

// IsLight reports whether the given CSS color is perceptually light. It
// understands hex colors (#rgb, #rgba, #rrggbb, #rrggbbaa), common named CSS
// colors, and numeric rgb()/rgba() forms. Unknown or unparseable formats
// (including hsl()) are assumed dark so that light text remains the safer
// default. Alpha channels are ignored.
func IsLight(color string) bool {
	color = strings.TrimSpace(color)
	if strings.HasPrefix(color, "#") {
		return isLightHex(color[1:])
	}
	lower := strings.ToLower(color)
	if light, ok := namedColorLight[lower]; ok {
		return light
	}
	if r, g, b, ok := parseRGBFunc(lower); ok {
		return luminance(r, g, b) > luminanceThreshold
	}
	return false
}

// isLightHex reports whether a hex color body (without the leading '#') is
// perceptually light. Alpha channels are ignored.
func isLightHex(hex string) bool {
	// Expand shorthand (#rgb -> #rrggbb, #rgba -> #rrggbb).
	if len(hex) == 3 || len(hex) == 4 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	// Strip alpha channel from #rrggbbaa.
	if len(hex) == 8 {
		hex = hex[:6]
	}
	if len(hex) < 6 {
		return false
	}
	r := hexVal(hex[0])*16 + hexVal(hex[1])
	g := hexVal(hex[2])*16 + hexVal(hex[3])
	b := hexVal(hex[4])*16 + hexVal(hex[5])
	return luminance(float64(r), float64(g), float64(b)) > luminanceThreshold
}

// parseRGBFunc parses numeric rgb()/rgba() color functions. It accepts both
// the legacy comma syntax (rgb(255, 250, 240)) and the modern space syntax
// (rgb(255 250 240 / 0.5)); percentage components are scaled to the 0-255
// range. The alpha channel is ignored. The input must already be lowercase.
func parseRGBFunc(color string) (r, g, b float64, ok bool) {
	var body string
	switch {
	case strings.HasPrefix(color, "rgba(") && strings.HasSuffix(color, ")"):
		body = color[len("rgba(") : len(color)-1]
	case strings.HasPrefix(color, "rgb(") && strings.HasSuffix(color, ")"):
		body = color[len("rgb(") : len(color)-1]
	default:
		return 0, 0, 0, false
	}
	body = strings.NewReplacer(",", " ", "/", " ").Replace(body)
	fields := strings.Fields(body)
	if len(fields) < 3 {
		return 0, 0, 0, false
	}
	var channels [3]float64
	for i := 0; i < 3; i++ {
		f := fields[i]
		percent := strings.HasSuffix(f, "%")
		f = strings.TrimSuffix(f, "%")
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, 0, 0, false
		}
		if percent {
			v = v * 255 / 100
		}
		channels[i] = v
	}
	return channels[0], channels[1], channels[2], true
}

// luminance computes perceived luminance (ITU-R BT.601) on a 0-255 scale:
// Y = 0.299R + 0.587G + 0.114B.
func luminance(r, g, b float64) float64 {
	return 0.299*r + 0.587*g + 0.114*b
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return 0
	}
}
