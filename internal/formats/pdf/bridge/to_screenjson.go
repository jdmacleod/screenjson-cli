// Package bridge provides conversion between PDF and ScreenJSON.
// Note: PDF decoding is handled directly in codec/decode.go as it produces
// a ScreenJSON document directly (no intermediate PDF model).
package bridge

// PDF to ScreenJSON conversion is handled in codec/decode.go
// because Poppler produces an intermediate format that is directly
// converted to ScreenJSON without an intermediate PDF model.
