// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mattermost/mattermost-server/v6/shared/i18n"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

const (
	LowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	UppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NUMBERS          = "0123456789"
	SYMBOLS          = " !\"\\#$%&'()*+,-./:;<=>?@[]^_`|~"
	BinaryParamKey   = "MM_BINARY_PARAMETERS"
	NoTranslation    = "<untranslated>"
	maxPropSizeBytes = 1024 * 1024
)

var ErrMaxPropSizeExceeded = fmt.Errorf("max prop size of %d exceeded", maxPropSizeBytes)

type StringInterface map[string]any
type StringArray []string

func (sa StringArray) Remove(input string) StringArray {
	for index := range sa {
		if sa[index] == input {
			ret := make(StringArray, 0, len(sa)-1)
			ret = append(ret, sa[:index]...)
			return append(ret, sa[index+1:]...)
		}
	}
	return sa
}

func (sa StringArray) Contains(input string) bool {
	for index := range sa {
		if sa[index] == input {
			return true
		}
	}

	return false
}
func (sa StringArray) Equals(input StringArray) bool {

	if len(sa) != len(input) {
		return false
	}

	for index := range sa {

		if sa[index] != input[index] {
			return false
		}
	}

	return true
}

// Value converts StringArray to database value
func (sa StringArray) Value() (driver.Value, error) {
	sz := 0
	for i := range sa {
		sz += len(sa[i])
		if sz > maxPropSizeBytes {
			return nil, ErrMaxPropSizeExceeded
		}
	}

	j, err := json.Marshal(sa)
	if err != nil {
		return nil, err
	}
	// non utf8 characters are not supported https://mattermost.atlassian.net/browse/MM-41066
	return string(j), err
}

// Scan converts database column value to StringArray
func (sa *StringArray) Scan(value any) error {
	if value == nil {
		return nil
	}

	buf, ok := value.([]byte)
	if ok {
		return json.Unmarshal(buf, sa)
	}

	str, ok := value.(string)
	if ok {
		return json.Unmarshal([]byte(str), sa)
	}

	return errors.New("received value is neither a byte slice nor string")
}

// Scan converts database column value to StringMap
func (m *StringMap) Scan(value any) error {
	if value == nil {
		return nil
	}

	buf, ok := value.([]byte)
	if ok {
		return json.Unmarshal(buf, m)
	}

	str, ok := value.(string)
	if ok {
		return json.Unmarshal([]byte(str), m)
	}

	return errors.New("received value is neither a byte slice nor string")
}

// Value converts StringMap to database value
func (m StringMap) Value() (driver.Value, error) {
	ok := m[BinaryParamKey]
	delete(m, BinaryParamKey)

	sz := 0
	for k := range m {
		sz += len(k) + len(m[k])
		if sz > maxPropSizeBytes {
			return nil, ErrMaxPropSizeExceeded
		}
	}

	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if ok == "true" {
		return append([]byte{0x01}, buf...), nil
	} else if ok == "false" {
		return buf, nil
	}
	// Key wasn't found. We fall back to the default case.
	return string(buf), nil
}

func (StringMap) ImplementsGraphQLType(name string) bool {
	return name == "StringMap"
}

func (m StringMap) MarshalJSON() ([]byte, error) {
	return json.Marshal((map[string]string)(m))
}

func (m *StringMap) UnmarshalGraphQL(input any) error {
	json, ok := input.(map[string]string)
	if !ok {
		return errors.New("wrong type")
	}

	*m = json
	return nil
}

func (si *StringInterface) Scan(value any) error {
	if value == nil {
		return nil
	}

	buf, ok := value.([]byte)
	if ok {
		return json.Unmarshal(buf, si)
	}

	str, ok := value.(string)
	if ok {
		return json.Unmarshal([]byte(str), si)
	}

	return errors.New("received value is neither a byte slice nor string")
}

// Value converts StringInterface to database value
func (si StringInterface) Value() (driver.Value, error) {
	j, err := json.Marshal(si)
	if err != nil {
		return nil, err
	}

	if len(j) > maxPropSizeBytes {
		return nil, ErrMaxPropSizeExceeded
	}

	// non utf8 characters are not supported https://mattermost.atlassian.net/browse/MM-41066
	return string(j), err
}

func (StringInterface) ImplementsGraphQLType(name string) bool {
	return name == "StringInterface"
}

func (si StringInterface) MarshalJSON() ([]byte, error) {
	return json.Marshal((map[string]any)(si))
}

func (si *StringInterface) UnmarshalGraphQL(input any) error {
	json, ok := input.(map[string]any)
	if !ok {
		return errors.New("wrong type")
	}

	*si = json
	return nil
}

var translateFunc i18n.TranslateFunc
var translateFuncOnce sync.Once

func AppErrorInit(t i18n.TranslateFunc) {
	translateFuncOnce.Do(func() {
		translateFunc = t
	})
}

type AppError struct {
	Id            string `json:"id"`
	Message       string `json:"message"`               // Message to be display to the end user without debugging information
	DetailedError string `json:"detailed_error"`        // Internal error string to help the developer
	RequestId     string `json:"request_id,omitempty"`  // The RequestId that's also set in the header
	StatusCode    int    `json:"status_code,omitempty"` // The http status code
	Where         string `json:"-"`                     // The function where it happened in the form of Struct.Func
	IsOAuth       bool   `json:"is_oauth,omitempty"`    // Whether the error is OAuth specific
	params        map[string]any
	wrapped       error
}

const maxErrorLength = 1024

func (er *AppError) Error() string {
	var sb strings.Builder

	// render the error information
	sb.WriteString(er.Where)
	sb.WriteString(": ")
	if er.Message != NoTranslation {
		sb.WriteString(er.Message)
	}

	// only render the detailed error when it's present
	if er.DetailedError != "" {
		if er.Message != NoTranslation {
			sb.WriteString(", ")
		}
		sb.WriteString(er.DetailedError)
	}

	// render the wrapped error
	err := er.wrapped
	if err != nil {
		sb.WriteString(", ")
		sb.WriteString(err.Error())
	}

	res := sb.String()
	if len(res) > maxErrorLength {
		res = res[:maxErrorLength] + "..."
	}
	return res
}

func (er *AppError) Translate(T i18n.TranslateFunc) {
	if T == nil {
		er.Message = er.Id
		return
	}

	if er.params == nil {
		er.Message = T(er.Id)
	} else {
		er.Message = T(er.Id, er.params)
	}
}

func (er *AppError) SystemMessage(T i18n.TranslateFunc) string {
	if er.params == nil {
		return T(er.Id)
	}
	return T(er.Id, er.params)
}

func (er *AppError) ToJSON() string {
	// turn the wrapped error into a detailed message
	detailed := er.DetailedError
	defer func() {
		er.DetailedError = detailed
	}()

	er.wrappedToDetailed()

	b, _ := json.Marshal(er)
	return string(b)
}

func (er *AppError) wrappedToDetailed() {
	if er.wrapped == nil {
		return
	}

	if er.DetailedError != "" {
		er.DetailedError += ", "
	}

	er.DetailedError += er.wrapped.Error()
}

func (er *AppError) Unwrap() error {
	return er.wrapped
}

func (er *AppError) Wrap(err error) *AppError {
	er.wrapped = err
	return er
}

// AppErrorFromJSON will decode the input and return an AppError
func AppErrorFromJSON(data io.Reader) *AppError {
	str := ""
	bytes, rerr := io.ReadAll(data)
	if rerr != nil {
		str = rerr.Error()
	} else {
		str = string(bytes)
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	var er AppError
	err := decoder.Decode(&er)
	if err != nil {
		return NewAppError("AppErrorFromJSON", "model.utils.decode_json.app_error", nil, "body: "+str, http.StatusInternalServerError).Wrap(err)
	}
	return &er
}

func NewAppError(where string, id string, params map[string]any, details string, status int) *AppError {
	ap := &AppError{
		Id:            id,
		params:        params,
		Message:       id,
		Where:         where,
		DetailedError: details,
		StatusCode:    status,
		IsOAuth:       false,
	}
	ap.Translate(translateFunc)
	return ap
}

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769").WithPadding(base32.NoPadding)

// NewId is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// without the padding.
func NewId() string {
	return encoding.EncodeToString(uuid.NewRandom())
}

// NewRandomTeamName is a NewId that will be a valid team name.
func NewRandomTeamName() string {
	teamName := NewId()
	for IsReservedTeamName(teamName) {
		teamName = NewId()
	}
	return teamName
}

// NewRandomString returns a random string of the given length.
// The resulting entropy will be (5 * length) bits.
func NewRandomString(length int) string {
	data := make([]byte, 1+(length*5/8))
	rand.Read(data)
	return encoding.EncodeToString(data)[:length]
}

// GetMillis is a convenience method to get milliseconds since epoch.
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetMillisForTime is a convenience method to get milliseconds since epoch for provided Time.
func GetMillisForTime(thisTime time.Time) int64 {
	return thisTime.UnixNano() / int64(time.Millisecond)
}

// GetTimeForMillis is a convenience method to get time.Time for milliseconds since epoch.
func GetTimeForMillis(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}

// PadDateStringZeros is a convenience method to pad 2 digit date parts with zeros to meet ISO 8601 format
func PadDateStringZeros(dateString string) string {
	parts := strings.Split(dateString, "-")
	for index, part := range parts {
		if len(part) == 1 {
			parts[index] = "0" + part
		}
	}
	dateString = strings.Join(parts[:], "-")
	return dateString
}

// GetStartOfDayMillis is a convenience method to get milliseconds since epoch for provided date's start of day
func GetStartOfDayMillis(thisTime time.Time, timeZoneOffset int) int64 {
	localSearchTimeZone := time.FixedZone("Local Search Time Zone", timeZoneOffset)
	resultTime := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, localSearchTimeZone)
	return GetMillisForTime(resultTime)
}

// GetEndOfDayMillis is a convenience method to get milliseconds since epoch for provided date's end of day
func GetEndOfDayMillis(thisTime time.Time, timeZoneOffset int) int64 {
	localSearchTimeZone := time.FixedZone("Local Search Time Zone", timeZoneOffset)
	resultTime := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 23, 59, 59, 999999999, localSearchTimeZone)
	return GetMillisForTime(resultTime)
}

func CopyStringMap(originalMap map[string]string) map[string]string {
	copyMap := make(map[string]string, len(originalMap))
	for k, v := range originalMap {
		copyMap[k] = v
	}
	return copyMap
}

// MapToJSON converts a map to a json string
func MapToJSON(objmap map[string]string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

// MapBoolToJSON converts a map to a json string
func MapBoolToJSON(objmap map[string]bool) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

// MapFromJSON will decode the key/value pair map
func MapFromJSON(data io.Reader) map[string]string {
	var objmap map[string]string

	json.NewDecoder(data).Decode(&objmap)
	if objmap == nil {
		return make(map[string]string)
	}

	return objmap
}

// MapFromJSON will decode the key/value pair map
func MapBoolFromJSON(data io.Reader) map[string]bool {
	var objmap map[string]bool

	json.NewDecoder(data).Decode(&objmap)
	if objmap == nil {
		return make(map[string]bool)
	}

	return objmap
}

func ArrayToJSON(objmap []string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func ArrayFromJSON(data io.Reader) []string {
	var objmap []string

	json.NewDecoder(data).Decode(&objmap)
	if objmap == nil {
		return make([]string, 0)
	}

	return objmap
}

func ArrayFromInterface(data any) []string {
	stringArray := []string{}

	dataArray, ok := data.([]any)
	if !ok {
		return stringArray
	}

	for _, v := range dataArray {
		if str, ok := v.(string); ok {
			stringArray = append(stringArray, str)
		}
	}

	return stringArray
}

func StringInterfaceToJSON(objmap map[string]any) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func StringInterfaceFromJSON(data io.Reader) map[string]any {
	var objmap map[string]any

	json.NewDecoder(data).Decode(&objmap)
	if objmap == nil {
		return make(map[string]any)
	}

	return objmap
}

// ToJSON serializes an arbitrary data type to JSON, discarding the error.
func ToJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func GetServerIPAddress(iface string) string {
	var addrs []net.Addr
	if iface == "" {
		var err error
		addrs, err = net.InterfaceAddrs()
		if err != nil {
			return ""
		}
	} else {
		interfaces, err := net.Interfaces()
		if err != nil {
			return ""
		}
		for _, i := range interfaces {
			if i.Name == iface {
				addrs, err = i.Addrs()
				if err != nil {
					return ""
				}
				break
			}
		}
	}

	for _, addr := range addrs {

		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && !ip.IP.IsLinkLocalUnicast() && !ip.IP.IsLinkLocalMulticast() {
			if ip.IP.To4() != nil {
				return ip.IP.String()
			}
		}
	}

	return ""
}

func isLower(s string) bool {
	return strings.ToLower(s) == s
}

func IsValidEmail(email string) bool {
	if !isLower(email) {
		return false
	}

	if addr, err := mail.ParseAddress(email); err != nil {
		return false
	} else if addr.Name != "" {
		// mail.ParseAddress accepts input of the form "Billy Bob <billy@example.com>" which we don't allow
		return false
	}

	return true
}

var reservedName = []string{
	"admin",
	"api",
	"channel",
	"claim",
	"error",
	"files",
	"help",
	"landing",
	"login",
	"mfa",
	"oauth",
	"plug",
	"plugins",
	"post",
	"signup",
	"boards",
	"playbooks",
}

func IsValidChannelIdentifier(s string) bool {
	return validSimpleAlphaNum.MatchString(s) && len(s) >= ChannelNameMinLength
}

var (
	validAlphaNum                           = regexp.MustCompile(`^[a-z0-9]+([a-z\-0-9]+|(__)?)[a-z0-9]+$`)
	validAlphaNumHyphenUnderscore           = regexp.MustCompile(`^[a-z0-9]+([a-z\-\_0-9]+|(__)?)[a-z0-9]+$`)
	validSimpleAlphaNum                     = regexp.MustCompile(`^[a-z0-9]+([a-z\-\_0-9]+|(__)?)[a-z0-9]*$`)
	validSimpleAlphaNumHyphenUnderscore     = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	validSimpleAlphaNumHyphenUnderscorePlus = regexp.MustCompile(`^[a-zA-Z0-9+_-]+$`)
)

func isValidAlphaNum(s string) bool {
	return validAlphaNum.MatchString(s)
}

func IsValidAlphaNumHyphenUnderscore(s string, withFormat bool) bool {
	if withFormat {
		return validAlphaNumHyphenUnderscore.MatchString(s)
	}
	return validSimpleAlphaNumHyphenUnderscore.MatchString(s)
}

func IsValidAlphaNumHyphenUnderscorePlus(s string) bool {
	return validSimpleAlphaNumHyphenUnderscorePlus.MatchString(s)
}

func Etag(parts ...any) string {

	etag := CurrentVersion

	for _, part := range parts {
		etag += fmt.Sprintf(".%v", part)
	}

	return etag
}

var (
	validHashtag = regexp.MustCompile(`^(#\pL[\pL\d\-_.]*[\pL\d])$`)
	puncStart    = regexp.MustCompile(`^[^\pL\d\s#]+`)
	hashtagStart = regexp.MustCompile(`^#{2,}`)
	puncEnd      = regexp.MustCompile(`[^\pL\d\s]+$`)
)

func ParseHashtags(text string) (string, string) {
	words := strings.Fields(text)

	hashtagString := ""
	plainString := ""
	for _, word := range words {
		// trim off surrounding punctuation
		word = puncStart.ReplaceAllString(word, "")
		word = puncEnd.ReplaceAllString(word, "")

		// and remove extra pound #s
		word = hashtagStart.ReplaceAllString(word, "#")

		if validHashtag.MatchString(word) {
			hashtagString += " " + word
		} else {
			plainString += " " + word
		}
	}

	if len(hashtagString) > 1000 {
		hashtagString = hashtagString[:999]
		lastSpace := strings.LastIndex(hashtagString, " ")
		if lastSpace > -1 {
			hashtagString = hashtagString[:lastSpace]
		} else {
			hashtagString = ""
		}
	}

	return strings.TrimSpace(hashtagString), strings.TrimSpace(plainString)
}

func ClearMentionTags(post string) string {
	post = strings.Replace(post, "<mention>", "", -1)
	post = strings.Replace(post, "</mention>", "", -1)
	return post
}

func IsValidHTTPURL(rawURL string) bool {
	if strings.Index(rawURL, "http://") != 0 && strings.Index(rawURL, "https://") != 0 {
		return false
	}

	if u, err := url.ParseRequestURI(rawURL); err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func IsValidId(value string) bool {
	if len(value) != 26 {
		return false
	}

	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

// RemoveDuplicateStrings does an in-place removal of duplicate strings
// from the input slice. The original slice gets modified.
func RemoveDuplicateStrings(in []string) []string {
	// In-place de-dup.
	// Copied from https://github.com/golang/go/wiki/SliceTricks#in-place-deduplicate-comparable
	if len(in) == 0 {
		return in
	}
	sort.Strings(in)
	j := 0
	for i := 1; i < len(in); i++ {
		if in[j] == in[i] {
			continue
		}
		j++
		in[j] = in[i]
	}
	return in[:j+1]
}

func GetPreferredTimezone(timezone StringMap) string {
	if timezone["useAutomaticTimezone"] == "true" {
		return timezone["automaticTimezone"]
	}

	return timezone["manualTimezone"]
}

// SanitizeUnicode will remove undesirable Unicode characters from a string.
func SanitizeUnicode(s string) string {
	return strings.Map(filterBlocklist, s)
}

// filterBlocklist returns `r` if it is not in the blocklist, otherwise drop (-1).
// Blocklist is taken from https://www.w3.org/TR/unicode-xml/#Charlist
func filterBlocklist(r rune) rune {
	const drop = -1
	switch r {
	case '\u0340', '\u0341': // clones of grave and acute; deprecated in Unicode
		return drop
	case '\u17A3', '\u17D3': // obsolete characters for Khmer; deprecated in Unicode
		return drop
	case '\u2028', '\u2029': // line and paragraph separator
		return drop
	case '\u202A', '\u202B', '\u202C', '\u202D', '\u202E': // BIDI embedding controls
		return drop
	case '\u206A', '\u206B': // activate/inhibit symmetric swapping; deprecated in Unicode
		return drop
	case '\u206C', '\u206D': // activate/inhibit Arabic form shaping; deprecated in Unicode
		return drop
	case '\u206E', '\u206F': // activate/inhibit national digit shapes; deprecated in Unicode
		return drop
	case '\uFFF9', '\uFFFA', '\uFFFB': // interlinear annotation characters
		return drop
	case '\uFEFF': // byte order mark
		return drop
	case '\uFFFC': // object replacement character
		return drop
	}

	// Scoping for musical notation
	if r >= 0x0001D173 && r <= 0x0001D17A {
		return drop
	}

	// Language tag code points
	if r >= 0x000E0000 && r <= 0x000E007F {
		return drop
	}

	return r
}

func IsCloud() bool {
	return os.Getenv("MM_CLOUD_INSTALLATION_ID") != ""
}
