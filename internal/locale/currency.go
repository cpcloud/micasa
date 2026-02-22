// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package locale

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Currency holds resolved currency formatting state. It is safe for
// concurrent read access but should be treated as immutable after creation.
type Currency struct {
	unit   currency.Unit
	tag    language.Tag
	symbol string
	prefix bool // true if symbol comes before the number
	code   string
}

var (
	ErrInvalidMoney  = errors.New("invalid money value")
	ErrNegativeMoney = errors.New("negative money value")
)

// Resolve creates a Currency from an ISO 4217 code (e.g. "USD", "EUR").
// Returns an error if the code is not a valid ISO 4217 currency.
func Resolve(code string) (Currency, error) {
	if code == "" {
		code = "USD"
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	unit, err := currency.ParseISO(code)
	if err != nil {
		return Currency{}, fmt.Errorf("unknown currency %q: %w", code, err)
	}
	tag := localeForCurrency(code)
	sym, pre := extractSymbol(unit, tag)
	return Currency{
		unit:   unit,
		tag:    tag,
		symbol: sym,
		prefix: pre,
		code:   code,
	}, nil
}

// MustResolve is like Resolve but panics on error.
func MustResolve(code string) Currency {
	c, err := Resolve(code)
	if err != nil {
		panic(err)
	}
	return c
}

// DefaultCurrency returns USD with standard US English formatting.
func DefaultCurrency() Currency {
	return MustResolve("USD")
}

// ResolveDefault resolves the currency code using the config layering:
// explicit code > MICASA_CURRENCY env > LC_MONETARY/LANG auto-detect > USD.
func ResolveDefault(configured string) (Currency, error) {
	code := configured
	if code == "" {
		code = os.Getenv("MICASA_CURRENCY")
	}
	if code == "" {
		code = detectCurrencyFromLocale()
	}
	if code == "" {
		code = "USD"
	}
	return Resolve(code)
}

// Code returns the ISO 4217 code (e.g. "USD", "EUR").
func (c Currency) Code() string {
	return c.code
}

// Symbol returns the narrow symbol glyph (e.g. "$", "EUR", "GBP", "JPY").
func (c Currency) Symbol() string {
	return c.symbol
}

// FormatCents formats an int64 cent value as a locale-appropriate currency
// string. Uses the locale's number grouping and decimal separator, with the
// currency symbol placed per locale convention (no extra space).
func (c Currency) FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		if cents == math.MinInt64 {
			cents = math.MaxInt64
		} else {
			cents = -cents
		}
	}
	dollars := cents / 100
	remainder := cents % 100
	p := message.NewPrinter(c.tag)
	numStr := p.Sprintf("%d", dollars)
	_, dec := c.separators()
	number := fmt.Sprintf("%s%s%02d", numStr, dec, remainder)
	if c.prefix {
		return sign + c.symbol + number
	}
	return sign + number + "\u00a0" + c.symbol
}

// FormatOptionalCents formats a *int64 cent value, returning "" for nil.
func (c Currency) FormatOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return c.FormatCents(*cents)
}

// FormatCompactCents formats cents using abbreviated notation for large
// values (e.g. 1.2k, 45k, 1.3M) with the correct currency symbol.
// Values under 1,000 in the base unit use full precision.
func (c Currency) FormatCompactCents(cents int64) string {
	sign := ""
	absCents := cents
	if cents < 0 {
		sign = "-"
		if cents == math.MinInt64 {
			absCents = math.MaxInt64
		} else {
			absCents = -cents
		}
	}
	dollars := float64(absCents) / 100.0
	if dollars < 1000 {
		if sign != "" {
			return sign + c.FormatCents(absCents)
		}
		return c.FormatCents(cents)
	}
	si := humanize.SIWithDigits(dollars, 1, "")
	si = strings.Replace(si, " ", "", 1)
	if c.prefix {
		return sign + c.symbol + si
	}
	return sign + si + "\u00a0" + c.symbol
}

// FormatCompactOptionalCents formats optional cents compactly.
func (c Currency) FormatCompactOptionalCents(cents *int64) string {
	if cents == nil {
		return ""
	}
	return c.FormatCompactCents(*cents)
}

// ParseRequiredCents parses a user-entered money string into cents.
// Strips the currency symbol if present; bare numbers always accepted.
func (c Currency) ParseRequiredCents(input string) (int64, error) {
	cents, err := c.parseCents(strings.TrimSpace(input))
	if err != nil {
		return 0, err
	}
	return cents, nil
}

// ParseOptionalCents parses an optional money string. Returns (nil, nil) for
// empty input.
func (c Currency) ParseOptionalCents(input string) (*int64, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	cents, err := c.parseCents(trimmed)
	if err != nil {
		return nil, err
	}
	return &cents, nil
}

func (c Currency) parseCents(input string) (int64, error) {
	// Normalize locale-specific separators: remove grouping separators,
	// replace decimal comma with period.
	clean := c.normalizeNumber(input)

	if strings.HasPrefix(clean, "-") {
		return 0, ErrNegativeMoney
	}
	// Strip the currency symbol (could be prefix or suffix).
	clean = strings.TrimPrefix(clean, c.symbol)
	clean = strings.TrimSuffix(clean, c.symbol)
	clean = strings.TrimSpace(clean)
	// Also strip $ as a universal fallback (for pasted values).
	clean = strings.TrimPrefix(clean, "$")
	if clean == "" {
		return 0, ErrInvalidMoney
	}
	parts := strings.Split(clean, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidMoney
	}
	wholePart, err := parseDigits(parts[0], true)
	if err != nil {
		return 0, ErrInvalidMoney
	}
	const maxDollars = math.MaxInt64 / 100
	if wholePart > maxDollars {
		return 0, ErrInvalidMoney
	}
	frac := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, ErrInvalidMoney
		}
		frac, err = parseDigits(parts[1], false)
		if err != nil {
			return 0, ErrInvalidMoney
		}
		if len(parts[1]) == 1 {
			frac *= 10
		}
	}
	cents := wholePart*100 + frac
	if cents < 0 {
		return 0, ErrInvalidMoney
	}
	return cents, nil
}

// normalizeNumber removes locale-specific grouping separators and replaces
// the locale-specific decimal separator with ".".
func (c Currency) normalizeNumber(input string) string {
	// Determine grouping and decimal separators from the locale by
	// formatting a known value and inspecting the result.
	group, decimal := c.separators()
	// Remove all grouping separators.
	result := strings.ReplaceAll(input, group, "")
	// Replace decimal separator with ".".
	if decimal != "." {
		result = strings.Replace(result, decimal, ".", 1)
	}
	return result
}

// separators returns the grouping and decimal separators for the currency's
// locale. Derived by formatting a known number and inspecting the output.
func (c Currency) separators() (group, decimal string) {
	p := message.NewPrinter(c.tag)
	// Format a number with known grouping and decimal.
	// 1234.5 in en-US → "1,234.5" (group=",", decimal=".")
	// 1234.5 in de    → "1.234,5" (group=".", decimal=",")
	formatted := p.Sprintf("%.1f", 1234.5)
	// Find the decimal separator (always at position of last non-digit
	// that's between the 4th and 5th digits from the right).
	// Simpler: look for the separator before the last digit.
	// In "1,234.5" → decimal is ".", group is ","
	// In "1.234,5" → decimal is ",", group is "."
	// In "1 234,5" → decimal is ",", group is " "
	//
	// The last non-digit character before trailing digits is the decimal separator.
	lastNonDigit := -1
	for i := len(formatted) - 1; i >= 0; i-- {
		if formatted[i] < '0' || formatted[i] > '9' {
			lastNonDigit = i
			break
		}
	}
	if lastNonDigit < 0 {
		return ",", "."
	}
	decimal = string(formatted[lastNonDigit])
	// The grouping separator is the other non-digit character.
	for i := 0; i < lastNonDigit; i++ {
		if formatted[i] < '0' || formatted[i] > '9' {
			return string(formatted[i]), decimal
		}
	}
	// No grouping separator found (number too small or locale doesn't group).
	if decimal == "." {
		return ",", "."
	}
	return ".", decimal
}

func parseDigits(input string, allowEmpty bool) (int64, error) {
	if input == "" {
		if allowEmpty {
			return 0, nil
		}
		return 0, ErrInvalidMoney
	}
	for _, r := range input {
		if r < '0' || r > '9' {
			return 0, ErrInvalidMoney
		}
	}
	return strconv.ParseInt(input, 10, 64)
}

// extractSymbol formats a zero amount and extracts the symbol glyph and its
// position (prefix or suffix) from the CLDR-formatted output.
func extractSymbol(unit currency.Unit, tag language.Tag) (sym string, prefix bool) {
	p := message.NewPrinter(tag)
	formatted := p.Sprint(currency.NarrowSymbol(unit.Amount(0)))
	// Find the first and last digit positions.
	firstDigit := strings.IndexAny(formatted, "0123456789")
	lastDigit := strings.LastIndexAny(formatted, "0123456789")
	if firstDigit < 0 {
		return unit.String(), true
	}
	pre := strings.TrimSpace(formatted[:firstDigit])
	suf := strings.TrimSpace(formatted[lastDigit+1:])
	if pre != "" {
		return pre, true
	}
	if suf != "" {
		return suf, false
	}
	return unit.String(), true
}

// localeForCurrency maps a currency code to its canonical formatting locale.
// This determines number grouping, decimal separator, and symbol placement.
var currencyLocales = map[string]language.Tag{
	"USD": language.AmericanEnglish,
	"CAD": language.MustParse("en-CA"),
	"GBP": language.BritishEnglish,
	"EUR": language.German,
	"JPY": language.Japanese,
	"CHF": language.MustParse("de-CH"),
	"AUD": language.MustParse("en-AU"),
	"NZD": language.MustParse("en-NZ"),
	"SEK": language.MustParse("sv"),
	"NOK": language.MustParse("nb"),
	"DKK": language.MustParse("da"),
	"PLN": language.MustParse("pl"),
	"CZK": language.MustParse("cs"),
	"HUF": language.MustParse("hu"),
	"INR": language.MustParse("en-IN"),
	"KRW": language.Korean,
	"CNY": language.SimplifiedChinese,
	"BRL": language.MustParse("pt-BR"),
	"MXN": language.MustParse("es-MX"),
	"ZAR": language.MustParse("en-ZA"),
	"TRY": language.Turkish,
	"RUB": language.MustParse("ru"),
	"ILS": language.MustParse("he"),
	"THB": language.MustParse("th"),
	"PHP": language.MustParse("fil"),
	"IDR": language.MustParse("id"),
	"MYR": language.MustParse("ms"),
	"SGD": language.MustParse("en-SG"),
	"HKD": language.MustParse("en-HK"),
	"TWD": language.MustParse("zh-TW"),
	"ARS": language.MustParse("es-AR"),
	"CLP": language.MustParse("es-CL"),
	"COP": language.MustParse("es-CO"),
	"PEN": language.MustParse("es-PE"),
	"AED": language.MustParse("ar-AE"),
	"SAR": language.MustParse("ar-SA"),
	"EGP": language.MustParse("ar-EG"),
}

func localeForCurrency(code string) language.Tag {
	if tag, ok := currencyLocales[code]; ok {
		return tag
	}
	return language.AmericanEnglish
}

// detectCurrencyFromLocale tries to determine the currency from environment
// locale variables (LC_MONETARY, LC_ALL, LANG).
func detectCurrencyFromLocale() string {
	for _, key := range []string{"LC_MONETARY", "LC_ALL", "LANG"} {
		if val := os.Getenv(key); val != "" {
			if code := currencyFromLocaleString(val); code != "" {
				return code
			}
		}
	}
	return ""
}

// currencyFromLocaleString extracts a currency from a locale string like
// "de_DE.UTF-8" by parsing the region and looking up its currency.
func currencyFromLocaleString(locale string) string {
	// Strip encoding suffix (e.g. ".UTF-8").
	if idx := strings.IndexByte(locale, '.'); idx >= 0 {
		locale = locale[:idx]
	}
	// Strip modifier (e.g. "@euro").
	if idx := strings.IndexByte(locale, '@'); idx >= 0 {
		locale = locale[:idx]
	}
	// Convert underscore to hyphen for BCP 47.
	locale = strings.ReplaceAll(locale, "_", "-")
	tag, err := language.Parse(locale)
	if err != nil {
		return ""
	}
	unit, conf := currency.FromTag(tag)
	if conf == language.No {
		return ""
	}
	return unit.String()
}
